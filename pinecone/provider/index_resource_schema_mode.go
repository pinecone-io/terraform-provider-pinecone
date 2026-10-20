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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
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
	computedString := func(what string) schema.StringAttribute {
		return schema.StringAttribute{
			MarkdownDescription: what,
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	}
	parameters := func(what, effective string) schema.MapAttribute {
		return schema.MapAttribute{
			MarkdownDescription: "Model-specific parameters applied when embedding " + what + ", such as `truncate`. The defaults are chosen " +
				"per model and are rarely worth changing. The parameters the index uses, including the model's defaults, are reported in `" +
				effective + "`.",
			Optional:    true,
			ElementType: types.StringType,
		}
	}
	effectiveParameters := func(configured string) schema.MapAttribute {
		return schema.MapAttribute{
			MarkdownDescription: "The parameters the index uses: `" + configured + "` merged with the model's defaults.",
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers:       []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
		}
	}
	modelGuide := "Refer to the [model guide](https://docs.pinecone.io/guides/index-data/create-an-index#embedding-models) for available " +
		"models. Pinecone embeds the field's text when documents are upserted or updated, and embeds the query of "

	return schema.SingleNestedAttribute{
		MarkdownDescription: "The index's schema: the typed fields its records contain. Use it with `deployment` instead of " +
			"`dimension`, `metric`, `vector_type`, `spec`, and `embed`. A schema of named fields creates a document index, used " +
			"with the documents API. A schema made only of the reserved fields `_values` (dense) and `_sparse_values` (sparse) " +
			"creates a vector index, used with the vectors API. Metadata fields don't need to be declared: they're indexed " +
			"automatically when you upsert data. The schema can't be changed after the index is created; changing it replaces the index.",
		Optional: true,
		PlanModifiers: []planmodifier.Object{
			requiresSchemaReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"fields": schema.MapNestedAttribute{
				MarkdownDescription: "The schema's fields, keyed by field name. Set exactly one of `dense_vector`, `sparse_vector`, or " +
					"`string` on each. Field names are at most 64 bytes and can't start with `$` or `_`, except for the reserved fields `_values` and `_sparse_values`.",
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
							MarkdownDescription: "A string field searched by full-text search, integrated embedding, or both. Set at least one of " +
								"`full_text_search`, `embed`, or `sparse_embed`. They can be combined, so the same text can be searched by full-text " +
								"search, dense integrated embedding, and sparse integrated embedding. An index can have at most 100 string fields " +
								"that store their text.",
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"full_text_search": schema.SingleNestedAttribute{
									MarkdownDescription: "How the field's text is analyzed for full-text search. Set it to `{}` for the defaults. A field " +
										"that stores its text is always indexed for full-text search, so when this isn't set, it's enabled with the defaults. " +
										"Can't be set with `store_text = false`.",
									Optional:      true,
									Computed:      true,
									PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
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
								"embed": schema.SingleNestedAttribute{
									MarkdownDescription: "Dense integrated embedding with a model hosted by Pinecone. Set it to `{}` for the default model " +
										"at its default dimension and metric. An index has one dense vector slot: at most one field can set `embed` or be " +
										"a `dense_vector` field. " + modelGuide + "an `embed` search with the same model.",
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"model": computedString("The name of a hosted dense embedding model. Defaults to `llama-text-embed-v2`."),
										"dimension": schema.Int32Attribute{
											MarkdownDescription: "The dimension of the embedding vectors. Defaults to the model's default dimension. " +
												"A dimension the model doesn't support is rejected.",
											Optional:      true,
											Computed:      true,
											Validators:    []validator.Int32{int32validator.AtLeast(1)},
											PlanModifiers: []planmodifier.Int32{int32planmodifier.UseStateForUnknown()},
										},
										"metric": schema.StringAttribute{
											MarkdownDescription: "The distance metric used for similarity search: `cosine`, `dotproduct`, or `euclidean`. " +
												"Defaults to the model's default metric. A metric the model doesn't support is rejected.",
											Optional:      true,
											Computed:      true,
											Validators:    []validator.String{stringvalidator.OneOf("cosine", "dotproduct", "euclidean")},
											PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
										},
										"read_parameters":            parameters("search query text", "effective_read_parameters"),
										"write_parameters":           parameters("document text", "effective_write_parameters"),
										"effective_read_parameters":  effectiveParameters("read_parameters"),
										"effective_write_parameters": effectiveParameters("write_parameters"),
									},
								},
								"sparse_embed": schema.SingleNestedAttribute{
									MarkdownDescription: "Sparse integrated embedding with a model hosted by Pinecone. Set it to `{}` for the default " +
										"model. Sparse embeddings have no dimension or metric. An index has one sparse vector slot: at most one field can " +
										"set `sparse_embed` or be a `sparse_vector` field. " + modelGuide + "a `sparse_embed` search with the same model.",
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"model":                      computedString("The name of a hosted sparse embedding model. Defaults to `pinecone-sparse-english-v0`."),
										"read_parameters":            parameters("search query text", "effective_read_parameters"),
										"write_parameters":           parameters("document text", "effective_write_parameters"),
										"effective_read_parameters":  effectiveParameters("read_parameters"),
										"effective_write_parameters": effectiveParameters("write_parameters"),
									},
								},
								"store_text": computedBool("Whether the field's text is stored. Defaults to `true`, in which case the field is also indexed " +
									"for full-text search. Set it to `false` only on a field with `embed` or `sparse_embed` and without `full_text_search`: " +
									"the text is embedded and then discarded. This can save space when the field is only used for similarity search, or " +
									"keep sensitive text out of Pinecone. The field then has no full-text search: it can only be searched with `embed` " +
									"or `sparse_embed` scoring, can't be matched by `$match_*` filters, and can't be named in a fetch's or search's " +
									"`include_fields`. If you might need full-text search later, keep text stored."),
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
			requiresReplaceUnlessStyleSwitchObject(),
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
				MarkdownDescription: "A BYOC (Bring Your Own Cloud) index. Only vector indexes, whose schema is made of the reserved fields, can run on BYOC. " +
					"BYOC indexes need `read_capacity.dedicated`: they don't support on-demand read capacity.",
				Optional: true,
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

const requiresSchemaReplaceDescription = requiresReplaceUnlessStyleSwitchDescription + " Integrated embedding parameters that " +
	"aren't recorded as configured, as after an import, are taken into state without replacing the index when the index already uses them."

// requiresSchemaReplace recreates the index when its schema changes, unless the style switches or
// the change only adopts integrated embedding parameters the index already uses.
//
// When the plan changes anything, the framework marks every computed attribute that isn't
// configured as unknown before this runs, and the nested UseStateForUnknown modifiers that restore
// them run after it. The adoption check resolves those values first, so they don't count as
// changes.
func requiresSchemaReplace() planmodifier.Object {
	return objectplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
		switching, diags := indexStyleSwitch(ctx, req.Config, req.State)
		resp.Diagnostics.Append(diags...)
		if switching {
			return
		}
		planned, ok := priorForUnconfiguredUnknowns(ctx, req.PlanValue, req.ConfigValue, req.StateValue).(types.Object)
		resp.RequiresReplace = !ok || !models.AdoptsConfiguredParameters(ctx, planned, req.StateValue)
	}, requiresSchemaReplaceDescription, requiresSchemaReplaceDescription)
}

// priorForUnconfiguredUnknowns returns planned with each unknown value that isn't configured
// replaced by its prior value, as UseStateForUnknown will. An unknown that is configured, such as a
// value from another resource, stays unknown.
func priorForUnconfiguredUnknowns(ctx context.Context, planned, config, prior attr.Value) attr.Value {
	if planned == nil || prior == nil || prior.IsUnknown() {
		return planned
	}
	if planned.IsUnknown() {
		if config == nil || config.IsNull() {
			return prior
		}
		return planned
	}
	if planned.IsNull() || prior.IsNull() {
		return planned
	}
	switch p := planned.(type) {
	case types.Object:
		configAttrs, priorAttrs := objectAttributes(config), objectAttributes(prior)
		attrs := make(map[string]attr.Value, len(p.Attributes()))
		for name, value := range p.Attributes() {
			attrs[name] = priorForUnconfiguredUnknowns(ctx, value, configAttrs[name], priorAttrs[name])
		}
		if resolved, diags := types.ObjectValue(p.AttributeTypes(ctx), attrs); !diags.HasError() {
			return resolved
		}
	case types.Map:
		configElements, priorElements := mapElements(config), mapElements(prior)
		elements := make(map[string]attr.Value, len(p.Elements()))
		for key, value := range p.Elements() {
			elements[key] = priorForUnconfiguredUnknowns(ctx, value, configElements[key], priorElements[key])
		}
		if resolved, diags := types.MapValue(p.ElementType(ctx), elements); !diags.HasError() {
			return resolved
		}
	}
	return planned
}

func objectAttributes(v attr.Value) map[string]attr.Value {
	if obj, ok := v.(types.Object); ok {
		return obj.Attributes()
	}
	return nil
}

func mapElements(v attr.Value) map[string]attr.Value {
	if m, ok := v.(types.Map); ok {
		return m.Elements()
	}
	return nil
}

func adoptsSchemaParameters(ctx context.Context, planned, prior attr.Value) bool {
	plannedObj, ok := planned.(types.Object)
	if !ok {
		return false
	}
	priorObj, ok := prior.(types.Object)
	return ok && models.AdoptsConfiguredParameters(ctx, plannedObj, priorObj)
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
	resp.Diagnostics.Append(validateIndexTags(config.Tags)...)
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
		diags.Append(validateIndexSpecConfig(config.Spec)...)
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
	fields, pending, diags := configSchemaFields(ctx, config.Schema)
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
		if kinds != 1 && !pending[name] {
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
		case pending[name]:
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
			stringPath := fieldPath.AtName("string")
			if field.String.Embed != nil {
				dense = append(dense, name)
			}
			if field.String.SparseEmbed != nil {
				sparse = append(sparse, name)
			}
			if field.String.StoreText.IsNull() || field.String.StoreText.IsUnknown() || field.String.StoreText.ValueBool() {
				fullText = append(fullText, name)
			}
			if !pending[name] {
				diags.Append(validateStringField(name, *field.String, stringPath)...)
			}
			diags.Append(validateFullTextSearch(field.String.FullTextSearch, stringPath.AtName("full_text_search"))...)
		}
	}

	if len(reserved) > 0 && len(reserved) != len(fields) {
		diags.AddAttributeError(fieldsPath, "Reserved fields mixed with named fields",
			fmt.Sprintf("The reserved fields %s and %s can only make up the entire schema, which creates a vector index. "+
				"A schema with named fields creates a document index; name its vector fields instead.", models.ReservedDenseFieldName, models.ReservedSparseFieldName))
	}
	if len(dense) > 1 {
		diags.AddAttributeError(fieldsPath, "Too many dense vector fields",
			fmt.Sprintf("An index has one dense vector slot: at most one field can be a dense_vector field or set string.embed; found %v.", dense))
	}
	if len(sparse) > 1 {
		diags.AddAttributeError(fieldsPath, "Too many sparse vector fields",
			fmt.Sprintf("An index has one sparse vector slot: at most one field can be a sparse_vector field or set string.sparse_embed; found %v.", sparse))
	}
	if len(fullText) > maxFullTextFields {
		diags.AddAttributeError(fieldsPath, "Too many full-text search fields",
			fmt.Sprintf("An index can have at most %d string fields that store their text, which are indexed for full-text search; found %d.", maxFullTextFields, len(fullText)))
	}

	diags.Append(validateIndexDeploymentConfig(ctx, config, models.IsDocumentIndexConfig(fields))...)
	return diags
}

// configSchemaFields decodes the configured schema fields for validation. The map holds every
// field name. A field whose value or type isn't known yet decodes empty and is marked pending. A
// string field with an attribute that isn't known yet is marked pending too, since an unknown
// nested object decodes as nil, the same as one that isn't set.
func configSchemaFields(ctx context.Context, obj types.Object) (map[string]models.IndexResourceSchemaFieldModel, map[string]bool, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil, nil
	}
	fieldsMap, ok := obj.Attributes()["fields"].(types.Map)
	if !ok || fieldsMap.IsNull() || fieldsMap.IsUnknown() {
		return nil, nil, nil
	}

	var diags diag.Diagnostics
	fields := make(map[string]models.IndexResourceSchemaFieldModel, len(fieldsMap.Elements()))
	pending := make(map[string]bool)
	for name, value := range fieldsMap.Elements() {
		fieldObj, ok := value.(types.Object)
		if !ok || fieldObj.IsNull() || fieldObj.IsUnknown() {
			fields[name], pending[name] = models.IndexResourceSchemaFieldModel{}, true
			continue
		}
		var field models.IndexResourceSchemaFieldModel
		diags.Append(fieldObj.As(ctx, &field, lenientObjectAs)...)
		fields[name] = field
		for kindName, kind := range fieldObj.Attributes() {
			if kind.IsUnknown() {
				pending[name] = true
			}
			if kindObj, ok := kind.(types.Object); ok && kindName == "string" && !kindObj.IsNull() && !kindObj.IsUnknown() {
				for _, value := range kindObj.Attributes() {
					if value.IsUnknown() {
						pending[name] = true
					}
				}
			}
		}
	}
	return fields, pending, diags
}

// validateIntegratedEmbeddingKept rejects a configuration that leaves out integrated embedding a
// field of the index has, which would otherwise replace the index. A field can have it without
// its configuration ever setting it: version 5.0.0 couldn't describe integrated embedding, so it
// imported such fields without it, and a refresh now records it.
func validateIntegratedEmbeddingKept(ctx context.Context, config, state models.IndexResourceModel) diag.Diagnostics {
	stateFields, diags := models.ResourceSchemaFields(ctx, state.Schema)
	configFields, pending, d := configSchemaFields(ctx, config.Schema)
	diags.Append(d...)
	if diags.HasError() || stateFields == nil || configFields == nil {
		return diags
	}
	for _, name := range slices.Sorted(maps.Keys(stateFields)) {
		current, configured := stateFields[name].String, configFields[name].String
		if current == nil || configured == nil || pending[name] {
			continue
		}
		stringPath := path.Root("schema").AtName("fields").AtMapKey(name).AtName("string")
		for _, embedding := range []struct {
			attribute string
			missing   bool
		}{
			{"embed", current.Embed != nil && configured.Embed == nil},
			{"sparse_embed", current.SparseEmbed != nil && configured.SparseEmbed == nil},
		} {
			if embedding.missing {
				diags.AddAttributeError(stringPath.AtName(embedding.attribute), "Integrated embedding can't be removed from a field",
					fmt.Sprintf("Field %q of this index uses %s, which is fixed when the index is created. Add %s = {} to the field "+
						"to match the index: its model and other settings are read from the index. %s",
						name, embedding.attribute, embedding.attribute, recreateIndexHint))
			}
		}
	}
	return diags
}

// validateStringField checks how a string field's text is searched and stored.
func validateStringField(name string, field models.StringResourceFieldModel, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	embedded := field.Embed != nil || field.SparseEmbed != nil
	if field.FullTextSearch == nil && !embedded {
		diags.AddAttributeError(p, "Invalid string field",
			fmt.Sprintf("Set at least one of full_text_search, embed, or sparse_embed on string field %q. Metadata fields don't need to be "+
				"declared: they're indexed automatically when you upsert data.", name))
	}
	if field.StoreText.IsNull() || field.StoreText.ValueBool() {
		return diags
	}
	if !embedded {
		diags.AddAttributeError(p.AtName("store_text"), "store_text = false requires integrated embedding",
			"A field that doesn't store its text has to be embedded. Set embed or sparse_embed, or remove store_text.")
	}
	if field.FullTextSearch != nil {
		diags.AddAttributeError(p.AtName("store_text"), "store_text = false can't be combined with full_text_search",
			"Full-text search needs the field's text stored. Remove full_text_search, or remove store_text.")
	}
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

func validateIndexSpecConfig(spec types.Object) diag.Diagnostics {
	var diags diag.Diagnostics
	if spec.IsUnknown() {
		return diags
	}
	if spec.IsNull() {
		diags.AddAttributeError(path.Root("spec"), "Missing index configuration",
			"Describe the index with spec, for example spec = { serverless = { cloud = \"aws\", region = \"us-east-1\" } }, or with schema and deployment.")
		return diags
	}
	kinds := 0
	for _, kind := range spec.Attributes() {
		if kind.IsUnknown() {
			return diags
		}
		if !kind.IsNull() {
			kinds++
		}
	}
	switch {
	case kinds == 0:
		diags.AddAttributeError(path.Root("spec"), "Missing spec type", "Set one of spec.serverless, spec.byoc, or spec.pod.")
	case kinds > 1:
		diags.AddAttributeError(path.Root("spec"), "Conflicting spec types", "Set only one of spec.serverless, spec.byoc, or spec.pod.")
	}
	return diags
}

func validateIndexDeploymentConfig(ctx context.Context, config models.IndexResourceModel, documentIndex bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if config.Deployment.IsNull() || config.Deployment.IsUnknown() {
		return diags
	}
	for _, kind := range config.Deployment.Attributes() {
		if kind.IsUnknown() {
			return diags
		}
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
