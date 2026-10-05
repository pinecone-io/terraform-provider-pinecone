// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
)

type IndexSchemaModel struct {
	Fields types.Map `tfsdk:"fields"`
}

func (m IndexSchemaModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"fields": types.MapType{ElemType: types.ObjectType{AttrTypes: IndexSchemaFieldModel{}.AttrTypes()}},
	}
}

// IndexSchemaFieldModel is one schema field. Exactly one of its attributes is set, naming the
// field's type.
type IndexSchemaFieldModel struct {
	DenseVector    *DenseVectorFieldModel    `tfsdk:"dense_vector"`
	SparseVector   *SparseVectorFieldModel   `tfsdk:"sparse_vector"`
	SemanticText   *SemanticTextFieldModel   `tfsdk:"semantic_text"`
	String         *StringFieldModel         `tfsdk:"string"`
	StringList     *MetadataFieldModel       `tfsdk:"string_list"`
	Boolean        *MetadataFieldModel       `tfsdk:"boolean"`
	Float          *MetadataFieldModel       `tfsdk:"float"`
	Integer        *MetadataFieldModel       `tfsdk:"integer"`
	LegacyMetadata *LegacyMetadataFieldModel `tfsdk:"legacy_metadata"`
}

func (m IndexSchemaFieldModel) AttrTypes() map[string]attr.Type {
	metadata := types.ObjectType{AttrTypes: MetadataFieldModel{}.AttrTypes()}
	return map[string]attr.Type{
		"dense_vector":    types.ObjectType{AttrTypes: DenseVectorFieldModel{}.AttrTypes()},
		"sparse_vector":   types.ObjectType{AttrTypes: SparseVectorFieldModel{}.AttrTypes()},
		"semantic_text":   types.ObjectType{AttrTypes: SemanticTextFieldModel{}.AttrTypes()},
		"string":          types.ObjectType{AttrTypes: StringFieldModel{}.AttrTypes()},
		"string_list":     metadata,
		"boolean":         metadata,
		"float":           metadata,
		"integer":         metadata,
		"legacy_metadata": types.ObjectType{AttrTypes: LegacyMetadataFieldModel{}.AttrTypes()},
	}
}

type DenseVectorFieldModel struct {
	Dimension   types.Int32  `tfsdk:"dimension"`
	Metric      types.String `tfsdk:"metric"`
	Description types.String `tfsdk:"description"`
}

func (m DenseVectorFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dimension":   types.Int32Type,
		"metric":      types.StringType,
		"description": types.StringType,
	}
}

type SparseVectorFieldModel struct {
	Description types.String `tfsdk:"description"`
}

func (m SparseVectorFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"description": types.StringType,
	}
}

type SemanticTextFieldModel struct {
	Model           types.String `tfsdk:"model"`
	Dimension       types.Int32  `tfsdk:"dimension"`
	Metric          types.String `tfsdk:"metric"`
	ReadParameters  types.Map    `tfsdk:"read_parameters"`
	WriteParameters types.Map    `tfsdk:"write_parameters"`
	Description     types.String `tfsdk:"description"`
}

func (m SemanticTextFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"model":            types.StringType,
		"dimension":        types.Int32Type,
		"metric":           types.StringType,
		"read_parameters":  types.MapType{ElemType: types.StringType},
		"write_parameters": types.MapType{ElemType: types.StringType},
		"description":      types.StringType,
	}
}

type StringFieldModel struct {
	FullTextSearch *FullTextSearchModel `tfsdk:"full_text_search"`
	Filterable     types.Bool           `tfsdk:"filterable"`
	Description    types.String         `tfsdk:"description"`
}

func (m StringFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"full_text_search": types.ObjectType{AttrTypes: FullTextSearchModel{}.AttrTypes()},
		"filterable":       types.BoolType,
		"description":      types.StringType,
	}
}

type FullTextSearchModel struct {
	Language  types.String `tfsdk:"language"`
	Stemming  types.Bool   `tfsdk:"stemming"`
	StopWords types.Bool   `tfsdk:"stop_words"`
	Ngram     *NgramModel  `tfsdk:"ngram"`
}

func (m FullTextSearchModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"language":   types.StringType,
		"stemming":   types.BoolType,
		"stop_words": types.BoolType,
		"ngram":      types.ObjectType{AttrTypes: NgramModel{}.AttrTypes()},
	}
}

type NgramModel struct {
	MinGram    types.Int64 `tfsdk:"min_gram"`
	MaxGram    types.Int64 `tfsdk:"max_gram"`
	PrefixOnly types.Bool  `tfsdk:"prefix_only"`
}

func (m NgramModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"min_gram":    types.Int64Type,
		"max_gram":    types.Int64Type,
		"prefix_only": types.BoolType,
	}
}

// MetadataFieldModel is a string_list, boolean, float, or integer field, which the API adds to the
// schema as data is upserted.
type MetadataFieldModel struct {
	Filterable  types.Bool   `tfsdk:"filterable"`
	Description types.String `tfsdk:"description"`
}

func (m MetadataFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"filterable":  types.BoolType,
		"description": types.StringType,
	}
}

type LegacyMetadataFieldModel struct {
	Filterable types.Bool `tfsdk:"filterable"`
}

func (m LegacyMetadataFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"filterable": types.BoolType,
	}
}

func NewIndexSchemaObject(ctx context.Context, schema *pinecone.IndexSchema) (types.Object, diag.Diagnostics) {
	if schema == nil {
		return types.ObjectNull(IndexSchemaModel{}.AttrTypes()), nil
	}

	fields := make(map[string]IndexSchemaFieldModel, len(schema.Fields))
	for name, field := range schema.Fields {
		model, diags := newIndexSchemaFieldModel(ctx, field)
		if diags.HasError() {
			return types.ObjectNull(IndexSchemaModel{}.AttrTypes()), diags
		}
		fields[name] = model
	}

	fieldsMap, diags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: IndexSchemaFieldModel{}.AttrTypes()}, fields)
	if diags.HasError() {
		return types.ObjectNull(IndexSchemaModel{}.AttrTypes()), diags
	}
	return types.ObjectValueFrom(ctx, IndexSchemaModel{}.AttrTypes(), IndexSchemaModel{Fields: fieldsMap})
}

func newIndexSchemaFieldModel(ctx context.Context, field pinecone.IndexSchemaField) (IndexSchemaFieldModel, diag.Diagnostics) {
	var model IndexSchemaFieldModel
	switch {
	case field.DenseVector != nil:
		model.DenseVector = &DenseVectorFieldModel{
			Dimension:   types.Int32Value(field.DenseVector.Dimension),
			Metric:      types.StringValue(string(field.DenseVector.Metric)),
			Description: types.StringPointerValue(field.DenseVector.Description),
		}
	case field.SparseVector != nil:
		model.SparseVector = &SparseVectorFieldModel{
			Description: types.StringPointerValue(field.SparseVector.Description),
		}
	case field.SemanticText != nil:
		semantic := field.SemanticText
		readParameters, diags := stringMapValue(ctx, semantic.ReadParameters)
		if diags.HasError() {
			return model, diags
		}
		writeParameters, diags := stringMapValue(ctx, semantic.WriteParameters)
		if diags.HasError() {
			return model, diags
		}
		model.SemanticText = &SemanticTextFieldModel{
			Model:           types.StringValue(semantic.Model),
			Dimension:       types.Int32PointerValue(semantic.Dimension),
			Metric:          types.StringPointerValue((*string)(semantic.Metric)),
			ReadParameters:  readParameters,
			WriteParameters: writeParameters,
			Description:     types.StringPointerValue(semantic.Description),
		}
	case field.String != nil:
		model.String = &StringFieldModel{
			FullTextSearch: newFullTextSearchModel(field.String.FullTextSearch),
			Filterable:     types.BoolPointerValue(field.String.Filterable),
			Description:    types.StringPointerValue(field.String.Description),
		}
	case field.StringList != nil:
		model.StringList = &MetadataFieldModel{
			Filterable:  types.BoolPointerValue(field.StringList.Filterable),
			Description: types.StringPointerValue(field.StringList.Description),
		}
	case field.Boolean != nil:
		model.Boolean = &MetadataFieldModel{
			Filterable:  types.BoolPointerValue(field.Boolean.Filterable),
			Description: types.StringPointerValue(field.Boolean.Description),
		}
	case field.Float != nil:
		model.Float = &MetadataFieldModel{
			Filterable:  types.BoolPointerValue(field.Float.Filterable),
			Description: types.StringPointerValue(field.Float.Description),
		}
	case field.Integer != nil:
		model.Integer = &MetadataFieldModel{
			Filterable:  types.BoolPointerValue(field.Integer.Filterable),
			Description: types.StringPointerValue(field.Integer.Description),
		}
	case field.LegacyMetadata != nil:
		model.LegacyMetadata = &LegacyMetadataFieldModel{
			Filterable: types.BoolValue(field.LegacyMetadata.Filterable),
		}
	}
	return model, nil
}

func newFullTextSearchModel(config *pinecone.FullTextSearchConfig) *FullTextSearchModel {
	if config == nil {
		return nil
	}
	model := &FullTextSearchModel{
		Language:  types.StringPointerValue(config.Language),
		Stemming:  types.BoolPointerValue(config.Stemming),
		StopWords: types.BoolPointerValue(config.StopWords),
	}
	if config.Ngram != nil {
		model.Ngram = &NgramModel{
			MinGram:    types.Int64Value(int64(config.Ngram.MinGram)),
			MaxGram:    types.Int64Value(int64(config.Ngram.MaxGram)),
			PrefixOnly: types.BoolPointerValue(config.Ngram.PrefixOnly),
		}
	}
	return model
}

func stringMapValue(ctx context.Context, in *map[string]interface{}) (types.Map, diag.Diagnostics) {
	values, ok := toMapStringString(in)
	if !ok {
		return types.MapNull(types.StringType), nil
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}

// IndexDeploymentModel is where an index runs. Exactly one of its attributes is set.
type IndexDeploymentModel struct {
	Managed *ManagedDeploymentModel `tfsdk:"managed"`
	Pod     *PodDeploymentModel     `tfsdk:"pod"`
	Byoc    *ByocDeploymentModel    `tfsdk:"byoc"`
}

func (m IndexDeploymentModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"managed": types.ObjectType{AttrTypes: ManagedDeploymentModel{}.AttrTypes()},
		"pod":     types.ObjectType{AttrTypes: PodDeploymentModel{}.AttrTypes()},
		"byoc":    types.ObjectType{AttrTypes: ByocDeploymentModel{}.AttrTypes()},
	}
}

type ManagedDeploymentModel struct {
	Cloud       types.String `tfsdk:"cloud"`
	Region      types.String `tfsdk:"region"`
	Environment types.String `tfsdk:"environment"`
}

func (m ManagedDeploymentModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cloud":       types.StringType,
		"region":      types.StringType,
		"environment": types.StringType,
	}
}

type PodDeploymentModel struct {
	Environment types.String `tfsdk:"environment"`
	PodType     types.String `tfsdk:"pod_type"`
	Replicas    types.Int32  `tfsdk:"replicas"`
	Shards      types.Int32  `tfsdk:"shards"`
}

func (m PodDeploymentModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"environment": types.StringType,
		"pod_type":    types.StringType,
		"replicas":    types.Int32Type,
		"shards":      types.Int32Type,
	}
}

type ByocDeploymentModel struct {
	Environment types.String `tfsdk:"environment"`
}

func (m ByocDeploymentModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"environment": types.StringType,
	}
}

// NewIndexDeploymentObject converts an index deployment to its Terraform object. It's null when
// the index reports no deployment, including one of a type the SDK doesn't recognize.
func NewIndexDeploymentObject(ctx context.Context, deployment *pinecone.IndexDeployment) (types.Object, diag.Diagnostics) {
	if deployment == nil {
		return types.ObjectNull(IndexDeploymentModel{}.AttrTypes()), nil
	}

	var model IndexDeploymentModel
	switch {
	case deployment.Managed != nil:
		model.Managed = &ManagedDeploymentModel{
			Cloud:       types.StringValue(string(deployment.Managed.Cloud)),
			Region:      types.StringValue(deployment.Managed.Region),
			Environment: types.StringPointerValue(deployment.Managed.Environment),
		}
	case deployment.Pod != nil:
		model.Pod = &PodDeploymentModel{
			Environment: types.StringValue(deployment.Pod.Environment),
			PodType:     types.StringValue(deployment.Pod.PodType),
			Replicas:    types.Int32PointerValue(deployment.Pod.Replicas),
			Shards:      types.Int32PointerValue(deployment.Pod.Shards),
		}
	case deployment.Byoc != nil:
		model.Byoc = &ByocDeploymentModel{
			Environment: types.StringValue(deployment.Byoc.Environment),
		}
	}
	return types.ObjectValueFrom(ctx, IndexDeploymentModel{}.AttrTypes(), model)
}
