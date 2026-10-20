// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

func westDeployment() models.IndexResourceDeploymentModel {
	deployment := managedDeployment()
	deployment.Managed.Region = types.StringValue("us-west-2")
	return deployment
}

func TestIndexResourceModifyPlan_styleSwitch(t *testing.T) {
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	vector := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, westDeployment())
	hybrid := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine"), "_sparse_values": sparseField()}, westDeployment())

	sparseSpec := base
	sparseSpec.Dimension, sparseSpec.Metric, sparseSpec.VectorType = types.Int32Null(), types.StringValue("dotproduct"), types.StringValue("sparse")
	sparseSchema := withSchemaMode(t, s, base, schemaFields{"_sparse_values": sparseField()}, westDeployment())

	byocSpec := withByocSpec(t, s, base, dedicatedReadCapacity)
	byocSchema := withReadCapacity(t, s, withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, byocDeployment()), true)
	dedicatedSpec := withServerlessReadCapacity(t, s, base, dedicatedReadCapacity)
	dedicatedSchema := withReadCapacity(t, s, vector, true)

	renamedVector := vector
	renamedVector.Name = types.StringValue("my-renamed-index")
	retaggedVector := vector
	retaggedVector.Tags = types.MapValueMust(types.StringType, map[string]attr.Value{"team": types.StringValue("search")})

	described := func(dimension int32, metric string, deployment models.IndexResourceDeploymentModel) models.IndexResourceModel {
		return withSchemaMode(t, s, base, schemaFields{"_values": denseField(dimension, metric)}, deployment)
	}
	renamed := base
	renamed.Name = types.StringValue("my-renamed-index")
	document := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine"), "body": fullTextField(models.FullTextSearchModel{})}, westDeployment())
	withDescription := withSchemaMode(t, s, base, schemaFields{"_values": {DenseVector: &models.DenseVectorFieldModel{
		Dimension: types.Int32Value(1024), Metric: types.StringValue("cosine"), Description: types.StringValue("embeddings"),
	}}}, westDeployment())

	const mismatch = "The configuration describes a different index"
	const backToSpec = "An index can't move back to spec"
	tests := []struct {
		name    string
		config  models.IndexResourceModel
		state   models.IndexResourceModel
		wantErr string
	}{
		{name: "spec to schema", config: vector, state: base},
		{name: "spec to schema with both reserved fields", config: hybrid, state: base},
		{name: "spec to schema with tags", config: retaggedVector, state: base},
		{name: "spec to schema while renaming", config: renamedVector, state: base},
		{name: "sparse spec to schema", config: sparseSchema, state: sparseSpec},
		{name: "byoc spec to schema", config: byocSchema, state: byocSpec},
		{name: "dedicated spec to schema", config: dedicatedSchema, state: dedicatedSpec},

		{name: "spec to schema with another dimension", config: described(768, "cosine", westDeployment()), state: base, wantErr: mismatch},
		{name: "spec to schema with another metric", config: described(1024, "dotproduct", westDeployment()), state: base, wantErr: mismatch},
		{name: "spec to schema in another region", config: described(1024, "cosine", managedDeployment()), state: base, wantErr: mismatch},
		{name: "spec to schema on byoc", config: described(1024, "cosine", byocDeployment()), state: base, wantErr: mismatch},
		{name: "spec to schema with a document field", config: document, state: base, wantErr: mismatch},
		{name: "spec to schema with a field description", config: withDescription, state: base, wantErr: mismatch},
		{name: "spec to schema without the dense field", config: sparseSchema, state: base, wantErr: mismatch},
		{name: "sparse spec to schema with a dense field", config: vector, state: sparseSpec, wantErr: mismatch},
		{name: "integrated spec to schema", config: vector, state: withEmbed(t, s, base, "multilingual-e5-large", "chunk_text"),
			wantErr: "Legacy integrated indexes are described with spec"},
		{name: "pod spec to schema", config: vector, state: withPod(t, s, base, "p1.x1", 1), wantErr: "Pod-based indexes are described with spec"},
		{name: "dedicated spec to on-demand schema", config: withReadCapacity(t, s, vector, false), state: dedicatedSpec,
			wantErr: "Read capacity can't return to on-demand"},

		{name: "schema to spec", config: base, state: vector, wantErr: backToSpec},
		{name: "sparse schema to spec", config: sparseSpec, state: sparseSchema, wantErr: backToSpec},
		{name: "byoc schema to spec", config: byocSpec, state: byocSchema, wantErr: backToSpec},
		{name: "document schema to spec", config: base, state: document, wantErr: backToSpec},
		{name: "schema to spec while renaming", config: renamed, state: vector},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := runModifyPlan(t, s, tt.config, &tt.state)
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

func TestPlanIndexStyleSwitch(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	state := withServerlessReadCapacity(t, s, base, dedicatedReadCapacity)
	config := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, westDeployment())

	restyled, diags := restyleIndexState(ctx, config, state)
	if diags.HasError() {
		t.Fatalf("restyleIndexState: %v", diags)
	}

	planned := config
	planned.ReadCapacity = types.ObjectUnknown(attrTypesOf(t, s, "read_capacity"))
	planned.Dimension, planned.VectorType = state.Dimension, state.VectorType
	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	if d := plan.Set(ctx, &planned); d.HasError() {
		t.Fatalf("setting plan: %v", d)
	}
	configValue := tfsdk.Config{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	configPlan := tfsdk.Plan{Schema: s, Raw: configValue.Raw}
	if d := configPlan.Set(ctx, &config); d.HasError() {
		t.Fatalf("setting config: %v", d)
	}
	configValue.Raw = configPlan.Raw
	if d := planIndexStyleSwitch(ctx, configValue, &plan, restyled); d.HasError() {
		t.Fatalf("planIndexStyleSwitch: %v", d)
	}

	var got models.IndexResourceModel
	if d := plan.Get(ctx, &got); d.HasError() {
		t.Fatalf("reading plan: %v", d)
	}
	if !got.Dimension.IsNull() || !got.VectorType.IsNull() {
		t.Errorf("dimension, vector_type = %v, %v, want null: schema doesn't record them", got.Dimension, got.VectorType)
	}
	var environment types.String
	if d := plan.GetAttribute(ctx, path.Root("deployment").AtName("managed").AtName("environment"), &environment); d.HasError() {
		t.Fatalf("reading deployment.managed.environment: %v", d)
	}
	if !environment.IsUnknown() {
		t.Errorf("deployment.managed.environment = %v, want unknown: the API reports it after apply", environment)
	}

	var readCapacity types.Object
	if d := plan.GetAttribute(ctx, path.Root("read_capacity"), &readCapacity); d.HasError() {
		t.Fatalf("reading read_capacity: %v", d)
	}
	var stateSpec models.IndexSpecModel
	if d := state.Spec.As(ctx, &stateSpec, lenientObjectAs); d.HasError() {
		t.Fatalf("reading spec: %v", d)
	}
	if !readCapacity.Equal(stateSpec.Serverless.ReadCapacity) {
		t.Errorf("read_capacity = %v, want %v", readCapacity, stateSpec.Serverless.ReadCapacity)
	}
}

func TestRequiresReplaceUnlessStyleSwitch(t *testing.T) {
	ctx := context.Background()
	s := indexResourceSchema(t)
	base := testIndexModel(t, s)
	vector := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, westDeployment())
	moved := withSchemaMode(t, s, base, schemaFields{"_values": denseField(1024, "cosine")}, managedDeployment())

	tests := []struct {
		name          string
		config, state models.IndexResourceModel
		want          bool
	}{
		{"same style", moved, vector, true},
		{"spec to schema", vector, base, false},
		{"schema to spec", base, vector, false},
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

			req := planmodifier.StringRequest{
				Path: path.Root("metric"), Config: tfsdk.Config{Schema: s, Raw: plan.Raw.Copy()}, Plan: plan, State: state,
				ConfigValue: types.StringValue("cosine"), PlanValue: types.StringValue("cosine"), StateValue: types.StringValue("dotproduct"),
			}
			var resp planmodifier.StringResponse
			requiresReplaceUnlessStyleSwitchString().PlanModifyString(ctx, req, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors: %v", resp.Diagnostics)
			}
			if resp.RequiresReplace != tt.want {
				t.Errorf("RequiresReplace = %v, want %v", resp.RequiresReplace, tt.want)
			}
		})
	}
}
