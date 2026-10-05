// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
)

func resourceSchemaObject(t *testing.T, fields map[string]IndexResourceSchemaFieldModel) types.Object {
	t.Helper()
	ctx := t.Context()
	fieldsMap, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: IndexResourceSchemaFieldModel{}.AttrTypes()}, fields)
	if d.HasError() {
		t.Fatalf("building fields: %v", d)
	}
	obj, d := types.ObjectValueFrom(ctx, IndexResourceSchemaModel{}.AttrTypes(), IndexResourceSchemaModel{Fields: fieldsMap})
	if d.HasError() {
		t.Fatalf("building schema: %v", d)
	}
	return obj
}

func sortedFieldNames(t *testing.T, obj types.Object) []string {
	t.Helper()
	fields, d := ResourceSchemaFields(t.Context(), obj)
	if d.HasError() {
		t.Fatalf("decoding schema: %v", d)
	}
	var names []string
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func TestToIndexSchema(t *testing.T) {
	obj := resourceSchemaObject(t, map[string]IndexResourceSchemaFieldModel{
		"embedding": {DenseVector: &DenseVectorFieldModel{Dimension: types.Int32Value(1536), Metric: types.StringValue("dotproduct"), Description: types.StringUnknown()}},
		"terms":     {SparseVector: &SparseVectorFieldModel{Description: types.StringValue("learned sparse")}},
		"body": {String: &StringResourceFieldModel{FullTextSearch: &FullTextSearchModel{
			Language: types.StringUnknown(), Stemming: types.BoolValue(true), StopWords: types.BoolUnknown(),
		}}},
		"title": {String: &StringResourceFieldModel{FullTextSearch: &FullTextSearchModel{
			Ngram: &NgramModel{MinGram: types.Int64Value(2), MaxGram: types.Int64Value(5), PrefixOnly: types.BoolUnknown()},
		}}},
	})

	schema, d := ToIndexSchema(t.Context(), obj)
	if d.HasError() {
		t.Fatalf("ToIndexSchema returned errors: %v", d)
	}

	dense := schema.Fields["embedding"].DenseVector
	if dense == nil || dense.Dimension != 1536 || dense.Metric != pinecone.IndexMetricDotproduct || dense.Description != nil {
		t.Errorf("embedding = %+v; an unknown description must be left out", dense)
	}
	if sparse := schema.Fields["terms"].SparseVector; sparse == nil || sparse.Description == nil || *sparse.Description != "learned sparse" {
		t.Errorf("terms = %+v", sparse)
	}
	body := schema.Fields["body"].String
	if body == nil || body.FullTextSearch == nil || body.FullTextSearch.Language != nil || body.FullTextSearch.StopWords != nil ||
		body.FullTextSearch.Stemming == nil || !*body.FullTextSearch.Stemming {
		t.Errorf("body = %+v; unknown values must be left for the API to default", body)
	}
	title := schema.Fields["title"].String
	if title == nil || title.FullTextSearch.Ngram == nil || title.FullTextSearch.Ngram.MinGram != 2 || title.FullTextSearch.Ngram.MaxGram != 5 ||
		title.FullTextSearch.Ngram.PrefixOnly != nil {
		t.Errorf("title = %+v", title)
	}
}

func TestToIndexDeployment(t *testing.T) {
	ctx := t.Context()
	managed, d := types.ObjectValueFrom(ctx, IndexResourceDeploymentModel{}.AttrTypes(), IndexResourceDeploymentModel{
		Managed: &ManagedDeploymentModel{Cloud: types.StringValue("gcp"), Region: types.StringValue("us-central1"), Environment: types.StringUnknown()},
	})
	if d.HasError() {
		t.Fatalf("building deployment: %v", d)
	}
	deployment, d := ToIndexDeployment(ctx, managed)
	if d.HasError() || deployment == nil || deployment.Managed == nil || deployment.Managed.Cloud != pinecone.CloudGCP ||
		deployment.Managed.Region != "us-central1" || deployment.Managed.Environment != nil {
		t.Errorf("ToIndexDeployment = %+v (%v)", deployment, d)
	}
}

func documentIndexSchema() *pinecone.IndexSchema {
	filterable := true
	language := "en"
	return &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
		"embedding": {DenseVector: &pinecone.DenseVectorField{Dimension: 1536, Metric: pinecone.IndexMetricDotproduct}},
		"body":      {String: &pinecone.StringField{FullTextSearch: &pinecone.FullTextSearchConfig{Language: &language}}},
		"author":    {String: &pinecone.StringField{Filterable: &filterable}},
		"year":      {Float: &pinecone.FloatField{Filterable: &filterable}},
	}}
}

func TestNewIndexResourceSchemaObject(t *testing.T) {
	ctx := t.Context()
	vectorSchema := &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
		"_values":        {DenseVector: &pinecone.DenseVectorField{Dimension: 1024, Metric: pinecone.IndexMetricCosine}},
		"_sparse_values": {SparseVector: &pinecone.SparseVectorField{}},
		"genre":          {LegacyMetadata: &pinecone.LegacyMetadataField{Filterable: true}},
	}}

	t.Run("keeps only declared fields", func(t *testing.T) {
		prior := resourceSchemaObject(t, map[string]IndexResourceSchemaFieldModel{
			"_values": {DenseVector: &DenseVectorFieldModel{Dimension: types.Int32Value(1024), Metric: types.StringValue("cosine")}},
		})
		obj, d := NewIndexResourceSchemaObject(ctx, vectorSchema, prior)
		if d.HasError() {
			t.Fatalf("returned errors: %v", d)
		}
		if got := sortedFieldNames(t, obj); !slices.Equal(got, []string{"_values"}) {
			t.Errorf("fields = %v, want [_values]: _sparse_values and metadata fields weren't declared", got)
		}
	})

	t.Run("fills server defaults of declared fields", func(t *testing.T) {
		prior := resourceSchemaObject(t, map[string]IndexResourceSchemaFieldModel{
			"body": {String: &StringResourceFieldModel{FullTextSearch: &FullTextSearchModel{Language: types.StringUnknown()}}},
		})
		obj, d := NewIndexResourceSchemaObject(ctx, documentIndexSchema(), prior)
		if d.HasError() {
			t.Fatalf("returned errors: %v", d)
		}
		fields, _ := ResourceSchemaFields(ctx, obj)
		if got := fields["body"].String.FullTextSearch.Language; got.ValueString() != "en" {
			t.Errorf("body language = %v, want en from the response", got)
		}
	})

	t.Run("keeps a declared field the response omits", func(t *testing.T) {
		prior := resourceSchemaObject(t, map[string]IndexResourceSchemaFieldModel{
			"missing": {SparseVector: &SparseVectorFieldModel{Description: types.StringValue("kept")}},
		})
		obj, d := NewIndexResourceSchemaObject(ctx, documentIndexSchema(), prior)
		if d.HasError() {
			t.Fatalf("returned errors: %v", d)
		}
		fields, _ := ResourceSchemaFields(ctx, obj)
		if f := fields["missing"].SparseVector; f == nil || f.Description.ValueString() != "kept" {
			t.Errorf("missing = %+v, want the prior value", fields["missing"])
		}
	})

	t.Run("import keeps declarable named fields", func(t *testing.T) {
		obj, d := NewIndexResourceSchemaObject(ctx, documentIndexSchema(), types.ObjectNull(IndexResourceSchemaModel{}.AttrTypes()))
		if d.HasError() {
			t.Fatalf("returned errors: %v", d)
		}
		if got := sortedFieldNames(t, obj); !slices.Equal(got, []string{"body", "embedding"}) {
			t.Errorf("fields = %v, want [body embedding]: author has no full-text search and year is metadata", got)
		}
	})
}

func TestIsDocumentIndexSchema(t *testing.T) {
	if !IsDocumentIndexSchema(documentIndexSchema()) {
		t.Error("a schema with named vector and full-text search fields is a document index")
	}
	vectorSchema := &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
		"_values": {DenseVector: &pinecone.DenseVectorField{Dimension: 1024, Metric: pinecone.IndexMetricCosine}},
		"genre":   {LegacyMetadata: &pinecone.LegacyMetadataField{Filterable: true}},
	}}
	if IsDocumentIndexSchema(vectorSchema) {
		t.Error("reserved vector fields with metadata make a vector index")
	}
	integrated := &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
		"chunk_text": {SemanticText: &pinecone.SemanticTextField{Model: "multilingual-e5-large"}},
	}}
	if IsDocumentIndexSchema(integrated) {
		t.Error("an integrated index's semantic_text field can't be declared, so it isn't a document index")
	}
}

func TestIndexResourceModelRead_schemaMode(t *testing.T) {
	ctx := t.Context()
	environment := "aped-4627-b74a"
	index := &pinecone.Index{
		Name:       "articles",
		Host:       "https://articles.example.com",
		Schema:     documentIndexSchema(),
		Deployment: &pinecone.IndexDeployment{Managed: &pinecone.ManagedDeployment{Cloud: pinecone.CloudAWS, Region: "us-east-1", Environment: &environment}},
		ReadCapacity: &pinecone.ReadCapacity{OnDemand: &pinecone.ReadCapacityOnDemand{
			Status: pinecone.ReadCapacityStatus{State: "Ready"},
		}},
		Metric:     pinecone.IndexMetricDotproduct,
		VectorType: "dense",
		Spec:       &pinecone.IndexSpec{Serverless: &pinecone.ServerlessSpec{Cloud: pinecone.CloudAWS, Region: "us-east-1"}},
		Status:     &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady},
	}

	var model IndexResourceModel
	if d := model.Read(ctx, index); d.HasError() {
		t.Fatalf("Read returned errors: %v", d)
	}
	if !model.UsesSchemaMode(index) || model.Schema.IsNull() {
		t.Fatal("an imported document index should be read in schema mode")
	}
	if !model.Spec.IsNull() || !model.Dimension.IsNull() || !model.Metric.IsNull() || !model.VectorType.IsNull() || !model.Embed.IsNull() {
		t.Errorf("spec-based attributes must be null in schema mode: spec=%v dimension=%v metric=%v vector_type=%v embed=%v",
			model.Spec, model.Dimension, model.Metric, model.VectorType, model.Embed)
	}
	if model.Deployment.IsNull() || model.ReadCapacity.IsNull() {
		t.Errorf("deployment = %v, read_capacity = %v; want both set", model.Deployment, model.ReadCapacity)
	}
	if got := sortedFieldNames(t, model.Schema); !slices.Equal(got, []string{"body", "embedding"}) {
		t.Errorf("fields = %v, want [body embedding]", got)
	}

	if d := model.Read(ctx, index); d.HasError() || model.Schema.IsNull() {
		t.Errorf("refresh should stay in schema mode (%v)", d)
	}
}

func TestIndexResourceModelRead_importVectorIndexUsesSpec(t *testing.T) {
	dimension := int32(1024)
	index := &pinecone.Index{
		Name:      "products",
		Metric:    pinecone.IndexMetricCosine,
		Dimension: &dimension,
		Schema: &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
			"_values":        {DenseVector: &pinecone.DenseVectorField{Dimension: dimension, Metric: pinecone.IndexMetricCosine}},
			"_sparse_values": {SparseVector: &pinecone.SparseVectorField{}},
		}},
		Deployment: &pinecone.IndexDeployment{Managed: &pinecone.ManagedDeployment{Cloud: pinecone.CloudAWS, Region: "us-east-1"}},
		Spec:       &pinecone.IndexSpec{Serverless: &pinecone.ServerlessSpec{Cloud: pinecone.CloudAWS, Region: "us-east-1"}},
	}

	var model IndexResourceModel
	if d := model.Read(t.Context(), index); d.HasError() {
		t.Fatalf("Read returned errors: %v", d)
	}
	if !model.Schema.IsNull() || !model.Deployment.IsNull() || model.Spec.IsNull() || model.Dimension.ValueInt32() != 1024 {
		t.Errorf("a vector index imports with spec: schema=%v deployment=%v spec=%v dimension=%v",
			model.Schema, model.Deployment, model.Spec, model.Dimension)
	}
}
