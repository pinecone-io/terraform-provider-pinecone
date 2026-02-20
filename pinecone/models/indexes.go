// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v5/pinecone"
)

type IndexModel struct {
	Name               types.String `tfsdk:"name"`
	Dimension          types.Int32  `tfsdk:"dimension"`
	Metric             types.String `tfsdk:"metric"`
	DeletionProtection types.String `tfsdk:"deletion_protection"`
	VectorType         types.String `tfsdk:"vector_type"`
	Tags               types.Map    `tfsdk:"tags"`
	Host               types.String `tfsdk:"host"`
	Spec               types.Object `tfsdk:"spec"`
	Status             types.Object `tfsdk:"status"`
	Embed              types.Object `tfsdk:"embed"`
}

func (model *IndexModel) Read(ctx context.Context, index *pinecone.Index) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Name = types.StringValue(index.Name)
	model.Metric = types.StringValue(string(index.Metric))
	model.VectorType = types.StringValue(index.VectorType)
	model.Host = types.StringValue(index.Host)

	if index.Dimension != nil {
		model.Dimension = types.Int32Value(*index.Dimension)
	} else {
		model.Dimension = types.Int32Null()
	}

	pod, diags := NewIndexPodSpecModel(ctx, index.Spec.Pod)
	if diags.HasError() {
		return diags
	}
	serverless, diags := NewIndexServerlessSpecModel(ctx, index.Spec.Serverless)
	if diags.HasError() {
		return diags
	}
	byoc, diags := NewIndexBYOCSpecModel(ctx, index.Spec.BYOC)
	if diags.HasError() {
		return diags
	}
	spec := IndexSpecModel{
		Pod:        pod,
		Serverless: serverless,
		BYOC:       byoc,
	}

	embed, diags := NewIndexEmbedModel(ctx, index.Embed)
	if diags.HasError() {
		return diags
	}
	if embed != nil {
		model.Embed, diags = types.ObjectValueFrom(ctx, IndexEmbedModel{}.AttrTypes(), embed)
		if diags.HasError() {
			return diags
		}
	} else {
		model.Embed = types.ObjectNull(IndexEmbedModel{}.AttrTypes())
	}

	model.Spec, diags = types.ObjectValueFrom(ctx, IndexSpecModel{}.AttrTypes(), spec)
	if diags.HasError() {
		return diags
	}

	model.Status, diags = types.ObjectValueFrom(ctx, IndexStatusModel{}.AttrTypes(), IndexStatusModel{
		Ready: types.BoolValue(index.Status.Ready),
		State: types.StringValue(string(index.Status.State)),
	})
	if diags.HasError() {
		return diags
	}

	if index.Tags != nil {
		model.Tags, diags = types.MapValueFrom(ctx, types.StringType, index.Tags)
		if diags.HasError() {
			return diags
		}
	} else {
		// API returned no tags - set to empty map with explicit type
		// This handles the case where config has tags = {} and API returns nothing
		model.Tags = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	return diags
}

// IndexResourceModel defined the Index model for the resource.
type IndexResourceModel struct {
	Id                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Dimension          types.Int32    `tfsdk:"dimension"`
	Metric             types.String   `tfsdk:"metric"`
	DeletionProtection types.String   `tfsdk:"deletion_protection"`
	VectorType         types.String   `tfsdk:"vector_type"`
	Tags               types.Map      `tfsdk:"tags"`
	Host               types.String   `tfsdk:"host"`
	Spec               types.Object   `tfsdk:"spec"`
	Status             types.Object   `tfsdk:"status"`
	Embed              types.Object   `tfsdk:"embed"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

func (model *IndexResourceModel) Read(ctx context.Context, index *pinecone.Index) diag.Diagnostics {
	var diags diag.Diagnostics

	if index == nil {
		diags.AddError(
			"Nil Index",
			"Cannot read index state from a nil index pointer",
		)
		return diags
	}

	model.Id = types.StringValue(index.Name)
	model.Name = types.StringValue(index.Name)
	model.Metric = types.StringValue(string(index.Metric))
	model.Host = types.StringValue(index.Host)
	model.DeletionProtection = types.StringValue(string(index.DeletionProtection))
	model.VectorType = types.StringValue(index.VectorType)

	if index.Dimension != nil {
		model.Dimension = types.Int32Value(*index.Dimension)
	} else {
		model.Dimension = types.Int32Null()
	}

	pod, diags := NewIndexPodSpecModel(ctx, index.Spec.Pod)
	if diags.HasError() {
		return diags
	}
	serverless, diags := NewIndexServerlessSpecModel(ctx, index.Spec.Serverless)
	if diags.HasError() {
		return diags
	}
	byoc, diags := NewIndexBYOCSpecModel(ctx, index.Spec.BYOC)
	if diags.HasError() {
		return diags
	}
	spec := IndexSpecModel{
		Pod:        pod,
		Serverless: serverless,
		BYOC:       byoc,
	}

	embed, diags := NewIndexEmbedModel(ctx, index.Embed)
	if diags.HasError() {
		return diags
	}
	if embed != nil {
		model.Embed, diags = types.ObjectValueFrom(ctx, IndexEmbedModel{}.AttrTypes(), embed)
		if diags.HasError() {
			return diags
		}
	} else {
		model.Embed = types.ObjectNull(IndexEmbedModel{}.AttrTypes())
	}

	model.Spec, diags = types.ObjectValueFrom(ctx, IndexSpecModel{}.AttrTypes(), spec)
	if diags.HasError() {
		return diags
	}

	if index.Status != nil {
		model.Status, diags = types.ObjectValueFrom(ctx, IndexStatusModel{}.AttrTypes(), IndexStatusModel{
			Ready: types.BoolValue(index.Status.Ready),
			State: types.StringValue(string(index.Status.State)),
		})
		if diags.HasError() {
			return diags
		}
	} else {
		model.Status = types.ObjectNull(IndexStatusModel{}.AttrTypes())
	}

	if index.Tags != nil {
		model.Tags, diags = types.MapValueFrom(ctx, types.StringType, index.Tags)
		if diags.HasError() {
			return diags
		}
	} else {
		// API returned no tags - set to empty map with explicit type
		// This handles the case where config has tags = {} and API returns nothing
		model.Tags = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	return diags
}

// IndexDatasourceeModel defined the Index model for the datasource.
type IndexDatasourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Dimension          types.Int32  `tfsdk:"dimension"`
	Metric             types.String `tfsdk:"metric"`
	DeletionProtection types.String `tfsdk:"deletion_protection"`
	VectorType         types.String `tfsdk:"vector_type"`
	Tags               types.Map    `tfsdk:"tags"`
	Host               types.String `tfsdk:"host"`
	Spec               types.Object `tfsdk:"spec"`
	Status             types.Object `tfsdk:"status"`
	Embed              types.Object `tfsdk:"embed"`
}

func (model *IndexDatasourceModel) Read(ctx context.Context, index *pinecone.Index) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(index.Name)
	model.Name = types.StringValue(index.Name)
	model.Metric = types.StringValue(string(index.Metric))
	model.Host = types.StringValue(index.Host)
	model.DeletionProtection = types.StringValue(string(index.DeletionProtection))
	model.VectorType = types.StringValue(index.VectorType)

	if index.Dimension != nil {
		model.Dimension = types.Int32Value(*index.Dimension)
	} else {
		model.Dimension = types.Int32Null()
	}

	pod, diags := NewIndexPodSpecModel(ctx, index.Spec.Pod)
	if diags.HasError() {
		return diags
	}
	serverless, diags := NewIndexServerlessSpecModel(ctx, index.Spec.Serverless)
	if diags.HasError() {
		return diags
	}
	byoc, diags := NewIndexBYOCSpecModel(ctx, index.Spec.BYOC)
	if diags.HasError() {
		return diags
	}
	spec := IndexSpecModel{
		Pod:        pod,
		Serverless: serverless,
		BYOC:       byoc,
	}

	embed, diags := NewIndexEmbedModel(ctx, index.Embed)
	if diags.HasError() {
		return diags
	}
	if embed != nil {
		model.Embed, diags = types.ObjectValueFrom(ctx, IndexEmbedModel{}.AttrTypes(), embed)
		if diags.HasError() {
			return diags
		}
	} else {
		model.Embed = types.ObjectNull(IndexEmbedModel{}.AttrTypes())
	}

	model.Spec, diags = types.ObjectValueFrom(ctx, IndexSpecModel{}.AttrTypes(), spec)
	if diags.HasError() {
		return diags
	}

	if index.Status != nil {
		model.Status, diags = types.ObjectValueFrom(ctx, IndexStatusModel{}.AttrTypes(), IndexStatusModel{
			Ready: types.BoolValue(index.Status.Ready),
			State: types.StringValue(string(index.Status.State)),
		})
		if diags.HasError() {
			return diags
		}
	} else {
		model.Status = types.ObjectNull(IndexStatusModel{}.AttrTypes())
	}

	if index.Tags != nil {
		model.Tags, diags = types.MapValueFrom(ctx, types.StringType, index.Tags)
		if diags.HasError() {
			return diags
		}
	} else {
		// API returned no tags - set to empty map with explicit type
		// This handles the case where config has tags = {} and API returns nothing
		model.Tags = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	return diags
}

type IndexSpecModel struct {
	Pod        *IndexPodSpecModel        `tfsdk:"pod"`
	Serverless *IndexServerlessSpecModel `tfsdk:"serverless"`
	BYOC       *IndexBYOCSpecModel       `tfsdk:"byoc"`
}

func (model IndexSpecModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"pod":        types.ObjectType{AttrTypes: IndexPodSpecModel{}.AttrTypes()},
		"serverless": types.ObjectType{AttrTypes: IndexServerlessSpecModel{}.AttrTypes()},
		"byoc":       types.ObjectType{AttrTypes: IndexBYOCSpecModel{}.AttrTypes()},
	}
}

type IndexPodSpecModel struct {
	Environment      types.String `tfsdk:"environment"`
	Replicas         types.Int64  `tfsdk:"replicas"`
	ShardCount       types.Int64  `tfsdk:"shards"`
	PodType          types.String `tfsdk:"pod_type"`
	PodCount         types.Int64  `tfsdk:"pods"`
	MetadataConfig   types.Object `tfsdk:"metadata_config"`
	SourceCollection types.String `tfsdk:"source_collection"`
}

func NewIndexPodSpec(ctx context.Context, spec *IndexPodSpecModel) (*pinecone.PodSpec, diag.Diagnostics) {
	if spec != nil {
		newSpec := &pinecone.PodSpec{
			Environment: spec.Environment.ValueString(),
			PodCount:    int(spec.PodCount.ValueInt64()),
			PodType:     spec.PodType.ValueString(),
			Replicas:    int32(spec.Replicas.ValueInt64()),
			ShardCount:  int32(spec.ShardCount.ValueInt64()),
		}

		var metadataConfig pinecone.PodSpecMetadataConfig
		if !spec.MetadataConfig.IsUnknown() {
			diags := spec.MetadataConfig.As(ctx, &metadataConfig, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, diags
			}
			newSpec.MetadataConfig = &metadataConfig
		}
		return newSpec, nil
	}
	return nil, nil
}

func NewIndexPodSpecModel(ctx context.Context, spec *pinecone.PodSpec) (*IndexPodSpecModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if spec != nil {
		newSpec := &IndexPodSpecModel{
			Environment: types.StringValue(spec.Environment),
			PodCount:    types.Int64Value(int64(spec.PodCount)),
			PodType:     types.StringValue(spec.PodType),
			Replicas:    types.Int64Value(int64(spec.Replicas)),
			ShardCount:  types.Int64Value(int64(spec.ShardCount)),
		}

		if spec.SourceCollection != nil {
			newSpec.SourceCollection = types.StringPointerValue(spec.SourceCollection)
		}

		var indexed basetypes.ListValue
		if spec.MetadataConfig != nil {
			indexed, diags = types.ListValueFrom(ctx, types.StringType, spec.MetadataConfig.Indexed)
			if diags.HasError() {
				return nil, diags
			}
		} else {
			indexed = types.ListNull(types.StringType)
		}
		metadataConfig := &IndexMetadataConfigModel{
			Indexed: indexed,
		}
		newSpec.MetadataConfig, diags = types.ObjectValueFrom(ctx, IndexMetadataConfigModel{}.AttrTypes(), metadataConfig)
		if diags.HasError() {
			return nil, diags
		}
		return newSpec, nil
	}
	return nil, nil
}

func (model IndexPodSpecModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"environment":       types.StringType,
		"replicas":          types.Int64Type,
		"shards":            types.Int64Type,
		"pod_type":          types.StringType,
		"pods":              types.Int64Type,
		"metadata_config":   types.ObjectType{AttrTypes: IndexMetadataConfigModel{}.AttrTypes()},
		"source_collection": types.StringType,
	}
}

func NewIndexEmbedModel(ctx context.Context, model *pinecone.IndexEmbed) (*IndexEmbedModel, diag.Diagnostics) {
	if model != nil {
		newModel := &IndexEmbedModel{
			Model: types.StringValue(model.Model),
		}

		if model.Dimension != nil {
			newModel.Dimension = types.Int32Value(*model.Dimension)
		} else {
			newModel.Dimension = types.Int32Null()
		}

		if model.Metric != nil {
			newModel.Metric = types.StringValue(string(*model.Metric))
		} else {
			newModel.Metric = types.StringNull()
		}

		if model.VectorType != nil {
			newModel.VectorType = types.StringValue(*model.VectorType)
		} else {
			newModel.VectorType = types.StringNull()
		}

		if fieldMap, ok := toMapStringString(model.FieldMap); ok {
			m, diags := types.MapValueFrom(ctx, types.StringType, fieldMap)
			if diags.HasError() {
				return nil, diags
			}
			newModel.FieldMap = m
		} else {
			newModel.FieldMap = types.MapNull(types.StringType)
		}

		if readParams, ok := toMapStringString(model.ReadParameters); ok {
			m, diags := types.MapValueFrom(ctx, types.StringType, readParams)
			if diags.HasError() {
				return nil, diags
			}
			newModel.ReadParameters = m
		} else {
			newModel.ReadParameters = types.MapNull(types.StringType)
		}

		if writeParams, ok := toMapStringString(model.WriteParameters); ok {
			m, diags := types.MapValueFrom(ctx, types.StringType, writeParams)
			if diags.HasError() {
				return nil, diags
			}
			newModel.WriteParameters = m
		} else {
			newModel.WriteParameters = types.MapNull(types.StringType)
		}

		return newModel, nil
	}
	return &IndexEmbedModel{
		Model:           types.StringNull(),
		Dimension:       types.Int32Null(),
		Metric:          types.StringNull(),
		VectorType:      types.StringNull(),
		FieldMap:        types.MapNull(types.StringType),
		ReadParameters:  types.MapNull(types.StringType),
		WriteParameters: types.MapNull(types.StringType),
	}, nil
}

func NewIndexEmbed(ctx context.Context, model *IndexEmbedModel) (*pinecone.IndexEmbed, diag.Diagnostics) {
	if model != nil {
		newModel := &pinecone.IndexEmbed{
			Model:           model.Model.ValueString(),
			Dimension:       model.Dimension.ValueInt32Pointer(),
			VectorType:      model.VectorType.ValueStringPointer(),
			Metric:          (*pinecone.IndexMetric)(model.Metric.ValueStringPointer()),
			FieldMap:        mapAttrToInterfacePtr(model.FieldMap),
			ReadParameters:  mapAttrToInterfacePtr(model.ReadParameters),
			WriteParameters: mapAttrToInterfacePtr(model.WriteParameters),
		}

		return newModel, nil
	}
	return nil, nil
}

type IndexEmbedModel struct {
	Model           types.String `tfsdk:"model"`
	Dimension       types.Int32  `tfsdk:"dimension"`
	Metric          types.String `tfsdk:"metric"`
	VectorType      types.String `tfsdk:"vector_type"`
	FieldMap        types.Map    `tfsdk:"field_map"`
	ReadParameters  types.Map    `tfsdk:"read_parameters"`
	WriteParameters types.Map    `tfsdk:"write_parameters"`
}

func (model IndexEmbedModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"model":            types.StringType,
		"dimension":        types.Int32Type,
		"metric":           types.StringType,
		"vector_type":      types.StringType,
		"field_map":        types.MapType{ElemType: types.StringType},
		"read_parameters":  types.MapType{ElemType: types.StringType},
		"write_parameters": types.MapType{ElemType: types.StringType},
	}
}

type IndexMetadataConfigModel struct {
	Indexed types.List `tfsdk:"indexed"`
}

func (metadataConfig IndexMetadataConfigModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"indexed": types.ListType{ElemType: types.StringType},
	}
}

// ── Serverless spec ──────────────────────────────────────────────────────────

type IndexServerlessSpecModel struct {
	Cloud        types.String `tfsdk:"cloud"`
	Region       types.String `tfsdk:"region"`
	ReadCapacity types.Object `tfsdk:"read_capacity"`
}

func NewIndexServerlessSpec(spec *IndexServerlessSpecModel) *pinecone.ServerlessSpec {
	if spec != nil {
		return &pinecone.ServerlessSpec{
			Cloud:  pinecone.Cloud(spec.Cloud.ValueString()),
			Region: spec.Region.ValueString(),
		}
	}
	return nil
}

func NewIndexServerlessSpecModel(ctx context.Context, spec *pinecone.ServerlessSpec) (*IndexServerlessSpecModel, diag.Diagnostics) {
	if spec == nil {
		return nil, nil
	}
	rc, diags := NewIndexReadCapacityModel(ctx, spec.ReadCapacity)
	if diags.HasError() {
		return nil, diags
	}
	var rcObj types.Object
	if rc != nil {
		rcObj, diags = types.ObjectValueFrom(ctx, IndexReadCapacityModel{}.AttrTypes(), rc)
		if diags.HasError() {
			return nil, diags
		}
	} else {
		rcObj = types.ObjectNull(IndexReadCapacityModel{}.AttrTypes())
	}
	return &IndexServerlessSpecModel{
		Cloud:        types.StringValue(string(spec.Cloud)),
		Region:       types.StringValue(spec.Region),
		ReadCapacity: rcObj,
	}, nil
}

func (model IndexServerlessSpecModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cloud":         types.StringType,
		"region":        types.StringType,
		"read_capacity": types.ObjectType{AttrTypes: IndexReadCapacityModel{}.AttrTypes()},
	}
}

// ── BYOC spec ─────────────────────────────────────────────────────────────────

type IndexBYOCSpecModel struct {
	Environment  types.String `tfsdk:"environment"`
	ReadCapacity types.Object `tfsdk:"read_capacity"`
}

func NewIndexBYOCSpecModel(ctx context.Context, spec *pinecone.BYOCSpec) (*IndexBYOCSpecModel, diag.Diagnostics) {
	if spec == nil {
		return nil, nil
	}
	rc, diags := NewIndexReadCapacityModel(ctx, spec.ReadCapacity)
	if diags.HasError() {
		return nil, diags
	}
	var rcObj types.Object
	if rc != nil {
		rcObj, diags = types.ObjectValueFrom(ctx, IndexReadCapacityModel{}.AttrTypes(), rc)
		if diags.HasError() {
			return nil, diags
		}
	} else {
		rcObj = types.ObjectNull(IndexReadCapacityModel{}.AttrTypes())
	}
	return &IndexBYOCSpecModel{
		Environment:  types.StringValue(spec.Environment),
		ReadCapacity: rcObj,
	}, nil
}

func (model IndexBYOCSpecModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"environment":   types.StringType,
		"read_capacity": types.ObjectType{AttrTypes: IndexReadCapacityModel{}.AttrTypes()},
	}
}

// ── ReadCapacity models ───────────────────────────────────────────────────────

// IndexReadCapacityModel mirrors pinecone.ReadCapacity — the full read-back from API.
type IndexReadCapacityModel struct {
	Dedicated types.Object `tfsdk:"dedicated"`
	OnDemand  types.Object `tfsdk:"on_demand"`
}

func (model IndexReadCapacityModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dedicated": types.ObjectType{AttrTypes: IndexReadCapacityDedicatedModel{}.AttrTypes()},
		"on_demand": types.ObjectType{AttrTypes: IndexReadCapacityOnDemandModel{}.AttrTypes()},
	}
}

type IndexReadCapacityDedicatedModel struct {
	NodeType        types.String `tfsdk:"node_type"`
	Replicas        types.Int32  `tfsdk:"replicas"`
	Shards          types.Int32  `tfsdk:"shards"`
	State           types.String `tfsdk:"state"`
	CurrentReplicas types.Int32  `tfsdk:"current_replicas"`
	CurrentShards   types.Int32  `tfsdk:"current_shards"`
	ErrorMessage    types.String `tfsdk:"error_message"`
}

func (model IndexReadCapacityDedicatedModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"node_type":        types.StringType,
		"replicas":         types.Int32Type,
		"shards":           types.Int32Type,
		"state":            types.StringType,
		"current_replicas": types.Int32Type,
		"current_shards":   types.Int32Type,
		"error_message":    types.StringType,
	}
}

type IndexReadCapacityOnDemandModel struct {
	State           types.String `tfsdk:"state"`
	CurrentReplicas types.Int32  `tfsdk:"current_replicas"`
	CurrentShards   types.Int32  `tfsdk:"current_shards"`
	ErrorMessage    types.String `tfsdk:"error_message"`
}

func (model IndexReadCapacityOnDemandModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"state":            types.StringType,
		"current_replicas": types.Int32Type,
		"current_shards":   types.Int32Type,
		"error_message":    types.StringType,
	}
}

// readCapacityStatusFromSDK maps pinecone.ReadCapacityStatus fields to the four common Terraform status fields.
// types.*PointerValue handles nil by returning the appropriate null value.
func readCapacityStatusFromSDK(s pinecone.ReadCapacityStatus) (state types.String, currentReplicas, currentShards types.Int32, errorMessage types.String) {
	return types.StringValue(s.State),
		types.Int32PointerValue(s.CurrentReplicas),
		types.Int32PointerValue(s.CurrentShards),
		types.StringPointerValue(s.ErrorMessage)
}

// NewIndexReadCapacityModel converts a *pinecone.ReadCapacity (API response) to the Terraform model.
func NewIndexReadCapacityModel(ctx context.Context, rc *pinecone.ReadCapacity) (*IndexReadCapacityModel, diag.Diagnostics) {
	if rc == nil {
		return nil, nil
	}

	model := &IndexReadCapacityModel{
		Dedicated: types.ObjectNull(IndexReadCapacityDedicatedModel{}.AttrTypes()),
		OnDemand:  types.ObjectNull(IndexReadCapacityOnDemandModel{}.AttrTypes()),
	}

	switch {
	case rc.Dedicated != nil:
		d := rc.Dedicated

		var replicas, shards types.Int32
		if d.Scaling != nil && d.Scaling.Manual != nil {
			replicas = types.Int32PointerValue(d.Scaling.Manual.Replicas)
			shards = types.Int32PointerValue(d.Scaling.Manual.Shards)
		} else {
			replicas, shards = types.Int32Null(), types.Int32Null()
		}

		state, currentReplicas, currentShards, errorMessage := readCapacityStatusFromSDK(d.Status)
		dedicated := IndexReadCapacityDedicatedModel{
			NodeType:        types.StringPointerValue(d.NodeType),
			Replicas:        replicas,
			Shards:          shards,
			State:           state,
			CurrentReplicas: currentReplicas,
			CurrentShards:   currentShards,
			ErrorMessage:    errorMessage,
		}

		var diags diag.Diagnostics
		model.Dedicated, diags = types.ObjectValueFrom(ctx, IndexReadCapacityDedicatedModel{}.AttrTypes(), dedicated)
		if diags.HasError() {
			return nil, diags
		}

	case rc.OnDemand != nil:
		state, currentReplicas, currentShards, errorMessage := readCapacityStatusFromSDK(rc.OnDemand.Status)
		onDemand := IndexReadCapacityOnDemandModel{
			State:           state,
			CurrentReplicas: currentReplicas,
			CurrentShards:   currentShards,
			ErrorMessage:    errorMessage,
		}

		var diags diag.Diagnostics
		model.OnDemand, diags = types.ObjectValueFrom(ctx, IndexReadCapacityOnDemandModel{}.AttrTypes(), onDemand)
		if diags.HasError() {
			return nil, diags
		}
	}

	return model, nil
}

// ToReadCapacityParams converts the Terraform model to SDK params for create/update requests.
//
// The presence of the dedicated or on_demand sub-block is what signals the desired mode:
//   - dedicated { ... }  → ReadCapacityParams{Dedicated: ...}
//   - on_demand {}       → ReadCapacityParams{OnDemand: &ReadCapacityOnDemandConfig{}}
//   - neither sub-block  → nil (treat as unset; API will preserve existing or default to OnDemand)
func ToReadCapacityParams(ctx context.Context, rcObj types.Object) (*pinecone.ReadCapacityParams, diag.Diagnostics) {
	if rcObj.IsNull() || rcObj.IsUnknown() {
		return nil, nil
	}

	var model IndexReadCapacityModel
	if diags := rcObj.As(ctx, &model, basetypes.ObjectAsOptions{}); diags.HasError() {
		return nil, diags
	}

	switch {
	case !model.Dedicated.IsNull() && !model.Dedicated.IsUnknown():
		var dedicated IndexReadCapacityDedicatedModel
		if diags := model.Dedicated.As(ctx, &dedicated, basetypes.ObjectAsOptions{}); diags.HasError() {
			return nil, diags
		}

		cfg := &pinecone.ReadCapacityDedicatedConfig{
			NodeType: dedicated.NodeType.ValueStringPointer(),
		}

		if !dedicated.Replicas.IsNull() || !dedicated.Shards.IsNull() {
			cfg.Scaling = &pinecone.ReadCapacityScaling{
				Manual: &pinecone.ReadCapacityManualScaling{
					Replicas: dedicated.Replicas.ValueInt32Pointer(),
					Shards:   dedicated.Shards.ValueInt32Pointer(),
				},
			}
		}

		return &pinecone.ReadCapacityParams{Dedicated: cfg}, nil

	case !model.OnDemand.IsNull() && !model.OnDemand.IsUnknown():
		return &pinecone.ReadCapacityParams{OnDemand: &pinecone.ReadCapacityOnDemandConfig{}}, nil

	default:
		// read_capacity block is present but neither sub-block is set — treat as unset.
		return nil, nil
	}
}

type IndexStatusModel struct {
	Ready types.Bool   `tfsdk:"ready"`
	State types.String `tfsdk:"state"`
}

func (status IndexStatusModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ready": types.BoolType,
		"state": types.StringType,
	}
}

type IndexesDataSourceModel struct {
	Indexes []IndexModel `tfsdk:"indexes"`
	Id      types.String `tfsdk:"id"`
}

func mapAttrToInterfacePtr(attr types.Map) *map[string]interface{} {
	if attr.IsUnknown() || attr.IsNull() {
		return nil
	}

	raw := make(map[string]interface{}, len(attr.Elements()))
	for k, v := range attr.Elements() {
		if sv, ok := v.(basetypes.StringValue); ok {
			raw[k] = sv.ValueString()
		} else {
			raw[k] = v.String()
		}
	}
	return &raw
}

// toMapStringString converts API map values to map[string]string so that
// types.MapValueFrom(..., types.StringType, ...) never sees non-string values,
// which would cause "can't unmarshal tftypes.Number into *string".
//
// The go-pinecone SDK uses *map[string]interface{} for FieldMap, ReadParameters,
// and WriteParameters, so we must handle both map and *map.
//
// Returns a map[string]string and a boolean indicating if the conversion was successful.
func toMapStringString(in interface{}) (map[string]string, bool) {
	if in == nil {
		return nil, false
	}
	switch m := in.(type) {
	case map[string]string:
		return m, true
	case map[string]interface{}:
		out := make(map[string]string, len(m))
		for k, v := range m {
			out[k] = fmt.Sprint(v)
		}
		return out, true
	case *map[string]interface{}:
		if m == nil {
			return nil, false
		}
		return toMapStringString(*m)
	default:
		return nil, false
	}
}
