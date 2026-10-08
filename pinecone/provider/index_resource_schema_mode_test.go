// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

type schemaFields = map[string]models.IndexResourceSchemaFieldModel

func pointerTo[T any](v T) *T { return &v }

func denseField(dimension int32, metric string) models.IndexResourceSchemaFieldModel {
	return models.IndexResourceSchemaFieldModel{DenseVector: &models.DenseVectorFieldModel{
		Dimension: types.Int32Value(dimension), Metric: types.StringValue(metric),
	}}
}

func sparseField() models.IndexResourceSchemaFieldModel {
	return models.IndexResourceSchemaFieldModel{SparseVector: &models.SparseVectorFieldModel{}}
}

func fullTextField(config models.FullTextSearchModel) models.IndexResourceSchemaFieldModel {
	return models.IndexResourceSchemaFieldModel{String: &models.StringResourceFieldModel{FullTextSearch: &config}}
}

func managedDeployment() models.IndexResourceDeploymentModel {
	return models.IndexResourceDeploymentModel{Managed: &models.ManagedDeploymentModel{
		Cloud: types.StringValue("aws"), Region: types.StringValue("us-east-1"),
	}}
}

func byocDeployment() models.IndexResourceDeploymentModel {
	return models.IndexResourceDeploymentModel{Byoc: &models.ByocDeploymentModel{Environment: types.StringValue("aws-us-east-1-b921")}}
}

func withSchemaMode(t *testing.T, s schema.Schema, model models.IndexResourceModel, fields schemaFields, deployment models.IndexResourceDeploymentModel) models.IndexResourceModel {
	t.Helper()
	ctx := context.Background()
	fieldsMap, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: models.IndexResourceSchemaFieldModel{}.AttrTypes()}, fields)
	if d.HasError() {
		t.Fatalf("building fields: %v", d)
	}
	model.Schema, d = types.ObjectValueFrom(ctx, attrTypesOf(t, s, "schema"), models.IndexResourceSchemaModel{Fields: fieldsMap})
	if d.HasError() {
		t.Fatalf("building schema: %v", d)
	}
	model.Deployment, d = types.ObjectValueFrom(ctx, attrTypesOf(t, s, "deployment"), deployment)
	if d.HasError() {
		t.Fatalf("building deployment: %v", d)
	}
	model.Spec = types.ObjectNull(attrTypesOf(t, s, "spec"))
	model.Dimension, model.Metric, model.VectorType = types.Int32Null(), types.StringNull(), types.StringNull()
	return model
}

func withReadCapacity(t *testing.T, s schema.Schema, model models.IndexResourceModel, dedicated bool) models.IndexResourceModel {
	t.Helper()
	rc := models.IndexReadCapacityResourceModel{
		Dedicated: types.ObjectNull(models.IndexReadCapacityDedicatedResourceModel{}.AttrTypes()),
		OnDemand:  types.ObjectNull(models.IndexReadCapacityOnDemandResourceModel{}.AttrTypes()),
	}
	if dedicated {
		rc.Dedicated = types.ObjectValueMust(models.IndexReadCapacityDedicatedResourceModel{}.AttrTypes(), map[string]attr.Value{
			"node_type": types.StringValue("b1"), "replicas": types.Int32Value(1), "shards": types.Int32Value(1),
		})
	} else {
		rc.OnDemand = types.ObjectValueMust(models.IndexReadCapacityOnDemandResourceModel{}.AttrTypes(), map[string]attr.Value{})
	}
	obj, d := types.ObjectValueFrom(context.Background(), attrTypesOf(t, s, "read_capacity"), rc)
	if d.HasError() {
		t.Fatalf("building read_capacity: %v", d)
	}
	model.ReadCapacity = obj
	return model
}

// withUnknownField replaces one schema field with value, which is unknown or has an unknown type.
func withUnknownField(t *testing.T, s schema.Schema, model models.IndexResourceModel, name string, value types.Object) models.IndexResourceModel {
	t.Helper()
	schemaAttrs := model.Schema.Attributes()
	fieldsMap, ok := schemaAttrs["fields"].(types.Map)
	if !ok {
		t.Fatalf("schema.fields is %T, want types.Map", schemaAttrs["fields"])
	}
	fields := fieldsMap.Elements()
	fields[name] = value
	schemaAttrs["fields"] = types.MapValueMust(types.ObjectType{AttrTypes: models.IndexResourceSchemaFieldModel{}.AttrTypes()}, fields)
	model.Schema = types.ObjectValueMust(attrTypesOf(t, s, "schema"), schemaAttrs)
	return model
}

// withUnknownDeployment makes the named deployment type unknown, as when it's set from another
// resource's attribute.
func withUnknownDeployment(t *testing.T, s schema.Schema, model models.IndexResourceModel, name string) models.IndexResourceModel {
	t.Helper()
	attrs := model.Deployment.Attributes()
	kindType, ok := attrs[name].Type(context.Background()).(types.ObjectType)
	if !ok {
		t.Fatalf("deployment.%s is %T, want types.ObjectType", name, attrs[name].Type(context.Background()))
	}
	attrs[name] = types.ObjectUnknown(kindType.AttrTypes)
	model.Deployment = types.ObjectValueMust(attrTypesOf(t, s, "deployment"), attrs)
	return model
}

func runValidateConfig(t *testing.T, s schema.Schema, config models.IndexResourceModel) []string {
	t.Helper()
	ctx := context.Background()
	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	if d := plan.Set(ctx, &config); d.HasError() {
		t.Fatalf("setting config: %v", d)
	}
	var resp resource.ValidateConfigResponse
	r := &IndexResource{PineconeResource: &PineconeResource{}}
	r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s, Raw: plan.Raw}}, &resp)
	var summaries []string
	for _, d := range resp.Diagnostics.Errors() {
		summaries = append(summaries, d.Summary())
	}
	return summaries
}

func TestIndexResourceValidateConfig(t *testing.T) {
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	document := func(fields schemaFields) models.IndexResourceModel {
		return withSchemaMode(t, s, base, fields, managedDeployment())
	}
	documentFields := schemaFields{"embedding": denseField(1536, "dotproduct"), "terms": sparseField(), "body": fullTextField(models.FullTextSearchModel{})}

	withDimension := document(documentFields)
	withDimension.Dimension = types.Int32Value(1536)
	noDeployment := document(documentFields)
	noDeployment.Deployment = types.ObjectNull(attrTypesOf(t, s, "deployment"))
	noSchema := document(documentFields)
	noSchema.Schema = types.ObjectNull(attrTypesOf(t, s, "schema"))
	specReadCapacity := withReadCapacity(t, s, base, false)
	specCmek := base
	specCmek.CmekId = types.StringValue("cmek-123")
	byocCmek := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, byocDeployment())
	byocCmek.CmekId = types.StringValue("cmek-123")
	bothDeployments := managedDeployment()
	bothDeployments.Byoc = byocDeployment().Byoc
	fieldType := types.ObjectType{AttrTypes: models.IndexResourceSchemaFieldModel{}.AttrTypes()}
	unknownField := withUnknownField(t, s, document(documentFields), "embedding", types.ObjectUnknown(fieldType.AttrTypes))
	unknownDense := withUnknownField(t, s, document(documentFields), "embedding", types.ObjectValueMust(fieldType.AttrTypes, map[string]attr.Value{
		"dense_vector":  types.ObjectUnknown(models.DenseVectorFieldModel{}.AttrTypes()),
		"sparse_vector": types.ObjectNull(models.SparseVectorFieldModel{}.AttrTypes()),
		"string":        types.ObjectNull(models.StringResourceFieldModel{}.AttrTypes()),
	}))
	unknownManaged := withUnknownDeployment(t, s, document(documentFields), "managed")
	unknownByoc := withUnknownDeployment(t, s, withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, byocDeployment()), "byoc")

	tests := []struct {
		name    string
		config  models.IndexResourceModel
		wantErr string
	}{
		{name: "spec-based index", config: base},
		{name: "document index", config: document(documentFields)},
		{name: "reserved vector index on BYOC", config: withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "dotproduct"), "_sparse_values": sparseField()}, byocDeployment())},
		{name: "sparse-only reserved index", config: document(schemaFields{"_sparse_values": sparseField()})},
		{name: "schema with dimension", config: withDimension, wantErr: "Conflicting index configuration"},
		{name: "schema without deployment", config: noDeployment, wantErr: "Missing deployment"},
		{name: "deployment without schema", config: noSchema, wantErr: "Missing schema"},
		{name: "top-level read_capacity with spec", config: specReadCapacity, wantErr: "read_capacity requires schema"},
		{name: "cmek_id with spec", config: specCmek, wantErr: "cmek_id requires schema"},
		{name: "cmek_id on BYOC", config: byocCmek, wantErr: "cmek_id isn't supported on BYOC"},
		{name: "empty schema", config: document(schemaFields{}), wantErr: "Empty schema"},
		{name: "reserved mixed with named", config: document(schemaFields{"_values": denseField(1024, "cosine"), "body": fullTextField(models.FullTextSearchModel{})}), wantErr: "Reserved fields mixed with named fields"},
		{name: "_values as sparse", config: document(schemaFields{"_values": sparseField()}), wantErr: "Invalid reserved field"},
		{name: "underscore prefix", config: document(schemaFields{"_private": sparseField()}), wantErr: "Invalid field name"},
		{name: "dollar prefix", config: document(schemaFields{"$price": sparseField()}), wantErr: "Invalid field name"},
		{name: "long field name", config: document(schemaFields{strings.Repeat("f", 65): sparseField()}), wantErr: "Invalid field name"},
		{name: "two dense fields", config: document(schemaFields{"a": denseField(8, "cosine"), "b": denseField(8, "cosine")}), wantErr: "Too many dense vector fields"},
		{name: "two field types", config: document(schemaFields{"a": {DenseVector: denseField(8, "cosine").DenseVector, SparseVector: &models.SparseVectorFieldModel{}}}), wantErr: "Invalid schema field"},
		{name: "stop words without stemming", config: document(schemaFields{"body": fullTextField(models.FullTextSearchModel{StopWords: types.BoolValue(true)})}), wantErr: "stop_words requires stemming"},
		{name: "ngram with stemming", config: document(schemaFields{"body": fullTextField(models.FullTextSearchModel{
			Stemming: types.BoolValue(true), Ngram: &models.NgramModel{MinGram: types.Int64Value(2), MaxGram: types.Int64Value(3)},
		})}), wantErr: "ngram can't be combined with stemming or stop_words"},
		{name: "ngram range", config: document(schemaFields{"body": fullTextField(models.FullTextSearchModel{
			Ngram: &models.NgramModel{MinGram: types.Int64Value(5), MaxGram: types.Int64Value(3)},
		})}), wantErr: "Invalid n-gram range"},
		{name: "document index on BYOC", config: withSchemaMode(t, s, base, documentFields, byocDeployment()), wantErr: "Document indexes require a managed deployment"},
		{name: "both deployment types", config: withSchemaMode(t, s, base, documentFields, bothDeployments), wantErr: "Conflicting deployment types"},
		{name: "no deployment type", config: withSchemaMode(t, s, base, documentFields, models.IndexResourceDeploymentModel{}), wantErr: "Missing deployment type"},
		{name: "unknown field", config: unknownField},
		{name: "unknown field type", config: unknownDense},
		{name: "unknown managed deployment", config: unknownManaged},
		{name: "unknown byoc deployment", config: unknownByoc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := runValidateConfig(t, s, tt.config)
			if tt.wantErr == "" {
				if len(errs) > 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
				return
			}
			if !slices.Contains(errs, tt.wantErr) {
				t.Fatalf("errors = %v, want %q", errs, tt.wantErr)
			}
		})
	}
}

func TestIndexResourceModifyPlan_schemaMode(t *testing.T) {
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	document := withSchemaMode(t, s, base, schemaFields{"embedding": denseField(1536, "dotproduct"), "body": fullTextField(models.FullTextSearchModel{})}, managedDeployment())
	vector := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, managedDeployment())
	retagged := document
	retagged.Tags = types.MapValueMust(types.StringType, map[string]attr.Value{"team": types.StringValue("search")})
	renamedVector := vector
	renamedVector.Name = types.StringValue("my-renamed-index")

	tests := []struct {
		name    string
		config  models.IndexResourceModel
		state   *models.IndexResourceModel
		wantErr string
	}{
		{name: "create document index", config: document},
		{name: "update document index tags", config: retagged, state: &document},
		{name: "spec to schema", config: vector, state: &base, wantErr: "An index can't change how it's described"},
		{name: "schema to spec", config: base, state: &vector, wantErr: "An index can't change how it's described"},
		{name: "spec to schema while renaming", config: renamedVector, state: &base, wantErr: "An index can't change how it's described"},
		{name: "document index dedicated to on-demand", config: withReadCapacity(t, s, document, false), state: pointerTo(withReadCapacity(t, s, document, true)),
			wantErr: "Document indexes can't return to on-demand read capacity"},
		{name: "document index on-demand to dedicated", config: withReadCapacity(t, s, document, true), state: pointerTo(withReadCapacity(t, s, document, false))},
		{name: "vector index dedicated to on-demand", config: withReadCapacity(t, s, vector, false), state: pointerTo(withReadCapacity(t, s, vector, true))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := runModifyPlan(t, s, tt.config, tt.state)
			if tt.wantErr == "" {
				if len(errs) > 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
				return
			}
			if len(errs) != 1 || errs[0] != tt.wantErr {
				t.Fatalf("errors = %v, want [%q]", errs, tt.wantErr)
			}
		})
	}
}

func TestPlansIndexReplacement_schemaMode(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	document := withSchemaMode(t, s, base, schemaFields{"body": fullTextField(models.FullTextSearchModel{})}, managedDeployment())
	otherRegion := managedDeployment()
	otherRegion.Managed.Region = types.StringValue("us-west-2")

	tests := []struct {
		name   string
		config models.IndexResourceModel
		want   bool
	}{
		{"unchanged", document, false},
		{"field added", withSchemaMode(t, s, base, schemaFields{"body": fullTextField(models.FullTextSearchModel{}), "title": fullTextField(models.FullTextSearchModel{})}, managedDeployment()), true},
		{"region changed", withSchemaMode(t, s, base, schemaFields{"body": fullTextField(models.FullTextSearchModel{})}, otherRegion), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaType := s.Type().TerraformType(ctx)
			plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(schemaType, nil)}
			state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(schemaType, nil)}
			if d := plan.Set(ctx, &tt.config); d.HasError() {
				t.Fatalf("setting plan: %v", d)
			}
			if d := state.Set(ctx, &document); d.HasError() {
				t.Fatalf("setting state: %v", d)
			}
			got, d := plansIndexReplacement(ctx, resource.ModifyPlanRequest{Plan: plan, State: state})
			if d.HasError() || got != tt.want {
				t.Errorf("plansIndexReplacement() = %v (%v), want %v", got, d, tt.want)
			}
		})
	}
}

func TestSpecMetricDefault(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	base.Metric = types.StringNull()
	document := withSchemaMode(t, s, base, schemaFields{"body": fullTextField(models.FullTextSearchModel{})}, managedDeployment())
	configured := base
	configured.Metric = types.StringValue("dotproduct")
	sparse := base
	sparse.Dimension = types.Int32Null()
	sparse.VectorType = types.StringValue("sparse")
	unknownVectorType := base
	unknownVectorType.VectorType = types.StringUnknown()
	integrated := withEmbed(t, s, base, "pinecone-sparse-english-v0", "chunk_text")
	// An existing integrated index whose model's metric isn't the default.
	existingIntegrated := integrated
	existingIntegrated.Metric = types.StringValue("dotproduct")

	tests := []struct {
		name   string
		config models.IndexResourceModel
		state  *models.IndexResourceModel
		want   types.String
	}{
		{name: "spec-based index", config: base, want: types.StringValue("cosine")},
		{name: "sparse index", config: sparse, want: types.StringValue("dotproduct")},
		{name: "unknown vector_type", config: unknownVectorType, want: types.StringUnknown()},
		{name: "embed uses the model's metric", config: integrated, want: types.StringUnknown()},
		{name: "schema-mode index", config: document, want: types.StringNull()},
		{name: "configured metric", config: configured, want: types.StringValue("dotproduct")},
		// Left unknown, so UseStateForUnknown keeps "dotproduct" instead of planning a replacement.
		{name: "existing index keeps its metric", config: base, state: &existingIntegrated, want: types.StringUnknown()},
		{name: "existing schema-mode index", config: document, state: &document, want: types.StringNull()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
			if d := plan.Set(ctx, &tt.config); d.HasError() {
				t.Fatalf("setting config: %v", d)
			}
			planValue := types.StringUnknown()
			if !tt.config.Metric.IsNull() {
				planValue = tt.config.Metric
			}
			req := planmodifier.StringRequest{Config: tfsdk.Config{Schema: s, Raw: plan.Raw}, ConfigValue: tt.config.Metric, PlanValue: planValue}
			if tt.state != nil {
				req.State = tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
				if d := req.State.Set(ctx, tt.state); d.HasError() {
					t.Fatalf("setting state: %v", d)
				}
			}
			resp := planmodifier.StringResponse{PlanValue: planValue}
			specMetricDefault{}.PlanModifyString(ctx, req, &resp)
			if resp.Diagnostics.HasError() || !resp.PlanValue.Equal(tt.want) {
				t.Errorf("planned metric = %v (%v), want %v", resp.PlanValue, resp.Diagnostics, tt.want)
			}
		})
	}
}

func TestIndexResource_schemaModeStateMatchesSchema(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	language := "en"
	index := &pinecone.Index{
		Name: "articles",
		Host: "https://articles.example.com",
		Schema: &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
			"embedding": {DenseVector: &pinecone.DenseVectorField{Dimension: 1536, Metric: pinecone.IndexMetricDotproduct}},
			"terms":     {SparseVector: &pinecone.SparseVectorField{}},
			"body":      {String: &pinecone.StringField{FullTextSearch: &pinecone.FullTextSearchConfig{Language: &language}}},
		}},
		Deployment:   &pinecone.IndexDeployment{Managed: &pinecone.ManagedDeployment{Cloud: pinecone.CloudAWS, Region: "us-east-1"}},
		ReadCapacity: &pinecone.ReadCapacity{OnDemand: &pinecone.ReadCapacityOnDemand{}},
		Status:       &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady},
	}

	model := testIndexModel(t, s)
	model.Spec = types.ObjectNull(attrTypesOf(t, s, "spec"))
	if d := model.Read(ctx, index); d.HasError() {
		t.Fatalf("Read returned errors: %v", d)
	}
	state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatalf("schema-mode state doesn't fit the resource schema: %v", d)
	}
	var region types.String
	if d := state.GetAttribute(ctx, path.Root("deployment").AtName("managed").AtName("region"), &region); d.HasError() || region.ValueString() != "us-east-1" {
		t.Errorf("deployment.managed.region = %v (%v)", region, d)
	}
}
