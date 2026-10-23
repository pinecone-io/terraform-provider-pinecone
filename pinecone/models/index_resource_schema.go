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

// StringResourceFieldModel is a string field searched by full-text search, integrated embedding,
// or both. Filterable is reported for metadata fields only, so the resource leaves it out.
type StringResourceFieldModel struct {
	FullTextSearch *FullTextSearchModel            `tfsdk:"full_text_search"`
	Embed          *EmbedConfigResourceModel       `tfsdk:"embed"`
	SparseEmbed    *SparseEmbedConfigResourceModel `tfsdk:"sparse_embed"`
	StoreText      types.Bool                      `tfsdk:"store_text"`
	Description    types.String                    `tfsdk:"description"`
}

func (m StringResourceFieldModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"full_text_search": types.ObjectType{AttrTypes: FullTextSearchModel{}.AttrTypes()},
		"embed":            types.ObjectType{AttrTypes: EmbedConfigResourceModel{}.AttrTypes()},
		"sparse_embed":     types.ObjectType{AttrTypes: SparseEmbedConfigResourceModel{}.AttrTypes()},
		"store_text":       types.BoolType,
		"description":      types.StringType,
	}
}

// EmbedConfigResourceModel is a string field's dense integrated embedding. ReadParameters and
// WriteParameters hold what was configured; the effective parameters hold the API's response,
// which merges them with the model's defaults.
type EmbedConfigResourceModel struct {
	Model                    types.String `tfsdk:"model"`
	Dimension                types.Int32  `tfsdk:"dimension"`
	Metric                   types.String `tfsdk:"metric"`
	ReadParameters           types.Map    `tfsdk:"read_parameters"`
	WriteParameters          types.Map    `tfsdk:"write_parameters"`
	EffectiveReadParameters  types.Map    `tfsdk:"effective_read_parameters"`
	EffectiveWriteParameters types.Map    `tfsdk:"effective_write_parameters"`
}

func (m EmbedConfigResourceModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"model":                      types.StringType,
		"dimension":                  types.Int32Type,
		"metric":                     types.StringType,
		"read_parameters":            types.MapType{ElemType: types.StringType},
		"write_parameters":           types.MapType{ElemType: types.StringType},
		"effective_read_parameters":  types.MapType{ElemType: types.StringType},
		"effective_write_parameters": types.MapType{ElemType: types.StringType},
	}
}

// SparseEmbedConfigResourceModel is a string field's sparse integrated embedding. Its parameters
// work as in EmbedConfigResourceModel.
type SparseEmbedConfigResourceModel struct {
	Model                    types.String `tfsdk:"model"`
	ReadParameters           types.Map    `tfsdk:"read_parameters"`
	WriteParameters          types.Map    `tfsdk:"write_parameters"`
	EffectiveReadParameters  types.Map    `tfsdk:"effective_read_parameters"`
	EffectiveWriteParameters types.Map    `tfsdk:"effective_write_parameters"`
}

func (m SparseEmbedConfigResourceModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"model":                      types.StringType,
		"read_parameters":            types.MapType{ElemType: types.StringType},
		"write_parameters":           types.MapType{ElemType: types.StringType},
		"effective_read_parameters":  types.MapType{ElemType: types.StringType},
		"effective_write_parameters": types.MapType{ElemType: types.StringType},
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
		if isDeclarableField(field) && !IsReservedFieldName(name) {
			return true
		}
	}
	return false
}

// ResourceSchemaFields decodes a schema object's fields. It returns nil for a null or unknown
// schema or field map. A nested object that isn't known yet, such as full_text_search before the
// API defaults it, decodes as nil.
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
	diags.Append(model.Fields.ElementsAs(ctx, &fields, true)...)
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
				Embed:          toEmbedConfig(field.String.Embed),
				SparseEmbed:    toSparseEmbedConfig(field.String.SparseEmbed),
				StoreText:      knownBoolPointer(field.String.StoreText),
				Description:    knownStringPointer(field.String.Description),
			}
		}
		schema.Fields[name] = sdkField
	}
	return schema, diags
}

// toFullTextSearchConfig leaves full-text search out when it isn't configured. The API enables it
// with its defaults on every field that stores its text, and rejects it with store_text = false.
func toFullTextSearchConfig(model *FullTextSearchModel) *pinecone.FullTextSearchConfig {
	if model == nil {
		return nil
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

func toEmbedConfig(model *EmbedConfigResourceModel) *pinecone.EmbedConfig {
	if model == nil {
		return nil
	}
	config := &pinecone.EmbedConfig{
		Model:           knownStringPointer(model.Model),
		ReadParameters:  mapAttrToInterfacePtr(model.ReadParameters),
		WriteParameters: mapAttrToInterfacePtr(model.WriteParameters),
	}
	if !model.Dimension.IsNull() && !model.Dimension.IsUnknown() {
		config.Dimension = model.Dimension.ValueInt32Pointer()
	}
	if metric := knownStringPointer(model.Metric); metric != nil {
		config.Metric = (*pinecone.IndexMetric)(metric)
	}
	return config
}

func toSparseEmbedConfig(model *SparseEmbedConfigResourceModel) *pinecone.SparseEmbedConfig {
	if model == nil {
		return nil
	}
	return &pinecone.SparseEmbedConfig{
		Model:           knownStringPointer(model.Model),
		ReadParameters:  mapAttrToInterfacePtr(model.ReadParameters),
		WriteParameters: mapAttrToInterfacePtr(model.WriteParameters),
	}
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
//
// The response reports integrated embedding parameters merged with the model's defaults, so they
// fill the effective parameters, and the configured parameters keep their prior values. On import,
// nothing was configured, so they're null.
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
			field, ok, d := newIndexResourceSchemaField(ctx, apiFields[name])
			diags.Append(d...)
			if !ok {
				fields[name] = priorField
				continue
			}
			keepConfiguredParameters(field.String, priorField.String)
			fields[name] = field
		}
	} else {
		for name, apiField := range apiFields {
			field, ok, d := newIndexResourceSchemaField(ctx, apiField)
			diags.Append(d...)
			if ok && !IsReservedFieldName(name) {
				fields[name] = field
			}
		}
	}
	if diags.HasError() {
		return types.ObjectNull(IndexResourceSchemaModel{}.AttrTypes()), diags
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

// isDeclarableField reports whether the resource can declare a reported field. String fields with
// neither full-text search nor integrated embedding are metadata fields, which aren't declared.
func isDeclarableField(field pinecone.IndexSchemaField) bool {
	if field.String != nil {
		return field.String.FullTextSearch != nil || field.String.Embed != nil || field.String.SparseEmbed != nil
	}
	return field.DenseVector != nil || field.SparseVector != nil
}

// newIndexResourceSchemaField converts a reported field to its resource model. It reports false for
// field types the resource can't declare.
func newIndexResourceSchemaField(ctx context.Context, field pinecone.IndexSchemaField) (IndexResourceSchemaFieldModel, bool, diag.Diagnostics) {
	switch {
	case field.DenseVector != nil:
		return IndexResourceSchemaFieldModel{DenseVector: &DenseVectorFieldModel{
			Dimension:   types.Int32Value(field.DenseVector.Dimension),
			Metric:      types.StringValue(string(field.DenseVector.Metric)),
			Description: types.StringPointerValue(field.DenseVector.Description),
		}}, true, nil
	case field.SparseVector != nil:
		return IndexResourceSchemaFieldModel{SparseVector: &SparseVectorFieldModel{
			Description: types.StringPointerValue(field.SparseVector.Description),
		}}, true, nil
	case field.String != nil && isDeclarableField(field):
		embed, diags := newEmbedConfigResourceModel(ctx, field.String.Embed)
		sparseEmbed, d := newSparseEmbedConfigResourceModel(ctx, field.String.SparseEmbed)
		diags.Append(d...)
		return IndexResourceSchemaFieldModel{String: &StringResourceFieldModel{
			FullTextSearch: newFullTextSearchModel(field.String.FullTextSearch),
			Embed:          embed,
			SparseEmbed:    sparseEmbed,
			StoreText:      types.BoolPointerValue(field.String.StoreText),
			Description:    types.StringPointerValue(field.String.Description),
		}}, true, diags
	}
	return IndexResourceSchemaFieldModel{}, false, nil
}

func newEmbedConfigResourceModel(ctx context.Context, config *pinecone.EmbedConfig) (*EmbedConfigResourceModel, diag.Diagnostics) {
	model, diags := newEmbedConfigModel(ctx, config)
	if model == nil || diags.HasError() {
		return nil, diags
	}
	return &EmbedConfigResourceModel{
		Model:                    model.Model,
		Dimension:                model.Dimension,
		Metric:                   model.Metric,
		ReadParameters:           types.MapNull(types.StringType),
		WriteParameters:          types.MapNull(types.StringType),
		EffectiveReadParameters:  model.ReadParameters,
		EffectiveWriteParameters: model.WriteParameters,
	}, nil
}

func newSparseEmbedConfigResourceModel(ctx context.Context, config *pinecone.SparseEmbedConfig) (*SparseEmbedConfigResourceModel, diag.Diagnostics) {
	model, diags := newSparseEmbedConfigModel(ctx, config)
	if model == nil || diags.HasError() {
		return nil, diags
	}
	return &SparseEmbedConfigResourceModel{
		Model:                    model.Model,
		ReadParameters:           types.MapNull(types.StringType),
		WriteParameters:          types.MapNull(types.StringType),
		EffectiveReadParameters:  model.ReadParameters,
		EffectiveWriteParameters: model.WriteParameters,
	}, nil
}

// AdoptsConfiguredParameters reports whether planned differs from prior only by integrated
// embedding parameters that prior doesn't record as configured, as after an import, and that the
// index already uses: each configured key has the same value in the effective parameters. Taking
// them into state then describes the same index, so it doesn't need to be replaced.
func AdoptsConfiguredParameters(ctx context.Context, planned, prior types.Object) bool {
	plannedFields, diags := ResourceSchemaFields(ctx, planned)
	priorFields, d := ResourceSchemaFields(ctx, prior)
	diags.Append(d...)
	if diags.HasError() || plannedFields == nil || priorFields == nil {
		return false
	}

	adopted := false
	for name, field := range plannedFields {
		priorString := priorFields[name].String
		if field.String == nil || priorString == nil {
			continue
		}
		adoptedString := *field.String
		if field.String.Embed != nil && priorString.Embed != nil {
			embed := *field.String.Embed
			embed.ReadParameters, adopted = adoptParameters(embed.ReadParameters, priorString.Embed.ReadParameters, priorString.Embed.EffectiveReadParameters, adopted)
			embed.WriteParameters, adopted = adoptParameters(embed.WriteParameters, priorString.Embed.WriteParameters, priorString.Embed.EffectiveWriteParameters, adopted)
			adoptedString.Embed = &embed
		}
		if field.String.SparseEmbed != nil && priorString.SparseEmbed != nil {
			sparseEmbed := *field.String.SparseEmbed
			sparseEmbed.ReadParameters, adopted = adoptParameters(sparseEmbed.ReadParameters, priorString.SparseEmbed.ReadParameters, priorString.SparseEmbed.EffectiveReadParameters, adopted)
			sparseEmbed.WriteParameters, adopted = adoptParameters(sparseEmbed.WriteParameters, priorString.SparseEmbed.WriteParameters, priorString.SparseEmbed.EffectiveWriteParameters, adopted)
			adoptedString.SparseEmbed = &sparseEmbed
		}
		field.String = &adoptedString
		plannedFields[name] = field
	}
	if !adopted {
		return false
	}

	fieldType := types.ObjectType{AttrTypes: IndexResourceSchemaFieldModel{}.AttrTypes()}
	plannedMap, d := types.MapValueFrom(ctx, fieldType, plannedFields)
	diags.Append(d...)
	priorMap, d := types.MapValueFrom(ctx, fieldType, priorFields)
	diags.Append(d...)
	return !diags.HasError() && plannedMap.Equal(priorMap)
}

// adoptParameters returns prior in place of planned when prior is null and planned is a known
// subset of effective, and reports whether any parameters were adopted.
func adoptParameters(planned, prior, effective types.Map, adopted bool) (types.Map, bool) {
	if !prior.IsNull() || planned.IsNull() || planned.IsUnknown() || effective.IsNull() || effective.IsUnknown() {
		return planned, adopted
	}
	effectiveElements := effective.Elements()
	for key, value := range planned.Elements() {
		if current, ok := effectiveElements[key]; !ok || !value.Equal(current) {
			return planned, adopted
		}
	}
	return prior, true
}

// keepConfiguredParameters copies the configured integrated embedding parameters from prior to
// field.
func keepConfiguredParameters(field, prior *StringResourceFieldModel) {
	if field == nil || prior == nil {
		return
	}
	if field.Embed != nil && prior.Embed != nil {
		field.Embed.ReadParameters, field.Embed.WriteParameters = prior.Embed.ReadParameters, prior.Embed.WriteParameters
	}
	if field.SparseEmbed != nil && prior.SparseEmbed != nil {
		field.SparseEmbed.ReadParameters, field.SparseEmbed.WriteParameters = prior.SparseEmbed.ReadParameters, prior.SparseEmbed.WriteParameters
	}
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
