// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
)

const (
	ReservedDenseFieldName  = "_values"
	ReservedSparseFieldName = "_sparse_values"
)

func IsReservedFieldName(name string) bool {
	return name == ReservedDenseFieldName || name == ReservedSparseFieldName
}

func HasInvalidFieldNamePrefix(name string) bool {
	return strings.HasPrefix(name, "$") || (strings.HasPrefix(name, "_") && !IsReservedFieldName(name))
}

// IndexResourceSchemaModel is the resource's schema: only the field types an index can be created
// with.
type IndexResourceSchemaModel struct {
	Fields types.Map `tfsdk:"fields"`
}

func (m IndexResourceSchemaModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"fields": types.MapType{ElemType: types.ObjectType{AttrTypes: IndexResourceSchemaFieldModel{}.AttrTypes()}},
	}
}

type IndexResourceSchemaFieldModel struct {
	DenseVector  *DenseVectorFieldModel    `tfsdk:"dense_vector"`
	SparseVector *SparseVectorFieldModel   `tfsdk:"sparse_vector"`
	String       *StringResourceFieldModel `tfsdk:"string"`
}

func (m IndexResourceSchemaFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dense_vector":  types.ObjectType{AttrTypes: DenseVectorFieldModel{}.AttrTypes()},
		"sparse_vector": types.ObjectType{AttrTypes: SparseVectorFieldModel{}.AttrTypes()},
		"string":        types.ObjectType{AttrTypes: StringResourceFieldModel{}.AttrTypes()},
	}
}

// StringResourceFieldModel is a full-text-search string field. Filterable is reported for
// metadata fields only, so the resource leaves it out.
type StringResourceFieldModel struct {
	FullTextSearch *FullTextSearchModel `tfsdk:"full_text_search"`
	Description    types.String         `tfsdk:"description"`
}

func (m StringResourceFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"full_text_search": types.ObjectType{AttrTypes: FullTextSearchModel{}.AttrTypes()},
		"description":      types.StringType,
	}
}

// IndexResourceDeploymentModel is where a schema-mode index runs. Pod-based indexes can't be
// created on API version 2026-07, so it has no pod deployment.
type IndexResourceDeploymentModel struct {
	Managed *ManagedDeploymentModel `tfsdk:"managed"`
	Byoc    *ByocDeploymentModel    `tfsdk:"byoc"`
}

func (m IndexResourceDeploymentModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"managed": types.ObjectType{AttrTypes: ManagedDeploymentModel{}.AttrTypes()},
		"byoc":    types.ObjectType{AttrTypes: ByocDeploymentModel{}.AttrTypes()},
	}
}

// IsDocumentIndexSchema reports whether an index schema has a field the resource can declare
// other than the reserved vector fields, which makes it a document index.
func IsDocumentIndexSchema(schema *pinecone.IndexSchema) bool {
	if schema == nil {
		return false
	}
	for name, field := range schema.Fields {
		if _, ok := newIndexResourceSchemaField(field); ok && !IsReservedFieldName(name) {
			return true
		}
	}
	return false
}

func IsDocumentIndexConfig(fields map[string]IndexResourceSchemaFieldModel) bool {
	for name := range fields {
		if !IsReservedFieldName(name) {
			return true
		}
	}
	return false
}

// ResourceSchemaFields decodes a schema object's fields. It returns nil for a null or unknown
// schema or field map.
func ResourceSchemaFields(ctx context.Context, obj types.Object) (map[string]IndexResourceSchemaFieldModel, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var model IndexResourceSchemaModel
	diags := obj.As(ctx, &model, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})
	if diags.HasError() || model.Fields.IsNull() || model.Fields.IsUnknown() {
		return nil, diags
	}
	var fields map[string]IndexResourceSchemaFieldModel
	diags.Append(model.Fields.ElementsAs(ctx, &fields, false)...)
	return fields, diags
}

// ToIndexSchema converts a configured schema to an SDK schema for CreateIndex. Unknown values,
// which Optional+Computed attributes plan when they aren't configured, are left for the API to
// default.
func ToIndexSchema(ctx context.Context, obj types.Object) (pinecone.IndexSchema, diag.Diagnostics) {
	fields, diags := ResourceSchemaFields(ctx, obj)
	if diags.HasError() {
		return pinecone.IndexSchema{}, diags
	}

	schema := pinecone.IndexSchema{Fields: make(map[string]pinecone.IndexSchemaField, len(fields))}
	for name, field := range fields {
		var sdkField pinecone.IndexSchemaField
		switch {
		case field.DenseVector != nil:
			sdkField.DenseVector = &pinecone.DenseVectorField{
				Dimension:   field.DenseVector.Dimension.ValueInt32(),
				Metric:      pinecone.IndexMetric(field.DenseVector.Metric.ValueString()),
				Description: knownStringPointer(field.DenseVector.Description),
			}
		case field.SparseVector != nil:
			sdkField.SparseVector = &pinecone.SparseVectorField{
				Description: knownStringPointer(field.SparseVector.Description),
			}
		case field.String != nil:
			sdkField.String = &pinecone.StringField{
				FullTextSearch: toFullTextSearchConfig(field.String.FullTextSearch),
				Description:    knownStringPointer(field.String.Description),
			}
		}
		schema.Fields[name] = sdkField
	}
	return schema, diags
}

func toFullTextSearchConfig(model *FullTextSearchModel) *pinecone.FullTextSearchConfig {
	if model == nil {
		return &pinecone.FullTextSearchConfig{}
	}
	config := &pinecone.FullTextSearchConfig{
		Language:  knownStringPointer(model.Language),
		Stemming:  knownBoolPointer(model.Stemming),
		StopWords: knownBoolPointer(model.StopWords),
	}
	if model.Ngram != nil {
		config.Ngram = &pinecone.NgramConfig{
			MinGram:    int(model.Ngram.MinGram.ValueInt64()),
			MaxGram:    int(model.Ngram.MaxGram.ValueInt64()),
			PrefixOnly: knownBoolPointer(model.Ngram.PrefixOnly),
		}
	}
	return config
}

func ToIndexDeployment(ctx context.Context, obj types.Object) (*pinecone.IndexDeployment, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var model IndexResourceDeploymentModel
	diags := obj.As(ctx, &model, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})
	if diags.HasError() {
		return nil, diags
	}
	switch {
	case model.Managed != nil:
		return &pinecone.IndexDeployment{Managed: &pinecone.ManagedDeployment{
			Cloud:  pinecone.Cloud(model.Managed.Cloud.ValueString()),
			Region: model.Managed.Region.ValueString(),
		}}, diags
	case model.Byoc != nil:
		return &pinecone.IndexDeployment{Byoc: &pinecone.ByocDeployment{
			Environment: model.Byoc.Environment.ValueString(),
		}}, diags
	}
	return nil, diags
}

// NewIndexResourceSchemaObject builds the resource's schema state from the index's schema. It keeps
// only the fields in prior, from state or the plan: the API also reports the reserved
// _sparse_values field on every index with dense vectors and the metadata fields it adds as data
// is upserted, which were never declared. A declared field the response doesn't describe keeps its
// prior value, since a schema can't change after the index is created. Without a prior schema, as
// on import, every field the resource can declare is kept.
func NewIndexResourceSchemaObject(ctx context.Context, schema *pinecone.IndexSchema, prior types.Object) (types.Object, diag.Diagnostics) {
	priorFields, diags := ResourceSchemaFields(ctx, prior)
	if diags.HasError() {
		return types.ObjectNull(IndexResourceSchemaModel{}.AttrTypes()), diags
	}

	var apiFields map[string]pinecone.IndexSchemaField
	if schema != nil {
		apiFields = schema.Fields
	}

	fields := make(map[string]IndexResourceSchemaFieldModel)
	if priorFields != nil {
		for name, priorField := range priorFields {
			if field, ok := newIndexResourceSchemaField(apiFields[name]); ok {
				fields[name] = field
			} else {
				fields[name] = priorField
			}
		}
	} else {
		for name, apiField := range apiFields {
			if field, ok := newIndexResourceSchemaField(apiField); ok && !IsReservedFieldName(name) {
				fields[name] = field
			}
		}
	}

	fieldsMap, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: IndexResourceSchemaFieldModel{}.AttrTypes()}, fields)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(IndexResourceSchemaModel{}.AttrTypes()), diags
	}
	obj, d := types.ObjectValueFrom(ctx, IndexResourceSchemaModel{}.AttrTypes(), IndexResourceSchemaModel{Fields: fieldsMap})
	diags.Append(d...)
	return obj, diags
}

// newIndexResourceSchemaField converts a reported field to its resource model. It reports false for
// field types the resource can't declare, including string fields without full-text search.
func newIndexResourceSchemaField(field pinecone.IndexSchemaField) (IndexResourceSchemaFieldModel, bool) {
	switch {
	case field.DenseVector != nil:
		return IndexResourceSchemaFieldModel{DenseVector: &DenseVectorFieldModel{
			Dimension:   types.Int32Value(field.DenseVector.Dimension),
			Metric:      types.StringValue(string(field.DenseVector.Metric)),
			Description: types.StringPointerValue(field.DenseVector.Description),
		}}, true
	case field.SparseVector != nil:
		return IndexResourceSchemaFieldModel{SparseVector: &SparseVectorFieldModel{
			Description: types.StringPointerValue(field.SparseVector.Description),
		}}, true
	case field.String != nil && field.String.FullTextSearch != nil:
		return IndexResourceSchemaFieldModel{String: &StringResourceFieldModel{
			FullTextSearch: newFullTextSearchModel(field.String.FullTextSearch),
			Description:    types.StringPointerValue(field.String.Description),
		}}, true
	}
	return IndexResourceSchemaFieldModel{}, false
}

// NewIndexResourceDeploymentObject builds the resource's deployment state from the index's
// deployment. It keeps prior when the response has no managed or BYOC deployment, so a deployment
// the SDK can't decode doesn't plan a replacement.
func NewIndexResourceDeploymentObject(ctx context.Context, deployment *pinecone.IndexDeployment, prior types.Object) (types.Object, diag.Diagnostics) {
	var model IndexResourceDeploymentModel
	switch {
	case deployment != nil && deployment.Managed != nil:
		model.Managed = &ManagedDeploymentModel{
			Cloud:       types.StringValue(string(deployment.Managed.Cloud)),
			Region:      types.StringValue(deployment.Managed.Region),
			Environment: types.StringPointerValue(deployment.Managed.Environment),
		}
	case deployment != nil && deployment.Byoc != nil:
		model.Byoc = &ByocDeploymentModel{Environment: types.StringValue(deployment.Byoc.Environment)}
	case !prior.IsUnknown():
		return prior, nil
	default:
		return types.ObjectNull(IndexResourceDeploymentModel{}.AttrTypes()), nil
	}
	return types.ObjectValueFrom(ctx, IndexResourceDeploymentModel{}.AttrTypes(), model)
}

func knownStringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
}

func knownBoolPointer(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueBoolPointer()
}
