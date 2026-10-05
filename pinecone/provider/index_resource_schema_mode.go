// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

const (
	maxFieldNameBytes    = 64
	maxFullTextFields    = 100
	maxNgramLength       = 10
	maxDescriptionLength = 256
)

var _ resource.ResourceWithValidateConfig = &IndexResource{}

var lenientObjectAs = basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true}

func indexSchemaResourceAttribute() schema.Attribute {
	description := schema.StringAttribute{
		MarkdownDescription: "A description of the field, at most 256 bytes.",
		Optional:            true,
		Computed:            true,
		Validators:          []validator.String{stringvalidator.LengthAtMost(maxDescriptionLength)},
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
	computedBool := func(what string) schema.BoolAttribute {
		return schema.BoolAttribute{
			MarkdownDescription: what,
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	}

	return schema.SingleNestedAttribute{
		MarkdownDescription: "The index's schema: the typed fields its records contain. Use it with `deployment` instead of " +
			"`dimension`, `metric`, `vector_type`, `spec`, and `embed`. A schema of named fields creates a document index, used " +
			"with the documents API. A schema made only of the reserved fields `_values` (dense) and `_sparse_values` (sparse) " +
			"creates a vector index, used with the vectors API. Metadata fields don't need to be declared: they're indexed " +
			"automatically when you upsert data. The schema can't be changed after the index is created; changing it replaces the index.",
		Optional: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"fields": schema.MapNestedAttribute{
				MarkdownDescription: "The schema's fields, keyed by field name. Set exactly one of `dense_vector`, `sparse_vector`, or " +
					"`string` on each. Field names are at most 64 bytes and can't start with `$` or `_`, except for the reserved fields.",
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dense_vector": schema.SingleNestedAttribute{
							MarkdownDescription: "A dense vector field. An index can have at most one.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"dimension": schema.Int32Attribute{
									MarkdownDescription: "The number of dimensions in the field's vectors.",
									Required:            true,
									Validators:          []validator.Int32{int32validator.AtLeast(1)},
								},
								"metric": schema.StringAttribute{
									MarkdownDescription: "The distance metric used for similarity search: `cosine`, `dotproduct`, or `euclidean`.",
									Required:            true,
									Validators:          []validator.String{stringvalidator.OneOf("cosine", "dotproduct", "euclidean")},
								},
								"description": description,
							},
						},
						"sparse_vector": schema.SingleNestedAttribute{
							MarkdownDescription: "A sparse vector field, which takes no dimension or metric. An index can have at most one.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"description": description,
							},
						},
						"string": schema.SingleNestedAttribute{
							MarkdownDescription: "A string field indexed for full-text search. An index can have at most 100.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"full_text_search": schema.SingleNestedAttribute{
									MarkdownDescription: "How the field's text is analyzed. Set it to `{}` for the defaults.",
									Required:            true,
									Attributes: map[string]schema.Attribute{
										"language": schema.StringAttribute{
											MarkdownDescription: "The language for text analysis. Defaults to `en`.",
											Optional:            true,
											Computed:            true,
											PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
										},
										"stemming": computedBool("Whether words are reduced to their root form, so \"moths\" matches \"moth\". Defaults to `false`."),
										"stop_words": computedBool("Whether common words such as \"the\" are filtered out. Requires `stemming`. " +
											"Defaults to `false`."),
										"ngram": schema.SingleNestedAttribute{
											MarkdownDescription: "Splits the field into character n-grams for substring or prefix matching. " +
												"Can't be combined with `stemming` or `stop_words`.",
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"min_gram": schema.Int64Attribute{
													MarkdownDescription: "The minimum n-gram length.",
													Required:            true,
													Validators:          []validator.Int64{int64validator.AtLeast(1)},
												},
												"max_gram": schema.Int64Attribute{
													MarkdownDescription: "The maximum n-gram length, at most 10.",
													Required:            true,
													Validators:          []validator.Int64{int64validator.Between(1, maxNgramLength)},
												},
												"prefix_only": computedBool("Whether only n-grams anchored at the start of each token are generated, " +
													"as for autocomplete. Defaults to `false`."),
											},
										},
									},
								},
								"description": description,
							},
						},
					},
				},
			},
		},
	}
}

func indexDeploymentResourceAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Where the index runs. Required with `schema`. Set exactly one of `managed` or `byoc`. " +
			"Changing it replaces the index.",
		Optional: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"managed": schema.SingleNestedAttribute{
				MarkdownDescription: "A serverless index.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"cloud": schema.StringAttribute{
						MarkdownDescription: "The public cloud where the index is hosted: `aws`, `gcp`, or `azure`.",
						Required:            true,
						Validators:          []validator.String{stringvalidator.OneOf("aws", "gcp", "azure")},
					},
					"region": schema.StringAttribute{
						MarkdownDescription: "The region where the index is hosted.",
						Required:            true,
					},
					"environment": schema.StringAttribute{
						MarkdownDescription: "The Pinecone environment hosting the index.",
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
			},
			"byoc": schema.SingleNestedAttribute{
				MarkdownDescription: "A BYOC (Bring Your Own Cloud) index. Only vector indexes, whose schema is made of the reserved fields, can run on BYOC.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"environment": schema.StringAttribute{
						MarkdownDescription: "The BYOC environment where the index is hosted.",
						Required:            true,
					},
				},
			},
		},
	}
}

// specMetricDefault plans the metric a new spec-based index gets when metric isn't configured:
// "dotproduct" for a sparse index and "cosine" otherwise. With embed, or with a vector_type that isn't
// known yet, it leaves the plan alone, so the metric is computed on create. On an existing index it
// also leaves the plan alone, so UseStateForUnknown keeps the index's metric: a metric can't change
// after creation, and a default that differs from it would plan a replacement. With schema it plans
// null: the metric belongs to the schema's dense vector field, and a static default would contradict
// the null state.
type specMetricDefault struct{}

func (specMetricDefault) Description(_ context.Context) string {
	return `Defaults to "dotproduct" for new sparse indexes, the model's metric with embed, and "cosine" otherwise, unless schema is set.`
}

func (m specMetricDefault) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (specMetricDefault) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}
	var schemaConfig, embed types.Object
	var vectorType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("schema"), &schemaConfig)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("embed"), &embed)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("vector_type"), &vectorType)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !schemaConfig.IsNull() {
		resp.PlanValue = types.StringNull()
		return
	}
	if !req.State.Raw.IsNull() || !embed.IsNull() || vectorType.IsUnknown() {
		return
	}
	if vectorType.ValueString() == "sparse" {
		resp.PlanValue = types.StringValue(string(pinecone.IndexMetricDotproduct))
	} else {
		resp.PlanValue = types.StringValue(string(pinecone.IndexMetricCosine))
	}
}

func (r *IndexResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config models.IndexResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateIndexStyleConfig(config)...)
	resp.Diagnostics.Append(validateIndexSchemaConfig(ctx, config)...)
}

// validateIndexStyleConfig checks that a configuration describes the index one way: with schema
// and deployment, or with dimension, metric, vector_type, spec, and embed.
func validateIndexStyleConfig(config models.IndexResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	schemaSet, deploymentSet := !config.Schema.IsNull(), !config.Deployment.IsNull()

	if !schemaSet && !deploymentSet {
		if !config.ReadCapacity.IsNull() {
			diags.AddAttributeError(path.Root("read_capacity"), "read_capacity requires schema",
				"Top-level read_capacity is used with schema and deployment. With spec, set spec.serverless.read_capacity or spec.byoc.read_capacity instead.")
		}
		if !config.CmekId.IsNull() {
			diags.AddAttributeError(path.Root("cmek_id"), "cmek_id requires schema",
				"cmek_id can only be set on an index created with schema and deployment.")
		}
		return diags
	}

	if !deploymentSet {
		diags.AddAttributeError(path.Root("deployment"), "Missing deployment",
			"An index with schema also needs deployment, which says where it runs, for example deployment = { managed = { cloud = \"aws\", region = \"us-east-1\" } }.")
	}
	if !schemaSet {
		diags.AddAttributeError(path.Root("schema"), "Missing schema", "deployment is used together with schema.")
	}

	for name, set := range map[string]bool{
		"dimension":   !config.Dimension.IsNull(),
		"metric":      !config.Metric.IsNull(),
		"vector_type": !config.VectorType.IsNull(),
		"spec":        !config.Spec.IsNull(),
		"embed":       !config.Embed.IsNull(),
	} {
		if set {
			diags.AddAttributeError(path.Root(name), "Conflicting index configuration",
				fmt.Sprintf("%s can't be used together with schema and deployment. An index is described either by schema and deployment, "+
					"or by dimension, metric, vector_type, spec, and embed. For a vector index described by schema, use the reserved "+
					"fields %s and %s.", name, models.ReservedDenseFieldName, models.ReservedSparseFieldName))
		}
	}
	return diags
}

// validateIndexSchemaConfig checks the schema and deployment against the rules the API enforces at
// creation, so they fail at plan time. Unknown values are skipped.
func validateIndexSchemaConfig(ctx context.Context, config models.IndexResourceModel) diag.Diagnostics {
	fields, diags := models.ResourceSchemaFields(ctx, config.Schema)
	if diags.HasError() || fields == nil {
		return diags
	}
	fieldsPath := path.Root("schema").AtName("fields")

	if len(fields) == 0 {
		diags.AddAttributeError(fieldsPath, "Empty schema", "A schema needs at least one field.")
	}

	var dense, sparse, fullText, reserved []string
	for _, name := range slices.Sorted(maps.Keys(fields)) {
		field, fieldPath := fields[name], fieldsPath.AtMapKey(name)

		kinds := 0
		for _, set := range []bool{field.DenseVector != nil, field.SparseVector != nil, field.String != nil} {
			if set {
				kinds++
			}
		}
		if kinds != 1 {
			diags.AddAttributeError(fieldPath, "Invalid schema field",
				fmt.Sprintf("Set exactly one of dense_vector, sparse_vector, or string on field %q.", name))
		}

		if len(name) > maxFieldNameBytes {
			diags.AddAttributeError(fieldPath, "Invalid field name", fmt.Sprintf("Field names can be at most %d bytes; %q is %d.", maxFieldNameBytes, name, len(name)))
		}
		if models.HasInvalidFieldNamePrefix(name) {
			diags.AddAttributeError(fieldPath, "Invalid field name",
				fmt.Sprintf("Field names can't start with $ or _ (%q). The only exceptions are the reserved vector fields %s and %s.",
					name, models.ReservedDenseFieldName, models.ReservedSparseFieldName))
		}

		switch {
		case name == models.ReservedDenseFieldName && field.DenseVector == nil:
			diags.AddAttributeError(fieldPath, "Invalid reserved field", fmt.Sprintf("%s must be a dense_vector field.", name))
		case name == models.ReservedSparseFieldName && field.SparseVector == nil:
			diags.AddAttributeError(fieldPath, "Invalid reserved field", fmt.Sprintf("%s must be a sparse_vector field.", name))
		}
		if models.IsReservedFieldName(name) {
			reserved = append(reserved, name)
		}

		switch {
		case field.DenseVector != nil:
			dense = append(dense, name)
		case field.SparseVector != nil:
			sparse = append(sparse, name)
		case field.String != nil:
			fullText = append(fullText, name)
			diags.Append(validateFullTextSearch(field.String.FullTextSearch, fieldPath.AtName("string").AtName("full_text_search"))...)
		}
	}

	if len(reserved) > 0 && len(reserved) != len(fields) {
		diags.AddAttributeError(fieldsPath, "Reserved fields mixed with named fields",
			fmt.Sprintf("The reserved fields %s and %s can only make up the entire schema, which creates a vector index. "+
				"A schema with named fields creates a document index; name its vector fields instead.", models.ReservedDenseFieldName, models.ReservedSparseFieldName))
	}
	if len(dense) > 1 {
		diags.AddAttributeError(fieldsPath, "Too many dense vector fields", fmt.Sprintf("An index can have at most one dense_vector field; found %v.", dense))
	}
	if len(sparse) > 1 {
		diags.AddAttributeError(fieldsPath, "Too many sparse vector fields", fmt.Sprintf("An index can have at most one sparse_vector field; found %v.", sparse))
	}
	if len(fullText) > maxFullTextFields {
		diags.AddAttributeError(fieldsPath, "Too many full-text search fields", fmt.Sprintf("An index can have at most %d string fields; found %d.", maxFullTextFields, len(fullText)))
	}

	diags.Append(validateIndexDeploymentConfig(ctx, config, models.IsDocumentIndexConfig(fields))...)
	return diags
}

func validateFullTextSearch(config *models.FullTextSearchModel, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if config == nil {
		return diags
	}
	stemming := !config.Stemming.IsUnknown() && config.Stemming.ValueBool()
	stopWords := !config.StopWords.IsUnknown() && config.StopWords.ValueBool()

	if stopWords && !config.Stemming.IsUnknown() && !stemming {
		diags.AddAttributeError(p.AtName("stop_words"), "stop_words requires stemming", "Set stemming = true to filter stop words.")
	}
	if config.Ngram != nil {
		if stemming || stopWords {
			diags.AddAttributeError(p.AtName("ngram"), "ngram can't be combined with stemming or stop_words",
				"N-gram tokenization replaces word-based analysis. Remove stemming and stop_words, or remove ngram.")
		}
		minGram, maxGram := config.Ngram.MinGram, config.Ngram.MaxGram
		if !minGram.IsUnknown() && !maxGram.IsUnknown() && minGram.ValueInt64() > maxGram.ValueInt64() {
			diags.AddAttributeError(p.AtName("ngram"), "Invalid n-gram range",
				fmt.Sprintf("min_gram (%d) can't be greater than max_gram (%d).", minGram.ValueInt64(), maxGram.ValueInt64()))
		}
	}
	return diags
}

func validateIndexDeploymentConfig(ctx context.Context, config models.IndexResourceModel, documentIndex bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if config.Deployment.IsNull() || config.Deployment.IsUnknown() {
		return diags
	}
	var deployment models.IndexResourceDeploymentModel
	diags.Append(config.Deployment.As(ctx, &deployment, lenientObjectAs)...)
	if diags.HasError() {
		return diags
	}

	switch {
	case deployment.Managed == nil && deployment.Byoc == nil:
		diags.AddAttributeError(path.Root("deployment"), "Missing deployment type", "Set one of deployment.managed or deployment.byoc.")
	case deployment.Managed != nil && deployment.Byoc != nil:
		diags.AddAttributeError(path.Root("deployment"), "Conflicting deployment types", "Set only one of deployment.managed or deployment.byoc.")
	case deployment.Byoc != nil && documentIndex:
		diags.AddAttributeError(path.Root("deployment").AtName("byoc"), "Document indexes require a managed deployment",
			"An index with named schema fields is a document index, which runs only on serverless (managed) deployments.")
	case deployment.Byoc != nil && !config.CmekId.IsNull():
		diags.AddAttributeError(path.Root("cmek_id"), "cmek_id isn't supported on BYOC", "Customer-managed encryption keys are only supported on managed deployments.")
	}
	return diags
}

// validateIndexStyleChange rejects switching an existing index between spec and schema with
// deployment. Both describe the same index, so replacing it would only lose its data.
func validateIndexStyleChange(config, state models.IndexResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	configSchema, stateSchema := !config.Schema.IsNull(), !state.Schema.IsNull()
	if configSchema == stateSchema {
		return diags
	}
	from, to := "spec", "schema and deployment"
	if stateSchema {
		from, to = to, from
	}
	diags.AddAttributeError(path.Root("schema"), "An index can't change how it's described",
		fmt.Sprintf("This index is managed with %s, and the configuration describes it with %s. An existing index can't switch between "+
			"the two. Describe it with %s again. %s", from, to, from, recreateIndexHint))
	return diags
}

// validateReadCapacityChange rejects moving a document index from dedicated read capacity back to
// on-demand, which the API doesn't support.
func validateReadCapacityChange(ctx context.Context, config, state models.IndexResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if state.Schema.IsNull() || state.ReadCapacity.IsNull() || config.ReadCapacity.IsNull() || config.ReadCapacity.IsUnknown() {
		return diags
	}
	fields, d := models.ResourceSchemaFields(ctx, state.Schema)
	diags.Append(d...)
	if diags.HasError() || !models.IsDocumentIndexConfig(fields) {
		return diags
	}

	var configured, current models.IndexReadCapacityResourceModel
	diags.Append(config.ReadCapacity.As(ctx, &configured, lenientObjectAs)...)
	diags.Append(state.ReadCapacity.As(ctx, &current, lenientObjectAs)...)
	if diags.HasError() {
		return diags
	}
	if !current.Dedicated.IsNull() && !configured.OnDemand.IsNull() {
		diags.AddAttributeError(path.Root("read_capacity").AtName("on_demand"), "Document indexes can't return to on-demand read capacity",
			"A document index with dedicated read capacity can't be switched back to on-demand. Keep read_capacity.dedicated. "+recreateIndexHint)
	}
	return diags
}

// indexReadCapacity returns the read capacity a model configures: top-level with schema, under spec
// otherwise.
func indexReadCapacity(ctx context.Context, model models.IndexResourceModel, diagnostics *diag.Diagnostics) types.Object {
	if !model.Schema.IsNull() {
		return model.ReadCapacity
	}
	return extractReadCapacityFromSpec(ctx, model.Spec, diagnostics)
}

func (r *IndexResource) createSchemaModeIndex(ctx context.Context, data models.IndexResourceModel, tags pinecone.IndexTags) diag.Diagnostics {
	indexSchema, diags := models.ToIndexSchema(ctx, data.Schema)
	if diags.HasError() {
		return diags
	}
	deployment, d := models.ToIndexDeployment(ctx, data.Deployment)
	diags.Append(d...)
	readCapacity, d := models.ToReadCapacityParams(ctx, data.ReadCapacity)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	deletionProtection := pinecone.DeletionProtection(data.DeletionProtection.ValueString())
	request := pinecone.CreateIndexRequest{
		Name:               data.Name.ValueString(),
		Schema:             indexSchema,
		Deployment:         deployment,
		ReadCapacity:       readCapacity,
		DeletionProtection: &deletionProtection,
	}
	if !data.CmekId.IsNull() && !data.CmekId.IsUnknown() {
		request.CmekId = data.CmekId.ValueStringPointer()
	}
	if tags != nil {
		request.Tags = &tags
	}

	if _, err := r.client.CreateIndex(ctx, &request); err != nil {
		diags.AddError("Failed to create index", err.Error())
	}
	return diags
}
