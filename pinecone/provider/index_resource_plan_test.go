// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

func indexResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	NewIndexResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("building schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func attrTypesOf(t *testing.T, s schema.Schema, name string) map[string]attr.Type {
	t.Helper()
	objectType, ok := s.Attributes[name].GetType().(types.ObjectType)
	if !ok {
		t.Fatalf("attribute %q is not an object", name)
	}
	return objectType.AttrTypes
}

func testIndexModel(t *testing.T, s schema.Schema) models.IndexResourceModel {
	t.Helper()
	ctx := context.Background()
	spec, d := types.ObjectValueFrom(ctx, attrTypesOf(t, s, "spec"), models.IndexSpecModel{
		Serverless: &models.IndexServerlessSpecModel{
			Cloud:        types.StringValue("aws"),
			Region:       types.StringValue("us-west-2"),
			ReadCapacity: types.ObjectNull(models.IndexReadCapacityResourceModel{}.AttrTypes()),
			Schema:       types.ObjectNull(models.IndexMetadataSchemaModel{}.AttrTypes()),
		},
	})
	if d.HasError() {
		t.Fatalf("building spec: %v", d)
	}
	timeoutsType, ok := s.Blocks["timeouts"].Type().(timeouts.Type)
	if !ok {
		t.Fatal("timeouts block has an unexpected type")
	}
	return models.IndexResourceModel{
		Id:                 types.StringValue("my-index"),
		Name:               types.StringValue("my-index"),
		Dimension:          types.Int32Value(1024),
		Metric:             types.StringValue("cosine"),
		DeletionProtection: types.StringValue("disabled"),
		VectorType:         types.StringValue("dense"),
		Tags:               types.MapValueMust(types.StringType, map[string]attr.Value{}),
		Host:               types.StringValue("https://my-index.example.com"),
		Spec:               spec,
		Status:             types.ObjectNull(attrTypesOf(t, s, "status")),
		Embed:              types.ObjectNull(attrTypesOf(t, s, "embed")),
		Timeouts:           timeouts.Value{Object: types.ObjectNull(timeoutsType.AttrTypes)},
	}
}

func withEmbed(t *testing.T, s schema.Schema, model models.IndexResourceModel, embedModel, textField string) models.IndexResourceModel {
	t.Helper()
	embed, d := types.ObjectValueFrom(context.Background(), attrTypesOf(t, s, "embed"), models.IndexEmbedResourceModel{
		Model:                    types.StringValue(embedModel),
		Dimension:                types.Int32Value(1024),
		Metric:                   types.StringValue("cosine"),
		VectorType:               types.StringValue("dense"),
		FieldMap:                 types.MapValueMust(types.StringType, map[string]attr.Value{"text": types.StringValue(textField)}),
		ReadParameters:           types.MapNull(types.StringType),
		WriteParameters:          types.MapNull(types.StringType),
		EffectiveReadParameters:  types.MapNull(types.StringType),
		EffectiveWriteParameters: types.MapNull(types.StringType),
	})
	if d.HasError() {
		t.Fatalf("building embed: %v", d)
	}
	model.Embed = embed
	return model
}

func withPod(t *testing.T, s schema.Schema, model models.IndexResourceModel, podType string, replicas int64) models.IndexResourceModel {
	t.Helper()
	spec, d := types.ObjectValueFrom(context.Background(), attrTypesOf(t, s, "spec"), models.IndexSpecModel{
		Pod: &models.IndexPodSpecModel{
			Environment:      types.StringValue("us-west4-gcp"),
			Replicas:         types.Int64Value(replicas),
			ShardCount:       types.Int64Value(1),
			PodType:          types.StringValue(podType),
			PodCount:         types.Int64Value(replicas),
			MetadataConfig:   types.ObjectNull(models.IndexMetadataConfigModel{}.AttrTypes()),
			SourceCollection: types.StringNull(),
		},
	})
	if d.HasError() {
		t.Fatalf("building pod spec: %v", d)
	}
	model.Spec = spec
	return model
}

func withServerlessMetadataSchema(t *testing.T, s schema.Schema, model models.IndexResourceModel) models.IndexResourceModel {
	t.Helper()
	return withServerlessMetadataSchemaField(t, s, model, "genre")
}

func withServerlessMetadataSchemaField(t *testing.T, s schema.Schema, model models.IndexResourceModel, field string) models.IndexResourceModel {
	t.Helper()
	ctx := context.Background()
	fields, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: models.IndexMetadataSchemaFieldModel{}.AttrTypes()},
		map[string]models.IndexMetadataSchemaFieldModel{field: {Filterable: types.BoolValue(true)}})
	if d.HasError() {
		t.Fatalf("building schema fields: %v", d)
	}
	metadataSchema, d := types.ObjectValueFrom(ctx, models.IndexMetadataSchemaModel{}.AttrTypes(), models.IndexMetadataSchemaModel{Fields: fields})
	if d.HasError() {
		t.Fatalf("building schema: %v", d)
	}
	spec, d := types.ObjectValueFrom(ctx, attrTypesOf(t, s, "spec"), models.IndexSpecModel{
		Serverless: &models.IndexServerlessSpecModel{
			Cloud:        types.StringValue("aws"),
			Region:       types.StringValue("us-west-2"),
			ReadCapacity: types.ObjectNull(models.IndexReadCapacityResourceModel{}.AttrTypes()),
			Schema:       metadataSchema,
		},
	})
	if d.HasError() {
		t.Fatalf("building spec: %v", d)
	}
	model.Spec = spec
	return model
}

// modifyPlan calls ModifyPlan with the given models; a nil state plans a create.
func modifyPlan(t *testing.T, s schema.Schema, config models.IndexResourceModel, state *models.IndexResourceModel) resource.ModifyPlanResponse {
	t.Helper()
	ctx := context.Background()
	schemaType := s.Type().TerraformType(ctx)

	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(schemaType, nil)}
	if d := plan.Set(ctx, &config); d.HasError() {
		t.Fatalf("setting plan: %v", d)
	}
	configValue := tfsdk.Config{Schema: s, Raw: plan.Raw.Copy()}
	priorState := tfsdk.State{Schema: s, Raw: tftypes.NewValue(schemaType, nil)}
	if state != nil {
		if d := priorState.Set(ctx, state); d.HasError() {
			t.Fatalf("setting state: %v", d)
		}
	}

	resp := resource.ModifyPlanResponse{Plan: plan}
	r := &IndexResource{PineconeResource: &PineconeResource{}}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Config: configValue, Plan: plan, State: priorState}, &resp)
	return resp
}

func runModifyPlan(t *testing.T, s schema.Schema, config models.IndexResourceModel, state *models.IndexResourceModel) []string {
	t.Helper()
	resp := modifyPlan(t, s, config, state)
	var summaries []string
	for _, d := range resp.Diagnostics.Errors() {
		summaries = append(summaries, d.Summary())
	}
	return summaries
}

func TestIndexResourceModifyPlan(t *testing.T) {
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	integrated := withEmbed(t, s, base, "multilingual-e5-large", "chunk_text")
	pod := withPod(t, s, base, "p1.x2", 1)
	withMetadataSchema := withServerlessMetadataSchema(t, s, base)

	renamed := func(m models.IndexResourceModel) models.IndexResourceModel {
		m.Name = types.StringValue("my-renamed-index")
		return m
	}

	tests := []struct {
		name    string
		config  models.IndexResourceModel
		state   *models.IndexResourceModel
		wantErr string
	}{
		{name: "create serverless", config: base},
		{name: "create integrated", config: integrated},
		{name: "create integrated with metadata schema", config: withServerlessMetadataSchema(t, s, integrated)},
		{name: "create pod", config: pod, wantErr: podCreateUnsupportedSummary},
		{name: "create serverless with metadata schema", config: withMetadataSchema, wantErr: "Metadata schema isn't supported"},
		{name: "update with no changes", config: base, state: &base},
		{name: "keep existing metadata schema", config: withMetadataSchema, state: &withMetadataSchema},
		{name: "remove metadata schema", config: base, state: &withMetadataSchema},
		{name: "change metadata schema", config: withServerlessMetadataSchemaField(t, s, base, "year"), state: &withMetadataSchema, wantErr: "Metadata schema isn't supported"},
		{name: "add embed", config: integrated, state: &base, wantErr: "embed can't be added to an existing index"},
		{name: "add embed while replacing", config: renamed(integrated), state: &base},
		{name: "remove embed", config: base, state: &integrated, wantErr: "embed can't be removed from an existing index"},
		{name: "change embed model", config: withEmbed(t, s, base, "llama-text-embed-v2", "chunk_text"), state: &integrated, wantErr: "embed.model can't be changed"},
		{name: "change embed field_map", config: withEmbed(t, s, base, "multilingual-e5-large", "body"), state: &integrated, wantErr: "embed.field_map can't be changed"},
		{name: "scale pod replicas", config: withPod(t, s, base, "p1.x2", 3), state: &pod},
		{name: "scale pod size up", config: withPod(t, s, base, "p1.x4", 1), state: &pod},
		{name: "scale pod size down", config: withPod(t, s, base, "p1.x1", 1), state: &pod, wantErr: "pod_type can't be changed this way"},
		{name: "change pod family", config: withPod(t, s, base, "s1.x2", 1), state: &pod, wantErr: "pod_type can't be changed this way"},
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

func withEmbedReadParameters(t *testing.T, s schema.Schema, model models.IndexResourceModel, configured, effective map[string]string) models.IndexResourceModel {
	t.Helper()
	ctx := context.Background()
	var embed models.IndexEmbedResourceModel
	if d := model.Embed.As(ctx, &embed, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding embed: %v", d)
	}
	embed.ReadParameters = types.MapValueMust(types.StringType, stringValues(configured))
	embed.EffectiveReadParameters = types.MapValueMust(types.StringType, stringValues(effective))
	var d diag.Diagnostics
	model.Embed, d = types.ObjectValueFrom(ctx, attrTypesOf(t, s, "embed"), embed)
	if d.HasError() {
		t.Fatalf("encoding embed: %v", d)
	}
	return model
}

func stringValues(m map[string]string) map[string]attr.Value {
	values := make(map[string]attr.Value, len(m))
	for k, v := range m {
		values[k] = types.StringValue(v)
	}
	return values
}

func TestIndexResourceModifyPlan_effectiveEmbedParameters(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	integrated := withEmbed(t, s, testIndexModel(t, s), "multilingual-e5-large", "chunk_text")
	effective := map[string]string{"input_type": "query", "truncate": "END"}
	state := withEmbedReadParameters(t, s, integrated, map[string]string{"truncate": "END"}, effective)

	tests := []struct {
		name        string
		configured  map[string]string
		wantUnknown bool
	}{
		{"unchanged", map[string]string{"truncate": "END"}, false},
		{"changed", map[string]string{"truncate": "NONE"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := withEmbedReadParameters(t, s, integrated, tt.configured, effective)
			resp := modifyPlan(t, s, config, &state)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors: %v", resp.Diagnostics)
			}

			var planned types.Map
			if d := resp.Plan.GetAttribute(ctx, path.Root("embed").AtName("effective_read_parameters"), &planned); d.HasError() {
				t.Fatalf("reading plan: %v", d)
			}
			if planned.IsUnknown() != tt.wantUnknown {
				t.Errorf("effective_read_parameters unknown = %v, want %v", planned.IsUnknown(), tt.wantUnknown)
			}

			var plannedWrite types.Map
			if d := resp.Plan.GetAttribute(ctx, path.Root("embed").AtName("effective_write_parameters"), &plannedWrite); d.HasError() {
				t.Fatalf("reading plan: %v", d)
			}
			if plannedWrite.IsUnknown() {
				t.Error("effective_write_parameters is unknown, but write_parameters didn't change")
			}
		})
	}
}

func TestPlansIndexReplacement(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	withMetadataSchema := withServerlessMetadataSchema(t, s, base)
	renamed := base
	renamed.Name = types.StringValue("my-renamed-index")

	tests := []struct {
		name   string
		config models.IndexResourceModel
		state  models.IndexResourceModel
		want   bool
	}{
		{"no changes", base, base, false},
		{"rename", renamed, base, true},
		{"add metadata schema", withMetadataSchema, base, true},
		{"change metadata schema", withServerlessMetadataSchemaField(t, s, base, "year"), withMetadataSchema, true},
		{"remove metadata schema", base, withMetadataSchema, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaType := s.Type().TerraformType(ctx)
			plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(schemaType, nil)}
			state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(schemaType, nil)}
			if d := plan.Set(ctx, &tt.config); d.HasError() {
				t.Fatalf("setting plan: %v", d)
			}
			if d := state.Set(ctx, &tt.state); d.HasError() {
				t.Fatalf("setting state: %v", d)
			}
			got, d := plansIndexReplacement(ctx, resource.ModifyPlanRequest{Plan: plan, State: state})
			if d.HasError() {
				t.Fatalf("unexpected errors: %v", d)
			}
			if got != tt.want {
				t.Errorf("plansIndexReplacement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequiresReplaceUnlessRemoved(t *testing.T) {
	attrTypes := models.IndexMetadataSchemaModel{}.AttrTypes()
	tests := []struct {
		name string
		plan types.Object
		want bool
	}{
		{"removed", types.ObjectNull(attrTypes), false},
		{"set or changed", types.ObjectValueMust(attrTypes, map[string]attr.Value{
			"fields": types.MapNull(types.ObjectType{AttrTypes: models.IndexMetadataSchemaFieldModel{}.AttrTypes()}),
		}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp objectplanmodifier.RequiresReplaceIfFuncResponse
			requiresReplaceUnlessRemoved(context.Background(), planmodifier.ObjectRequest{PlanValue: tt.plan}, &resp)
			if resp.RequiresReplace != tt.want {
				t.Errorf("RequiresReplace = %v, want %v", resp.RequiresReplace, tt.want)
			}
		})
	}
}

func TestPodTypeChange(t *testing.T) {
	tests := []struct {
		from, to string
		ok       bool
	}{
		{"p1.x1", "p1.x1", true},
		{"p1.x1", "p1.x8", true},
		{"s1.x4", "s1.x2", false},
		{"p1.x2", "p2.x2", false},
		{"p1", "p1.x2", true}, // unparseable: left to the API
	}
	for _, tt := range tests {
		if detail, ok := podTypeChange(tt.from, tt.to); ok != tt.ok {
			t.Errorf("podTypeChange(%q, %q) ok = %v, want %v (%s)", tt.from, tt.to, ok, tt.ok, detail)
		}
	}
}

func TestSemanticTextFieldName(t *testing.T) {
	tests := []struct {
		name     string
		fieldMap types.Map
		want     string
		ok       bool
	}{
		{"text entry", types.MapValueMust(types.StringType, map[string]attr.Value{"text": types.StringValue("chunk_text")}), "chunk_text", true},
		{"single other entry", types.MapValueMust(types.StringType, map[string]attr.Value{"body": types.StringValue("content")}), "content", true},
		{"ambiguous entries", types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("x"), "b": types.StringValue("y")}), "", false},
		{"null", types.MapNull(types.StringType), "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := semanticTextFieldName(tt.fieldMap)
			if got != tt.want || ok != tt.ok {
				t.Errorf("semanticTextFieldName() = %q, %v; want %q, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestIndexReadyRetry(t *testing.T) {
	tests := []struct {
		name      string
		status    *pinecone.IndexStatus
		done      bool
		retryable bool
	}{
		{"no status", nil, false, true},
		{"initializing", &pinecone.IndexStatus{State: pinecone.IndexStatusStateInitializing}, false, true},
		{"ready", &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady}, true, false},
		{"failed", &pinecone.IndexStatus{State: pinecone.IndexStatusStateFailed}, false, false},
		{"initialization failed", &pinecone.IndexStatus{State: pinecone.IndexStatusStateInitializationFailed}, false, false},
		{"disabled", &pinecone.IndexStatus{State: pinecone.IndexStatusStateDisabled}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryErr := indexReadyRetry(&pinecone.Index{Status: tt.status})
			if done := retryErr == nil; done != tt.done {
				t.Fatalf("done = %v, want %v", done, tt.done)
			}
			if retryErr != nil && retryErr.Retryable != tt.retryable {
				t.Errorf("retryable = %v, want %v (%v)", retryErr.Retryable, tt.retryable, retryErr.Err)
			}
			if retryErr != nil && !tt.retryable && !strings.Contains(retryErr.Err.Error(), string(tt.status.State)) {
				t.Errorf("error %q doesn't name the state", retryErr.Err)
			}
		})
	}
}

func TestPodDeploymentMatches(t *testing.T) {
	target := &models.IndexPodSpecModel{PodType: types.StringValue("p1.x2"), Replicas: types.Int64Value(1)}
	three := int32(3)
	tests := []struct {
		name       string
		deployment *pinecone.IndexDeployment
		want       bool
	}{
		{"no deployment", nil, false},
		{"replicas default to one", &pinecone.IndexDeployment{Pod: &pinecone.PodDeployment{PodType: "p1.x2"}}, true},
		{"old pod type", &pinecone.IndexDeployment{Pod: &pinecone.PodDeployment{PodType: "p1.x1"}}, false},
		{"other replicas", &pinecone.IndexDeployment{Pod: &pinecone.PodDeployment{PodType: "p1.x2", Replicas: &three}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := podDeploymentMatches(&pinecone.Index{Deployment: tt.deployment}, target); got != tt.want {
				t.Errorf("podDeploymentMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricDefault(t *testing.T) {
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	base.Metric = types.StringNull()
	sparse := base
	sparse.Dimension = types.Int32Null()
	sparse.VectorType = types.StringValue("sparse")
	unknownVectorType := base
	unknownVectorType.VectorType = types.StringUnknown()
	integrated := withEmbed(t, s, base, "pinecone-sparse-english-v0", "chunk_text")
	configured := base
	configured.Metric = types.StringValue("euclidean")
	// An existing integrated index whose model's metric isn't the default.
	existingIntegrated := integrated
	existingIntegrated.Metric = types.StringValue("dotproduct")

	tests := []struct {
		name   string
		config models.IndexResourceModel
		state  *models.IndexResourceModel
		want   types.String
	}{
		{name: "dense", config: base, want: types.StringValue("cosine")},
		{name: "sparse", config: sparse, want: types.StringValue("dotproduct")},
		{name: "unknown vector_type", config: unknownVectorType, want: types.StringUnknown()},
		{name: "embed uses the model's metric", config: integrated, want: types.StringUnknown()},
		{name: "configured", config: configured, want: types.StringValue("euclidean")},
		// Left unknown, so UseStateForUnknown keeps "dotproduct" instead of planning a replacement.
		{name: "existing index keeps its metric", config: base, state: &existingIntegrated, want: types.StringUnknown()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			config := tfsdk.Config{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
			plan := tfsdk.Plan(config)
			if d := plan.Set(ctx, &tt.config); d.HasError() {
				t.Fatalf("setting config: %v", d)
			}
			config.Raw = plan.Raw
			planValue := tt.config.Metric
			if planValue.IsNull() {
				planValue = types.StringUnknown()
			}
			req := planmodifier.StringRequest{
				Path:        path.Root("metric"),
				Config:      config,
				ConfigValue: tt.config.Metric,
				PlanValue:   planValue,
				State:       tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)},
				StateValue:  types.StringNull(),
			}
			if tt.state != nil {
				if d := req.State.Set(ctx, tt.state); d.HasError() {
					t.Fatalf("setting state: %v", d)
				}
				req.StateValue = tt.state.Metric
			}
			resp := planmodifier.StringResponse{PlanValue: req.PlanValue}
			metricDefault{}.PlanModifyString(ctx, req, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("diagnostics: %v", resp.Diagnostics)
			}
			if !resp.PlanValue.Equal(tt.want) {
				t.Errorf("plan = %v, want %v", resp.PlanValue, tt.want)
			}
		})
	}
}
