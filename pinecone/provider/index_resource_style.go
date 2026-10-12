// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// An index can be described with spec or with schema and deployment. Both describe the same index,
// so moving an existing vector index from spec to schema and deployment only changes how its state
// is recorded: the API isn't called, and the index isn't replaced. Import records vector indexes with
// spec without seeing the configuration, so this is also how an imported index takes a schema
// configuration. spec is planned for deprecation, so an index can't move back to it.

const requiresReplaceUnlessStyleSwitchDescription = "If the value of this attribute changes, Terraform will destroy and recreate " +
	"the resource, unless the configuration changes whether the index is described with spec or with schema and deployment."

const indexStyleSwitchHint = "Moving an index from spec to schema and deployment only changes how Terraform records the index, so " +
	"the configuration has to describe the existing index: its vectors, dimension, metric, and location can't change in place."

// indexStyleSwitch reports whether the configuration describes an existing index in the other
// style than its state.
func indexStyleSwitch(ctx context.Context, config tfsdk.Config, state tfsdk.State) (bool, diag.Diagnostics) {
	if state.Raw.IsNull() {
		return false, nil
	}
	var configSchema, stateSchema types.Object
	diags := config.GetAttribute(ctx, path.Root("schema"), &configSchema)
	diags.Append(state.GetAttribute(ctx, path.Root("schema"), &stateSchema)...)
	return !diags.HasError() && configSchema.IsNull() != stateSchema.IsNull(), diags
}

// The style's attributes go from null to set, or back, when the style switches. ModifyPlan checks
// that a move to schema still describes the same index, and rejects a move back to spec, so these
// don't plan a replacement either way.

func requiresReplaceUnlessStyleSwitchString() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		switching, diags := indexStyleSwitch(ctx, req.Config, req.State)
		resp.Diagnostics.Append(diags...)
		resp.RequiresReplace = !switching
	}, requiresReplaceUnlessStyleSwitchDescription, requiresReplaceUnlessStyleSwitchDescription)
}

func requiresReplaceUnlessStyleSwitchInt32() planmodifier.Int32 {
	return int32planmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.RequiresReplaceIfFuncResponse) {
		switching, diags := indexStyleSwitch(ctx, req.Config, req.State)
		resp.Diagnostics.Append(diags...)
		resp.RequiresReplace = !switching
	}, requiresReplaceUnlessStyleSwitchDescription, requiresReplaceUnlessStyleSwitchDescription)
}

func requiresReplaceUnlessStyleSwitchObject() planmodifier.Object {
	return objectplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
		switching, diags := indexStyleSwitch(ctx, req.Config, req.State)
		resp.Diagnostics.Append(diags...)
		resp.RequiresReplace = !switching
	}, requiresReplaceUnlessStyleSwitchDescription, requiresReplaceUnlessStyleSwitchDescription)
}

// restyleIndexState converts the state of an existing index to the style the configuration
// describes it with. It reports an error when the configuration describes a different index, an
// index schema can't describe, or moves an index back to spec.
func restyleIndexState(ctx context.Context, config, state models.IndexResourceModel) (models.IndexResourceModel, diag.Diagnostics) {
	if config.Schema.IsNull() {
		var diags diag.Diagnostics
		diags.AddAttributeError(path.Root("spec"), "An index can't move back to spec",
			"This index is described with schema and deployment. Describing indexes with dimension, metric, vector_type, spec, and embed "+
				"is planned for deprecation, so an index can move from spec to schema and deployment, but not back. Describe it with schema and deployment again.")
		return state, diags
	}
	return restyleIndexStateToSchema(ctx, config, state)
}

func addIndexMismatch(diags *diag.Diagnostics, p path.Path, detail string) {
	diags.AddAttributeError(p, "The configuration describes a different index", detail+" "+indexStyleSwitchHint+" "+recreateIndexHint)
}

// mismatchedValue reports whether a configured value is known and differs from the index's.
func mismatchedValue(configured, current attr.Value) bool {
	return !configured.IsNull() && !configured.IsUnknown() && !configured.Equal(current)
}

func restyleIndexStateToSchema(ctx context.Context, config, state models.IndexResourceModel) (models.IndexResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaPath, fieldsPath := path.Root("schema"), path.Root("schema").AtName("fields")

	var spec models.IndexSpecModel
	diags.Append(state.Spec.As(ctx, &spec, lenientObjectAs)...)
	if diags.HasError() {
		return state, diags
	}
	switch {
	case spec.Pod != nil:
		diags.AddAttributeError(schemaPath, "Pod-based indexes are described with spec",
			"deployment can't describe a pod-based index, so it keeps spec.pod. Describe this index with spec again.")
	case !state.Embed.IsNull():
		diags.AddAttributeError(schemaPath, "Indexes with integrated embedding are described with spec",
			"schema doesn't support integrated embedding yet, so an index created with embed keeps spec and embed. Describe this index with spec and embed again.")
	case spec.Serverless == nil && spec.BYOC == nil:
		diags.AddAttributeError(schemaPath, "Unknown index deployment",
			"The index's state doesn't say where it runs, so it can't be described with deployment. Describe this index with spec again.")
	}
	if diags.HasError() {
		return state, diags
	}

	dense := state.VectorType.ValueString() != "sparse"
	denseValue := models.IndexResourceSchemaFieldModel{DenseVector: &models.DenseVectorFieldModel{
		Dimension: state.Dimension, Metric: state.Metric, Description: types.StringNull(),
	}}
	sparseValue := models.IndexResourceSchemaFieldModel{SparseVector: &models.SparseVectorFieldModel{Description: types.StringNull()}}

	configFields, d := models.ResourceSchemaFields(ctx, config.Schema)
	diags.Append(d...)
	if diags.HasError() {
		return state, diags
	}
	fields := map[string]models.IndexResourceSchemaFieldModel{}
	if configFields == nil {
		if dense {
			fields[models.ReservedDenseFieldName] = denseValue
		} else {
			fields[models.ReservedSparseFieldName] = sparseValue
		}
	}
	for _, name := range slices.Sorted(maps.Keys(configFields)) {
		field, fieldPath := configFields[name], fieldsPath.AtMapKey(name)
		switch name {
		case models.ReservedDenseFieldName:
			switch {
			case field.DenseVector == nil:
				addIndexMismatch(&diags, fieldPath, fmt.Sprintf("%s is the reserved dense vector field, so it's a dense_vector field.", name))
			case !dense:
				addIndexMismatch(&diags, fieldPath, fmt.Sprintf("The index stores sparse vectors only, so its schema has %s and no %s.", models.ReservedSparseFieldName, name))
			case mismatchedValue(field.DenseVector.Dimension, state.Dimension):
				addIndexMismatch(&diags, fieldPath.AtName("dense_vector").AtName("dimension"),
					fmt.Sprintf("The index's dimension is %s, and the configuration sets %s.", state.Dimension, field.DenseVector.Dimension))
			case mismatchedValue(field.DenseVector.Metric, state.Metric):
				addIndexMismatch(&diags, fieldPath.AtName("dense_vector").AtName("metric"),
					fmt.Sprintf("The index's metric is %s, and the configuration sets %s.", state.Metric, field.DenseVector.Metric))
			case mismatchedValue(field.DenseVector.Description, types.StringNull()):
				addIndexMismatch(&diags, fieldPath.AtName("dense_vector").AtName("description"), "A field's description can only be set when the index is created.")
			}
			fields[name] = denseValue
		case models.ReservedSparseFieldName:
			switch {
			case field.SparseVector == nil:
				addIndexMismatch(&diags, fieldPath, fmt.Sprintf("%s is the reserved sparse vector field, so it's a sparse_vector field.", name))
			case mismatchedValue(field.SparseVector.Description, types.StringNull()):
				addIndexMismatch(&diags, fieldPath.AtName("sparse_vector").AtName("description"), "A field's description can only be set when the index is created.")
			}
			fields[name] = sparseValue
		default:
			addIndexMismatch(&diags, fieldPath, fmt.Sprintf("The index is a vector index, and the field %q describes a document index. "+
				"Describe a vector index with the reserved fields %s and %s only.", name, models.ReservedDenseFieldName, models.ReservedSparseFieldName))
		}
	}
	if configFields != nil {
		if _, ok := configFields[models.ReservedDenseFieldName]; dense && !ok {
			addIndexMismatch(&diags, fieldsPath, fmt.Sprintf("The index stores dense vectors, so its schema has %s.", models.ReservedDenseFieldName))
		}
		if _, ok := configFields[models.ReservedSparseFieldName]; !dense && !ok {
			addIndexMismatch(&diags, fieldsPath, fmt.Sprintf("The index stores sparse vectors, so its schema has %s.", models.ReservedSparseFieldName))
		}
	}

	var configDeployment models.IndexResourceDeploymentModel
	if !config.Deployment.IsNull() && !config.Deployment.IsUnknown() {
		diags.Append(config.Deployment.As(ctx, &configDeployment, lenientObjectAs)...)
	}
	var deployment models.IndexResourceDeploymentModel
	var readCapacity types.Object
	deploymentPath := path.Root("deployment")
	if spec.Serverless != nil {
		deployment.Managed = &models.ManagedDeploymentModel{Cloud: spec.Serverless.Cloud, Region: spec.Serverless.Region, Environment: types.StringUnknown()}
		readCapacity = spec.Serverless.ReadCapacity
		switch {
		case configDeployment.Byoc != nil:
			addIndexMismatch(&diags, deploymentPath.AtName("byoc"), "The index is a serverless index, so it's described with deployment.managed.")
		case configDeployment.Managed == nil:
		case mismatchedValue(configDeployment.Managed.Cloud, spec.Serverless.Cloud):
			addIndexMismatch(&diags, deploymentPath.AtName("managed").AtName("cloud"),
				fmt.Sprintf("The index runs on %s, and the configuration sets %s.", spec.Serverless.Cloud, configDeployment.Managed.Cloud))
		case mismatchedValue(configDeployment.Managed.Region, spec.Serverless.Region):
			addIndexMismatch(&diags, deploymentPath.AtName("managed").AtName("region"),
				fmt.Sprintf("The index runs in %s, and the configuration sets %s.", spec.Serverless.Region, configDeployment.Managed.Region))
		}
	} else {
		deployment.Byoc = &models.ByocDeploymentModel{Environment: spec.BYOC.Environment}
		readCapacity = spec.BYOC.ReadCapacity
		switch {
		case configDeployment.Managed != nil:
			addIndexMismatch(&diags, deploymentPath.AtName("managed"), "The index is a BYOC index, so it's described with deployment.byoc.")
		case configDeployment.Byoc == nil:
		case mismatchedValue(configDeployment.Byoc.Environment, spec.BYOC.Environment):
			addIndexMismatch(&diags, deploymentPath.AtName("byoc").AtName("environment"),
				fmt.Sprintf("The index runs in the BYOC environment %s, and the configuration sets %s.", spec.BYOC.Environment, configDeployment.Byoc.Environment))
		}
	}
	if diags.HasError() {
		return state, diags
	}

	restyled := state
	fieldsMap, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: models.IndexResourceSchemaFieldModel{}.AttrTypes()}, fields)
	diags.Append(d...)
	restyled.Schema, d = types.ObjectValueFrom(ctx, models.IndexResourceSchemaModel{}.AttrTypes(), models.IndexResourceSchemaModel{Fields: fieldsMap})
	diags.Append(d...)
	restyled.Deployment, d = types.ObjectValueFrom(ctx, models.IndexResourceDeploymentModel{}.AttrTypes(), deployment)
	diags.Append(d...)
	restyled.ReadCapacity = readCapacity
	restyled.Spec = types.ObjectNull(state.Spec.AttributeTypes(ctx))
	restyled.Dimension, restyled.Metric, restyled.VectorType = types.Int32Null(), types.StringNull(), types.StringNull()
	return restyled, diags
}

// indexStyleAttributes are the attributes that describe an index in one style or the other.
var indexStyleAttributes = []string{"schema", "deployment", "read_capacity", "spec", "dimension", "metric", "vector_type"}

// planIndexStyleSwitch completes the plan for an index switching style, as if the index had been
// recorded in the new style before. The new style's computed values have no prior state to keep and
// plan as unknown, while the old style's computed values keep their prior state though the new
// style leaves them null: both take the restyled state's values. Any attribute in
// indexReplacePaths that still differs from the restyled state is an error, since the switch must
// leave the index as it is.
func planIndexStyleSwitch(ctx context.Context, config tfsdk.Config, plan *tfsdk.Plan, restyled models.IndexResourceModel) diag.Diagnostics {
	prior := tfsdk.State{Schema: plan.Schema}
	diags := prior.Set(ctx, &restyled)
	if diags.HasError() {
		return diags
	}

	// fillUnknown replaces the unknown style attributes in value with from's value at the same path.
	fillUnknown := func(value, from tftypes.Value, onlyKnown bool) (tftypes.Value, error) {
		return tftypes.Transform(value, func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
			if v.IsKnown() || len(p.Steps()) == 0 {
				return v, nil
			}
			if name, ok := p.Steps()[0].(tftypes.AttributeName); !ok || !slices.Contains(indexStyleAttributes, string(name)) {
				return v, nil
			}
			if found, _, err := tftypes.WalkAttributePath(from, p); err == nil {
				if fromValue, ok := found.(tftypes.Value); ok && (!onlyKnown || (fromValue.IsKnown() && !fromValue.IsNull())) {
					return fromValue, nil
				}
			}
			return v, nil
		})
	}

	filled, err := fillUnknown(plan.Raw, prior.Raw, true)
	if err != nil {
		diags.AddError("Failed to plan index", err.Error())
		return diags
	}
	plan.Raw = filled
	for _, name := range indexStyleAttributes {
		var configured, current attr.Value
		diags.Append(config.GetAttribute(ctx, path.Root(name), &configured)...)
		diags.Append(prior.GetAttribute(ctx, path.Root(name), &current)...)
		if diags.HasError() {
			return diags
		}
		if configured.IsNull() {
			diags.Append(plan.SetAttribute(ctx, path.Root(name), current)...)
		}
	}
	if diags.HasError() {
		return diags
	}

	// A computed value the restyled state can't know yet, such as deployment.managed.environment,
	// has no prior value in the new style, so UseStateForUnknown plans it as null. It's known after
	// apply instead.
	filled, err = tftypes.Transform(plan.Raw, func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if !v.IsNull() || len(p.Steps()) == 0 {
			return v, nil
		}
		if name, ok := p.Steps()[0].(tftypes.AttributeName); !ok || !slices.Contains(indexStyleAttributes, string(name)) {
			return v, nil
		}
		if found, _, err := tftypes.WalkAttributePath(prior.Raw, p); err == nil {
			if priorValue, ok := found.(tftypes.Value); ok && !priorValue.IsKnown() {
				return priorValue, nil
			}
		}
		return v, nil
	})
	if err != nil {
		diags.AddError("Failed to plan index", err.Error())
		return diags
	}
	plan.Raw = filled

	// Values still unknown on either side, such as a computed value the other style doesn't
	// record, can't contradict each other.
	plannedRaw, err := fillUnknown(filled, prior.Raw, false)
	if err == nil {
		prior.Raw, err = fillUnknown(prior.Raw, plannedRaw, false)
	}
	if err != nil {
		diags.AddError("Failed to plan index", err.Error())
		return diags
	}

	plannedState := tfsdk.Plan{Schema: plan.Schema, Raw: plannedRaw}
	var changed []path.Path
	for _, p := range indexReplacePaths {
		var planned, current attr.Value
		diags.Append(plannedState.GetAttribute(ctx, p, &planned)...)
		diags.Append(prior.GetAttribute(ctx, p, &current)...)
		if diags.HasError() {
			return diags
		}
		if !planned.Equal(current) {
			changed = append(changed, p)
		}
	}
	if len(changed) > 0 {
		names := make([]string, len(changed))
		for i, p := range changed {
			names[i] = p.String()
		}
		addIndexMismatch(&diags, changed[0], fmt.Sprintf("These attributes don't match the index: %s.", strings.Join(names, ", ")))
	}
	return diags
}
