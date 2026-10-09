// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"maps"
	"slices"
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
		Schema:             types.ObjectNull(attrTypesOf(t, s, "schema")),
		Deployment:         types.ObjectNull(attrTypesOf(t, s, "deployment")),
		ReadCapacity:       types.ObjectNull(attrTypesOf(t, s, "read_capacity")),
		CmekId:             types.StringNull(),
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

type readCapacityMode int

const (
	noReadCapacity readCapacityMode = iota
	onDemandReadCapacity
	dedicatedReadCapacity
)

func readCapacityValue(t *testing.T, attrTypes map[string]attr.Type, mode readCapacityMode) types.Object {
	t.Helper()
	if mode == noReadCapacity {
		return types.ObjectNull(attrTypes)
	}
	rc := models.IndexReadCapacityResourceModel{
		Dedicated: types.ObjectNull(models.IndexReadCapacityDedicatedResourceModel{}.AttrTypes()),
		OnDemand:  types.ObjectNull(models.IndexReadCapacityOnDemandResourceModel{}.AttrTypes()),
	}
	if mode == dedicatedReadCapacity {
		rc.Dedicated = types.ObjectValueMust(models.IndexReadCapacityDedicatedResourceModel{}.AttrTypes(), map[string]attr.Value{
			"node_type": types.StringValue("b1"), "replicas": types.Int32Value(1), "shards": types.Int32Value(1),
		})
	} else {
		rc.OnDemand = types.ObjectValueMust(models.IndexReadCapacityOnDemandResourceModel{}.AttrTypes(), map[string]attr.Value{})
	}
	obj, d := types.ObjectValueFrom(context.Background(), attrTypes, rc)
	if d.HasError() {
		t.Fatalf("building read_capacity: %v", d)
	}
	return obj
}

func withByocSpec(t *testing.T, s schema.Schema, model models.IndexResourceModel, mode readCapacityMode) models.IndexResourceModel {
	t.Helper()
	spec, d := types.ObjectValueFrom(context.Background(), attrTypesOf(t, s, "spec"), models.IndexSpecModel{
		BYOC: &models.IndexBYOCSpecModel{
			Environment:  types.StringValue("aws-us-east-1-b921"),
			ReadCapacity: readCapacityValue(t, models.IndexReadCapacityResourceModel{}.AttrTypes(), mode),
			Schema:       types.ObjectNull(models.IndexMetadataSchemaModel{}.AttrTypes()),
		},
	})
	if d.HasError() {
		t.Fatalf("building byoc spec: %v", d)
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
	byocDedicated := withByocSpec(t, s, base, dedicatedReadCapacity)
	byocOnDemand := withByocSpec(t, s, base, onDemandReadCapacity)
	byocNoReadCapacity := withByocSpec(t, s, base, noReadCapacity)

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
		{name: "create byoc dedicated", config: byocDedicated},
		{name: "create byoc without read capacity", config: byocNoReadCapacity, wantErr: "BYOC indexes need dedicated read capacity"},
		{name: "create byoc on-demand", config: byocOnDemand, wantErr: "BYOC indexes need dedicated read capacity"},
		{name: "update byoc without read capacity", config: byocNoReadCapacity, state: &byocDedicated},
		{name: "replace byoc without read capacity", config: renamed(byocNoReadCapacity), state: &byocOnDemand, wantErr: "BYOC indexes need dedicated read capacity"},
		{name: "replace byoc dedicated", config: renamed(byocDedicated), state: &byocOnDemand},
		{name: "byoc dedicated to on-demand", config: byocOnDemand, state: &byocDedicated, wantErr: "BYOC indexes can't use on-demand read capacity"},
		{name: "keep byoc on-demand", config: byocOnDemand, state: &byocOnDemand},
		{name: "byoc on-demand to dedicated", config: byocDedicated, state: &byocOnDemand},
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

func TestIndexResourceModifyPlan_reportsEveryCreateError(t *testing.T) {
	s := indexResourceSchema(t)
	ctx := context.Background()
	fields, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: models.IndexMetadataSchemaFieldModel{}.AttrTypes()},
		map[string]models.IndexMetadataSchemaFieldModel{"genre": {Filterable: types.BoolValue(true)}})
	if d.HasError() {
		t.Fatalf("building schema fields: %v", d)
	}
	spec, d := types.ObjectValueFrom(ctx, attrTypesOf(t, s, "spec"), models.IndexSpecModel{
		BYOC: &models.IndexBYOCSpecModel{
			Environment:  types.StringValue("aws-us-east-1-b921"),
			ReadCapacity: types.ObjectNull(models.IndexReadCapacityResourceModel{}.AttrTypes()),
			Schema:       types.ObjectValueMust(models.IndexMetadataSchemaModel{}.AttrTypes(), map[string]attr.Value{"fields": fields}),
		},
	})
	if d.HasError() {
		t.Fatalf("building byoc spec: %v", d)
	}
	config := testIndexModel(t, s)
	config.Spec = spec

	errs := runModifyPlan(t, s, config, nil)
	want := []string{"BYOC indexes need dedicated read capacity", "Metadata schema isn't supported"}
	if !slices.Equal(errs, want) {
		t.Fatalf("errors = %v, want %v", errs, want)
	}
}

func requiresReplace[M interface{ Description(context.Context) string }](mods []M) bool {
	return slices.ContainsFunc(mods, func(m M) bool {
		return strings.Contains(m.Description(context.Background()), "destroy and recreate")
	})
}

func collectReplacePaths(t *testing.T, parent path.Path, attrs map[string]schema.Attribute) []string {
	t.Helper()
	var paths []string
	for name, attribute := range attrs {
		p := parent.AtName(name)
		var replace bool
		switch a := attribute.(type) {
		case schema.StringAttribute:
			replace = requiresReplace(a.PlanModifiers)
		case schema.Int32Attribute:
			replace = requiresReplace(a.PlanModifiers)
		case schema.Int64Attribute:
			replace = requiresReplace(a.PlanModifiers)
		case schema.BoolAttribute:
			replace = requiresReplace(a.PlanModifiers)
		case schema.MapAttribute:
			replace = requiresReplace(a.PlanModifiers)
		case schema.ListAttribute:
			replace = requiresReplace(a.PlanModifiers)
		case schema.MapNestedAttribute:
			replace = requiresReplace(a.PlanModifiers)
		case schema.SingleNestedAttribute:
			replace = requiresReplace(a.PlanModifiers)
			paths = append(paths, collectReplacePaths(t, p, a.Attributes)...)
		default:
			t.Fatalf("%s has unhandled attribute type %T", p, attribute)
		}
		if replace {
			paths = append(paths, p.String())
		}
	}
	return paths
}

func TestIndexReplacePathsMatchSchema(t *testing.T) {
	s := indexResourceSchema(t)
	got := collectReplacePaths(t, path.Empty(), s.Attributes)
	var want []string
	for _, p := range indexReplacePaths {
		want = append(want, p.String())
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("attributes with RequiresReplace = %v, indexReplacePaths = %v", got, want)
	}
}

func TestPlansIndexReplacement(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	withMetadataSchema := withServerlessMetadataSchema(t, s, base)
	renamed := base
	renamed.Name = types.StringValue("my-renamed-index")
	sparse := base
	sparse.VectorType = types.StringValue("sparse")

	tests := []struct {
		name   string
		config models.IndexResourceModel
		state  models.IndexResourceModel
		want   bool
	}{
		{"no changes", base, base, false},
		{"rename", renamed, base, true},
		{"change vector_type", sparse, base, true},
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

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name     string
		old, new map[string]string
		want     map[string]string
	}{
		{name: "unchanged", old: map[string]string{"team": "search"}, new: map[string]string{"team": "search"}, want: nil},
		{name: "both empty", old: map[string]string{}, new: map[string]string{}, want: nil},
		{name: "no prior tags", old: nil, new: map[string]string{}, want: nil},
		{name: "add", old: map[string]string{"team": "search"}, new: map[string]string{"team": "search", "env": "prod"}, want: map[string]string{"env": "prod"}},
		{name: "change", old: map[string]string{"team": "search"}, new: map[string]string{"team": "ranking"}, want: map[string]string{"team": "ranking"}},
		{name: "remove", old: map[string]string{"team": "search", "env": "prod"}, new: map[string]string{"team": "search"}, want: map[string]string{"env": ""}},
		{name: "remove all", old: map[string]string{"team": "search"}, new: map[string]string{}, want: map[string]string{"team": ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeTags(tt.old, tt.new)
			if (got == nil) != (tt.want == nil) || !maps.Equal(got, tt.want) {
				t.Errorf("mergeTags() = %#v, want %#v", got, tt.want)
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

func TestReadCapacityRetry(t *testing.T) {
	int32Ptr := func(v int32) *int32 { return &v }
	stringPtr := func(v string) *string { return &v }
	dedicatedTarget := func(nodeType string, replicas, shards int32) *pinecone.ReadCapacityParams {
		return &pinecone.ReadCapacityParams{Dedicated: &pinecone.ReadCapacityDedicatedConfig{
			NodeType: stringPtr(nodeType),
			Scaling:  &pinecone.ReadCapacityScaling{Manual: &pinecone.ReadCapacityManualScaling{Replicas: int32Ptr(replicas), Shards: int32Ptr(shards)}},
		}}
	}
	dedicated := func(nodeType, state string, replicas, shards *int32) *pinecone.ReadCapacity {
		return &pinecone.ReadCapacity{Dedicated: &pinecone.ReadCapacityDedicated{
			NodeType: stringPtr(nodeType),
			Status:   pinecone.ReadCapacityStatus{State: state, CurrentReplicas: replicas, CurrentShards: shards},
		}}
	}
	onDemandTarget := &pinecone.ReadCapacityParams{OnDemand: &pinecone.ReadCapacityOnDemandConfig{}}
	onDemand := func(state string) *pinecone.ReadCapacity {
		return &pinecone.ReadCapacity{OnDemand: &pinecone.ReadCapacityOnDemand{Status: pinecone.ReadCapacityStatus{State: state}}}
	}
	failed := dedicated("b1", "Error", int32Ptr(1), int32Ptr(1))
	failed.Dedicated.Status.ErrorMessage = stringPtr("insufficient capacity for b1")

	tests := []struct {
		name         string
		target       *pinecone.ReadCapacityParams
		readCapacity *pinecone.ReadCapacity
		done         bool
		retryable    bool
		errContains  string
	}{
		{name: "no target", target: nil, readCapacity: nil, done: true},
		{name: "not reported", target: dedicatedTarget("b1", 2, 1), readCapacity: nil, retryable: true},
		{name: "still on-demand", target: dedicatedTarget("b1", 2, 1), readCapacity: onDemand("Ready"), retryable: true},
		{name: "scaling", target: dedicatedTarget("b1", 2, 1), readCapacity: dedicated("b1", "Scaling", int32Ptr(1), int32Ptr(1)), retryable: true},
		{name: "migrating", target: dedicatedTarget("t1", 1, 1), readCapacity: dedicated("t1", "Migrating", int32Ptr(1), int32Ptr(1)), retryable: true},
		{name: "ready before scaling starts", target: dedicatedTarget("b1", 2, 1), readCapacity: dedicated("b1", "Ready", int32Ptr(1), int32Ptr(1)), retryable: true},
		{name: "replicas not provisioned", target: dedicatedTarget("b1", 1, 1), readCapacity: dedicated("b1", "Ready", nil, nil), retryable: true},
		{name: "old node type", target: dedicatedTarget("t1", 1, 1), readCapacity: dedicated("b1", "Ready", int32Ptr(1), int32Ptr(1)), retryable: true},
		{name: "scaled", target: dedicatedTarget("b1", 2, 3), readCapacity: dedicated("b1", "Ready", int32Ptr(2), int32Ptr(3)), done: true},
		{name: "paused", target: dedicatedTarget("b1", 0, 1), readCapacity: dedicated("b1", "Ready", nil, int32Ptr(1)), done: true},
		{name: "error", target: dedicatedTarget("b1", 1, 1), readCapacity: failed, errContains: "insufficient capacity for b1"},
		{name: "partial target", target: &pinecone.ReadCapacityParams{Dedicated: &pinecone.ReadCapacityDedicatedConfig{
			Scaling: &pinecone.ReadCapacityScaling{Manual: &pinecone.ReadCapacityManualScaling{Replicas: int32Ptr(2)}},
		}}, readCapacity: dedicated("b1", "Ready", int32Ptr(2), int32Ptr(4)), done: true},
		{name: "on-demand ready", target: onDemandTarget, readCapacity: onDemand("Ready"), done: true},
		{name: "on-demand still dedicated", target: onDemandTarget, readCapacity: dedicated("b1", "Ready", int32Ptr(1), int32Ptr(1)), retryable: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryErr := readCapacityRetry(&pinecone.Index{ReadCapacity: tt.readCapacity}, tt.target)
			if done := retryErr == nil; done != tt.done {
				t.Fatalf("done = %v, want %v (%v)", done, tt.done, retryErr)
			}
			if retryErr == nil {
				return
			}
			if retryErr.Retryable != tt.retryable {
				t.Errorf("retryable = %v, want %v (%v)", retryErr.Retryable, tt.retryable, retryErr.Err)
			}
			if tt.errContains != "" && !strings.Contains(retryErr.Err.Error(), tt.errContains) {
				t.Errorf("error %q doesn't contain %q", retryErr.Err, tt.errContains)
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
