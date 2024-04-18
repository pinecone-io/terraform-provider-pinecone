// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

type ProjectsModel struct {
	Id       types.String   `tfsdk:"id"`
	Projects []ProjectModel `tfsdk:"projects"`
}

type ProjectModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (model *ProjectModel) Read(ctx context.Context, project *pinecone.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(project.Id.String())
	model.Name = types.StringValue(project.Name)

	return diags
}

// IndexResourceModel defined the Index model for the resource.
// type ProjectResourceModel struct {
// 	Id        types.String   `tfsdk:"id"`
// 	Name      types.String   `tfsdk:"name"`
// 	Dimension types.Int64    `tfsdk:"dimension"`
// 	Metric    types.String   `tfsdk:"metric"`
// 	Host      types.String   `tfsdk:"host"`
// 	Spec      types.Object   `tfsdk:"spec"`
// 	Status    types.Object   `tfsdk:"status"`
// 	Timeouts  timeouts.Value `tfsdk:"timeouts"`
// }

// func (model *IndexResourceModel) Read(ctx context.Context, index *pinecone.Index) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	model.Id = types.StringValue(index.Name)
// 	model.Name = types.StringValue(index.Name)
// 	model.Dimension = types.Int64Value(int64(index.Dimension))
// 	model.Metric = types.StringValue(index.Metric.String())
// 	model.Host = types.StringValue(index.Host)

// 	pod, diags := NewIndexPodSpecModel(ctx, index.Spec.Pod)
// 	if diags.HasError() {
// 		return diags
// 	}
// 	spec := IndexSpecModel{
// 		Pod:        pod,
// 		Serverless: NewIndexServerlessSpecModel(index.Spec.Serverless),
// 	}

// 	model.Spec, diags = types.ObjectValueFrom(ctx, IndexSpecModel{}.AttrTypes(), spec)
// 	if diags.HasError() {
// 		return diags
// 	}

// 	model.Status, diags = types.ObjectValueFrom(ctx, IndexStatusModel{}.AttrTypes(), IndexStatusModel{
// 		Ready: types.BoolValue(index.Status.Ready),
// 		State: types.StringValue(index.Status.State.String()),
// 	})
// 	if diags.HasError() {
// 		return diags
// 	}

// 	return diags
// }

// // IndexDatasourceeModel defined the Index model for the datasource.
// type IndexDatasourceModel struct {
// 	Id        types.String `tfsdk:"id"`
// 	Name      types.String `tfsdk:"name"`
// 	Dimension types.Int64  `tfsdk:"dimension"`
// 	Metric    types.String `tfsdk:"metric"`
// 	Host      types.String `tfsdk:"host"`
// 	Spec      types.Object `tfsdk:"spec"`
// 	Status    types.Object `tfsdk:"status"`
// }

// func (model *IndexDatasourceModel) Read(ctx context.Context, index *pinecone.Index) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	model.Id = types.StringValue(index.Name)
// 	model.Name = types.StringValue(index.Name)
// 	model.Dimension = types.Int64Value(int64(index.Dimension))
// 	model.Metric = types.StringValue(index.Metric.String())
// 	model.Host = types.StringValue(index.Host)

// 	pod, diags := NewIndexPodSpecModel(ctx, index.Spec.Pod)
// 	if diags.HasError() {
// 		return diags
// 	}
// 	spec := IndexSpecModel{
// 		Pod:        pod,
// 		Serverless: NewIndexServerlessSpecModel(index.Spec.Serverless),
// 	}

// 	model.Spec, diags = types.ObjectValueFrom(ctx, IndexSpecModel{}.AttrTypes(), spec)
// 	if diags.HasError() {
// 		return diags
// 	}

// 	model.Status, diags = types.ObjectValueFrom(ctx, IndexStatusModel{}.AttrTypes(), IndexStatusModel{
// 		Ready: types.BoolValue(index.Status.Ready),
// 		State: types.StringValue(index.Status.State.String()),
// 	})
// 	if diags.HasError() {
// 		return diags
// 	}

// 	return diags
// }

// type IndexSpecModel struct {
// 	Pod        *IndexPodSpecModel        `tfsdk:"pod"`
// 	Serverless *IndexServerlessSpecModel `tfsdk:"serverless"`
// }

// func (model IndexSpecModel) AttrTypes() map[string]attr.Type {
// 	return map[string]attr.Type{
// 		"pod":        types.ObjectType{AttrTypes: IndexPodSpecModel{}.AttrTypes()},
// 		"serverless": types.ObjectType{AttrTypes: IndexServerlessSpecModel{}.AttrTypes()},
// 	}
// }

// type IndexPodSpecModel struct {
// 	Environment      types.String `tfsdk:"environment"`
// 	Replicas         types.Int64  `tfsdk:"replicas"`
// 	Shards           types.Int64  `tfsdk:"shards"`
// 	PodType          types.String `tfsdk:"pod_type"`
// 	Pods             types.Int64  `tfsdk:"pods"`
// 	MetadataConfig   types.Object `tfsdk:"metadata_config"`
// 	SourceCollection types.String `tfsdk:"source_collection"`
// }

// func NewIndexPodSpec(ctx context.Context, spec *IndexPodSpecModel) (*pinecone.IndexPodSpec, diag.Diagnostics) {
// 	if spec != nil {
// 		newSpec := &pinecone.IndexPodSpec{
// 			Environment: spec.Environment.ValueString(),
// 			Pods:        int(spec.Pods.ValueInt64()),
// 			PodType:     spec.PodType.ValueString(),
// 			Replicas:    int(spec.Replicas.ValueInt64()),
// 			Shards:      int(spec.Shards.ValueInt64()),
// 		}

// 		var metadataConfig pinecone.IndexMetadataConfig
// 		if !spec.MetadataConfig.IsUnknown() {
// 			diags := spec.MetadataConfig.As(ctx, &metadataConfig, basetypes.ObjectAsOptions{})
// 			if diags.HasError() {
// 				return nil, diags
// 			}
// 		}
// 		newSpec.MetadataConfig = metadataConfig
// 		return newSpec, nil
// 	}
// 	return nil, nil
// }

// func NewIndexPodSpecModel(ctx context.Context, spec *pinecone.IndexPodSpec) (*IndexPodSpecModel, diag.Diagnostics) {
// 	if spec != nil {
// 		newSpec := &IndexPodSpecModel{
// 			Environment:      types.StringValue(spec.Environment),
// 			Pods:             types.Int64Value(int64(spec.Pods)),
// 			PodType:          types.StringValue(spec.PodType),
// 			Replicas:         types.Int64Value(int64(spec.Replicas)),
// 			Shards:           types.Int64Value(int64(spec.Shards)),
// 			SourceCollection: types.StringValue(spec.SourceCollection),
// 		}

// 		indexed, diags := types.ListValueFrom(ctx, types.StringType, spec.MetadataConfig.Indexed)
// 		if diags.HasError() {
// 			return nil, diags
// 		}
// 		metadataConfig := &IndexMetadataConfigModel{
// 			Indexed: indexed,
// 		}
// 		newSpec.MetadataConfig, diags = types.ObjectValueFrom(ctx, IndexMetadataConfigModel{}.AttrTypes(), metadataConfig)
// 		if diags.HasError() {
// 			return nil, diags
// 		}
// 		return newSpec, diags
// 	}
// 	return nil, nil
// }

// func (model IndexPodSpecModel) AttrTypes() map[string]attr.Type {
// 	return map[string]attr.Type{
// 		"environment":       types.StringType,
// 		"replicas":          types.Int64Type,
// 		"shards":            types.Int64Type,
// 		"pod_type":          types.StringType,
// 		"pods":              types.Int64Type,
// 		"metadata_config":   types.ObjectType{AttrTypes: IndexMetadataConfigModel{}.AttrTypes()},
// 		"source_collection": types.StringType,
// 	}
// }

// type IndexMetadataConfigModel struct {
// 	Indexed types.List `tfsdk:"indexed"`
// }

// func (metadataConfig IndexMetadataConfigModel) AttrTypes() map[string]attr.Type {
// 	return map[string]attr.Type{
// 		"indexed": types.ListType{ElemType: types.StringType},
// 	}
// }

// type IndexServerlessSpecModel struct {
// 	Cloud  types.String `tfsdk:"cloud"`
// 	Region types.String `tfsdk:"region"`
// }

// func NewIndexServerlessSpec(spec *IndexServerlessSpecModel) *pinecone.IndexServerlessSpec {
// 	if spec != nil {
// 		return &pinecone.IndexServerlessSpec{
// 			Cloud:  spec.Cloud.ValueString(),
// 			Region: spec.Region.ValueString(),
// 		}
// 	}
// 	return nil
// }

// func NewIndexServerlessSpecModel(spec *pinecone.IndexServerlessSpec) *IndexServerlessSpecModel {
// 	if spec != nil {
// 		return &IndexServerlessSpecModel{
// 			Cloud:  types.StringValue(spec.Cloud),
// 			Region: types.StringValue(spec.Region),
// 		}
// 	}
// 	return nil
// }

// func (model IndexServerlessSpecModel) AttrTypes() map[string]attr.Type {
// 	return map[string]attr.Type{
// 		"cloud":  types.StringType,
// 		"region": types.StringType,
// 	}
// }

// type IndexStatusModel struct {
// 	Ready types.Bool   `tfsdk:"ready"`
// 	State types.String `tfsdk:"state"`
// }

// func (status IndexStatusModel) AttrTypes() map[string]attr.Type {
// 	return map[string]attr.Type{
// 		"ready": types.BoolType,
// 		"state": types.StringType,
// 	}
// }

// type IndexesDataSourceModel struct {
// 	Indexes []IndexModel `tfsdk:"indexes"`
// 	Id      types.String `tfsdk:"id"`
// }
