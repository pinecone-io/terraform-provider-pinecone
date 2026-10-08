// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
)

func ptr[T any](v T) *T { return &v }

func schemaFields(t *testing.T, obj types.Object) map[string]IndexSchemaFieldModel {
	t.Helper()
	ctx := t.Context()
	var schema IndexSchemaModel
	if d := obj.As(ctx, &schema, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding schema: %v", d)
	}
	var fields map[string]IndexSchemaFieldModel
	if d := schema.Fields.ElementsAs(ctx, &fields, false); d.HasError() {
		t.Fatalf("decoding fields: %v", d)
	}
	return fields
}

func TestNewIndexSchemaObject(t *testing.T) {
	metric := pinecone.IndexMetricCosine
	schema := &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
		"embedding": {DenseVector: &pinecone.DenseVectorField{Dimension: 1536, Metric: pinecone.IndexMetricDotproduct, Description: ptr("vectors")}},
		"terms":     {SparseVector: &pinecone.SparseVectorField{}},
		"chunk_text": {SemanticText: &pinecone.SemanticTextField{
			Model:          "multilingual-e5-large",
			Dimension:      ptr(int32(1024)),
			Metric:         &metric,
			ReadParameters: &map[string]interface{}{"input_type": "query", "truncate": "END"},
		}},
		"body": {String: &pinecone.StringField{FullTextSearch: &pinecone.FullTextSearchConfig{
			Language: ptr("en"),
			Stemming: ptr(true),
		}}},
		"title": {String: &pinecone.StringField{FullTextSearch: &pinecone.FullTextSearchConfig{
			Ngram: &pinecone.NgramConfig{MinGram: 2, MaxGram: 5, PrefixOnly: ptr(true)},
		}}},
		"author": {String: &pinecone.StringField{Filterable: ptr(true)}},
		"tags":   {StringList: &pinecone.StringListField{Filterable: ptr(true)}},
		"active": {Boolean: &pinecone.BooleanField{Filterable: ptr(false)}},
		"score":  {Float: &pinecone.FloatField{Filterable: ptr(true)}},
		"year":   {Integer: &pinecone.IntegerField{Filterable: ptr(true)}},
		"genre":  {LegacyMetadata: &pinecone.LegacyMetadataField{Filterable: true}},
	}}

	obj, diags := NewIndexSchemaObject(t.Context(), schema)
	if diags.HasError() {
		t.Fatalf("NewIndexSchemaObject returned errors: %v", diags)
	}
	fields := schemaFields(t, obj)
	if len(fields) != len(schema.Fields) {
		t.Fatalf("got %d fields, want %d", len(fields), len(schema.Fields))
	}

	if f := fields["embedding"].DenseVector; f == nil || f.Dimension.ValueInt32() != 1536 || f.Metric.ValueString() != "dotproduct" || f.Description.ValueString() != "vectors" {
		t.Errorf("embedding = %+v", f)
	}
	if f := fields["terms"].SparseVector; f == nil || !f.Description.IsNull() {
		t.Errorf("terms = %+v", f)
	}
	if f := fields["chunk_text"].SemanticText; f == nil || f.Model.ValueString() != "multilingual-e5-large" ||
		f.Dimension.ValueInt32() != 1024 || f.Metric.ValueString() != "cosine" || len(f.ReadParameters.Elements()) != 2 || !f.WriteParameters.IsNull() {
		t.Errorf("chunk_text = %+v", f)
	}
	if f := fields["body"].String; f == nil || f.FullTextSearch == nil || f.FullTextSearch.Language.ValueString() != "en" ||
		!f.FullTextSearch.Stemming.ValueBool() || !f.FullTextSearch.StopWords.IsNull() || f.FullTextSearch.Ngram != nil {
		t.Errorf("body = %+v", f)
	}
	if f := fields["title"].String; f == nil || f.FullTextSearch == nil || f.FullTextSearch.Ngram == nil ||
		f.FullTextSearch.Ngram.MinGram.ValueInt64() != 2 || f.FullTextSearch.Ngram.MaxGram.ValueInt64() != 5 || !f.FullTextSearch.Ngram.PrefixOnly.ValueBool() {
		t.Errorf("title = %+v", f)
	}
	if f := fields["author"].String; f == nil || f.FullTextSearch != nil || !f.Filterable.ValueBool() {
		t.Errorf("author = %+v", f)
	}
	if f := fields["tags"].StringList; f == nil || !f.Filterable.ValueBool() {
		t.Errorf("tags = %+v", f)
	}
	if f := fields["active"].Boolean; f == nil || f.Filterable.ValueBool() {
		t.Errorf("active = %+v", f)
	}
	if f := fields["score"].Float; f == nil || !f.Filterable.ValueBool() {
		t.Errorf("score = %+v", f)
	}
	if f := fields["year"].Integer; f == nil || !f.Filterable.ValueBool() {
		t.Errorf("year = %+v", f)
	}
	if f := fields["genre"].LegacyMetadata; f == nil || !f.Filterable.ValueBool() {
		t.Errorf("genre = %+v", f)
	}

	for name, field := range fields {
		set := 0
		for _, isSet := range []bool{field.DenseVector != nil, field.SparseVector != nil, field.SemanticText != nil, field.String != nil,
			field.StringList != nil, field.Boolean != nil, field.Float != nil, field.Integer != nil, field.LegacyMetadata != nil} {
			if isSet {
				set++
			}
		}
		if set != 1 {
			t.Errorf("field %q has %d types set, want 1", name, set)
		}
	}
}

func TestNewIndexSchemaObject_nil(t *testing.T) {
	obj, diags := NewIndexSchemaObject(t.Context(), nil)
	if diags.HasError() || !obj.IsNull() {
		t.Errorf("NewIndexSchemaObject(nil) = %v, %v; want null object", obj, diags)
	}
}

func TestNewIndexDeploymentObject(t *testing.T) {
	tests := []struct {
		name       string
		deployment *pinecone.IndexDeployment
		check      func(t *testing.T, m IndexDeploymentModel)
	}{
		{"managed", &pinecone.IndexDeployment{Managed: &pinecone.ManagedDeployment{Cloud: pinecone.CloudAWS, Region: "us-east-1", Environment: ptr("aped-4627-b74a")}},
			func(t *testing.T, m IndexDeploymentModel) {
				if m.Managed == nil || m.Managed.Cloud.ValueString() != "aws" || m.Managed.Region.ValueString() != "us-east-1" ||
					m.Managed.Environment.ValueString() != "aped-4627-b74a" || m.Pod != nil || m.Byoc != nil {
					t.Errorf("deployment = %+v", m)
				}
			}},
		{"pod", &pinecone.IndexDeployment{Pod: &pinecone.PodDeployment{Environment: "us-west4-gcp", PodType: "p1.x2", Replicas: ptr(int32(2))}},
			func(t *testing.T, m IndexDeploymentModel) {
				if m.Pod == nil || m.Pod.PodType.ValueString() != "p1.x2" || m.Pod.Replicas.ValueInt32() != 2 || !m.Pod.Shards.IsNull() || m.Managed != nil {
					t.Errorf("deployment = %+v", m)
				}
			}},
		{"byoc", &pinecone.IndexDeployment{Byoc: &pinecone.ByocDeployment{Environment: "aws-us-east-1-b921"}},
			func(t *testing.T, m IndexDeploymentModel) {
				if m.Byoc == nil || m.Byoc.Environment.ValueString() != "aws-us-east-1-b921" || m.Managed != nil {
					t.Errorf("deployment = %+v", m)
				}
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, diags := NewIndexDeploymentObject(t.Context(), tt.deployment)
			if diags.HasError() {
				t.Fatalf("NewIndexDeploymentObject returned errors: %v", diags)
			}
			var m IndexDeploymentModel
			if d := obj.As(t.Context(), &m, basetypes.ObjectAsOptions{}); d.HasError() {
				t.Fatalf("decoding deployment: %v", d)
			}
			tt.check(t, m)
		})
	}

	if obj, diags := NewIndexDeploymentObject(t.Context(), nil); diags.HasError() || !obj.IsNull() {
		t.Errorf("NewIndexDeploymentObject(nil) = %v, %v; want null object", obj, diags)
	}
	if obj, diags := NewIndexDeploymentObject(t.Context(), &pinecone.IndexDeployment{}); diags.HasError() || !obj.IsNull() {
		t.Errorf("NewIndexDeploymentObject(empty) = %v, %v; want null object", obj, diags)
	}
}

func TestIndexModelRead(t *testing.T) {
	nodeType := "b1"
	index := &pinecone.Index{
		Name:               "my-index",
		DeletionProtection: pinecone.DeletionProtectionEnabled,
		PrivateHost:        ptr("my-index.private.example.com"),
		CmekId:             ptr("cmek-123"),
		SourceBackupId:     ptr("backup-456"),
		ReadCapacity: &pinecone.ReadCapacity{Dedicated: &pinecone.ReadCapacityDedicated{
			NodeType: &nodeType,
			Scaling:  &pinecone.ReadCapacityScaling{Manual: &pinecone.ReadCapacityManualScaling{Replicas: ptr(int32(1)), Shards: ptr(int32(1))}},
			Status:   pinecone.ReadCapacityStatus{State: "Ready"},
		}},
	}

	var model IndexModel
	if diags := model.Read(t.Context(), index); diags.HasError() {
		t.Fatalf("Read returned errors: %v", diags)
	}
	if got := model.DeletionProtection.ValueString(); got != "enabled" {
		t.Errorf("deletion_protection = %q, want %q", got, "enabled")
	}
	if model.PrivateHost.ValueString() != "my-index.private.example.com" || model.CmekId.ValueString() != "cmek-123" ||
		model.SourceBackupId.ValueString() != "backup-456" || !model.SourceCollection.IsNull() {
		t.Errorf("model = %+v", model)
	}
	if !model.Schema.IsNull() || !model.Deployment.IsNull() {
		t.Errorf("schema = %v, deployment = %v; want null", model.Schema, model.Deployment)
	}
	var readCapacity IndexReadCapacityModel
	if d := model.ReadCapacity.As(t.Context(), &readCapacity, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding read_capacity: %v", d)
	}
	if readCapacity.Dedicated.IsNull() || !readCapacity.OnDemand.IsNull() {
		t.Errorf("read_capacity = %+v, want dedicated", readCapacity)
	}
}

func TestIndexDatasourceModelRead(t *testing.T) {
	var model IndexDatasourceModel
	if diags := model.Read(t.Context(), &pinecone.Index{Name: "my-index", DeletionProtection: pinecone.DeletionProtectionDisabled}); diags.HasError() {
		t.Fatalf("Read returned errors: %v", diags)
	}
	if model.Id.ValueString() != "my-index" || model.Name.ValueString() != "my-index" || model.DeletionProtection.ValueString() != "disabled" {
		t.Errorf("id = %v, name = %v, deletion_protection = %v", model.Id, model.Name, model.DeletionProtection)
	}
}
