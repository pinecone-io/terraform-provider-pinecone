// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
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
	spec := IndexSpecModel{
		Pod:        pod,
		Serverless: NewIndexServerlessSpecModel(index.Spec.Serverless),
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
		model.Tags = types.MapNull(types.StringType)
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
	spec := IndexSpecModel{
		Pod:        pod,
		Serverless: NewIndexServerlessSpecModel(index.Spec.Serverless),
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
		model.Tags = types.MapNull(types.StringType)
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
	spec := IndexSpecModel{
		Pod:        pod,
		Serverless: NewIndexServerlessSpecModel(index.Spec.Serverless),
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
		model.Tags = types.MapNull(types.StringType)
	}

	return diags
}

type IndexSpecModel struct {
	Pod        *IndexPodSpecModel        `tfsdk:"pod"`
	Serverless *IndexServerlessSpecModel `tfsdk:"serverless"`
}

func (model IndexSpecModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"pod":        types.ObjectType{AttrTypes: IndexPodSpecModel{}.AttrTypes()},
		"serverless": types.ObjectType{AttrTypes: IndexServerlessSpecModel{}.AttrTypes()},
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

		if model.FieldMap != nil {
			fieldMap, diags := types.MapValueFrom(ctx, types.StringType, model.FieldMap)
			if diags.HasError() {
				return nil, diags
			}
			newModel.FieldMap = fieldMap
		} else {
			newModel.FieldMap = types.MapNull(types.StringType)
		}

		if model.ReadParameters != nil {
			readParameters, diags := types.MapValueFrom(ctx, types.StringType, model.ReadParameters)
			if diags.HasError() {
				return nil, diags
			}
			newModel.ReadParameters = readParameters
		}

		if model.WriteParameters != nil {
			writeParameters, diags := types.MapValueFrom(ctx, types.StringType, model.WriteParameters)
			if diags.HasError() {
				return nil, diags
			}
			newModel.WriteParameters = writeParameters
		}

		return newModel, nil
	}
	return nil, nil
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

type IndexServerlessSpecModel struct {
	Cloud  types.String `tfsdk:"cloud"`
	Region types.String `tfsdk:"region"`
}

func NewIndexServerlessSpec(spec *IndexServerlessSpecModel) *pinecone.ServerlessSpec {
	if spec != nil {
		return &pinecone.ServerlessSpec{
			Cloud:  pinecone.Cloud(spec.Cloud.String()),
			Region: spec.Region.ValueString(),
		}
	}
	return nil
}

func NewIndexServerlessSpecModel(spec *pinecone.ServerlessSpec) *IndexServerlessSpecModel {
	if spec != nil {
		return &IndexServerlessSpecModel{
			Cloud:  types.StringValue(string(spec.Cloud)),
			Region: types.StringValue(spec.Region),
		}
	}
	return nil
}

func (model IndexServerlessSpecModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cloud":  types.StringType,
		"region": types.StringType,
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
