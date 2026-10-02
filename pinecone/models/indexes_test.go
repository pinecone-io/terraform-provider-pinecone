// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
)

// TestIndexResourceModelRead_pod covers the pod-based state mapping. Pod indexes
// are a legacy feature no longer exercised by acceptance tests, so this unit
// test guards the API-response -> Terraform-state translation that used to be
// covered end-to-end by TestAccIndexResource_pod_basic.
func TestIndexResourceModelRead_pod(t *testing.T) {
	ctx := t.Context()
	dim := int32(1536)
	tags := pinecone.IndexTags{"team": "search"}
	index := &pinecone.Index{
		Name:               "my-pod-index",
		Host:               "https://my-pod-index.example.com",
		Metric:             pinecone.IndexMetricCosine,
		VectorType:         "dense",
		DeletionProtection: pinecone.DeletionProtectionEnabled,
		Dimension:          &dim,
		Spec: &pinecone.IndexSpec{
			Pod: &pinecone.PodSpec{
				Environment: "us-west4-gcp",
				PodType:     "s1.x1",
				PodCount:    2,
				Replicas:    2,
				ShardCount:  1,
			},
		},
		Status: &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady},
		Tags:   &tags,
	}

	var model IndexResourceModel
	if diags := model.Read(ctx, index); diags.HasError() {
		t.Fatalf("Read returned errors: %v", diags)
	}

	if got := model.Id.ValueString(); got != "my-pod-index" {
		t.Errorf("Id = %q, want %q", got, "my-pod-index")
	}
	if got := model.Name.ValueString(); got != "my-pod-index" {
		t.Errorf("Name = %q, want %q", got, "my-pod-index")
	}
	if got := model.Dimension.ValueInt32(); got != 1536 {
		t.Errorf("Dimension = %d, want 1536", got)
	}
	if got := model.Metric.ValueString(); got != "cosine" {
		t.Errorf("Metric = %q, want %q", got, "cosine")
	}
	if got := model.Host.ValueString(); got != "https://my-pod-index.example.com" {
		t.Errorf("Host = %q, want %q", got, "https://my-pod-index.example.com")
	}
	if got := model.DeletionProtection.ValueString(); got != "enabled" {
		t.Errorf("DeletionProtection = %q, want %q", got, "enabled")
	}

	var spec IndexSpecModel
	if d := model.Spec.As(ctx, &spec, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding spec: %v", d)
	}
	if spec.Pod == nil {
		t.Fatal("expected a pod spec, got nil")
	}
	if spec.Serverless != nil {
		t.Error("expected nil serverless spec on a pod index")
	}
	if got := spec.Pod.Environment.ValueString(); got != "us-west4-gcp" {
		t.Errorf("pod.environment = %q, want %q", got, "us-west4-gcp")
	}
	if got := spec.Pod.PodType.ValueString(); got != "s1.x1" {
		t.Errorf("pod.pod_type = %q, want %q", got, "s1.x1")
	}
	if got := spec.Pod.Replicas.ValueInt64(); got != 2 {
		t.Errorf("pod.replicas = %d, want 2", got)
	}
	if got := spec.Pod.ShardCount.ValueInt64(); got != 1 {
		t.Errorf("pod.shards = %d, want 1", got)
	}
	// pods is the computed shards * replicas value returned by the API.
	if got := spec.Pod.PodCount.ValueInt64(); got != 2 {
		t.Errorf("pod.pods = %d, want 2", got)
	}
}

// TestIndexResourceModelRead_nil verifies Read fails cleanly (diagnostic, not
// panic) when handed a nil index pointer.
func TestIndexResourceModelRead_nil(t *testing.T) {
	var model IndexResourceModel
	if diags := model.Read(t.Context(), nil); !diags.HasError() {
		t.Fatal("expected an error diagnostic for a nil index, got none")
	}
}

// serverlessIndexWithMetadata returns a dense serverless index as 2026-07 describes it: the
// reserved vector fields, a legacy metadata field from a metadata schema set at creation, and a
// typed metadata field the API added when data was upserted.
func serverlessIndexWithMetadata() *pinecone.Index {
	dim := int32(1024)
	filterable := true
	return &pinecone.Index{
		Name:       "my-index",
		Host:       "https://my-index.example.com",
		Metric:     pinecone.IndexMetricCosine,
		VectorType: "dense",
		Dimension:  &dim,
		Schema: &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
			"_values":        {DenseVector: &pinecone.DenseVectorField{Dimension: dim, Metric: pinecone.IndexMetricCosine}},
			"_sparse_values": {SparseVector: &pinecone.SparseVectorField{}},
			"genre":          {LegacyMetadata: &pinecone.LegacyMetadataField{Filterable: true}},
			"year":           {Float: &pinecone.FloatField{Filterable: &filterable}},
		}},
		Spec:   &pinecone.IndexSpec{Serverless: &pinecone.ServerlessSpec{Cloud: pinecone.CloudAWS, Region: "us-west-2"}},
		Status: &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady},
	}
}

func serverlessSchemaFields(t *testing.T, model IndexResourceModel) map[string]IndexMetadataSchemaFieldModel {
	t.Helper()
	ctx := t.Context()
	var spec IndexSpecModel
	if d := model.Spec.As(ctx, &spec, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding spec: %v", d)
	}
	if spec.Serverless == nil {
		t.Fatal("expected a serverless spec, got nil")
	}
	if spec.Serverless.Schema.IsNull() {
		return nil
	}
	var schema IndexMetadataSchemaModel
	if d := spec.Serverless.Schema.As(ctx, &schema, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding schema: %v", d)
	}
	var fields map[string]IndexMetadataSchemaFieldModel
	if d := schema.Fields.ElementsAs(ctx, &fields, false); d.HasError() {
		t.Fatalf("decoding schema fields: %v", d)
	}
	return fields
}

// TestIndexResourceModelRead_importRebuildsMetadataSchema covers import, where there's no prior
// state: the metadata schema comes from the legacy metadata fields alone, not from vector fields
// or metadata fields the API indexed on upsert.
func TestIndexResourceModelRead_importRebuildsMetadataSchema(t *testing.T) {
	var model IndexResourceModel
	if diags := model.Read(t.Context(), serverlessIndexWithMetadata()); diags.HasError() {
		t.Fatalf("Read returned errors: %v", diags)
	}

	fields := serverlessSchemaFields(t, model)
	if len(fields) != 1 || !fields["genre"].Filterable.ValueBool() {
		t.Errorf("schema fields = %v, want only genre (filterable)", fields)
	}
}

// TestIndexResourceModelRead_keepsPriorMetadataSchema covers refresh: describe responses no longer
// report the metadata schema, so the value in prior state is kept as is.
func TestIndexResourceModelRead_keepsPriorMetadataSchema(t *testing.T) {
	ctx := t.Context()
	var model IndexResourceModel
	if diags := model.Read(ctx, serverlessIndexWithMetadata()); diags.HasError() {
		t.Fatalf("first Read returned errors: %v", diags)
	}

	// Simulate an index created without a metadata schema: prior state has none.
	var spec IndexSpecModel
	if d := model.Spec.As(ctx, &spec, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding spec: %v", d)
	}
	spec.Serverless.Schema = types.ObjectNull(IndexMetadataSchemaModel{}.AttrTypes())
	var d diag.Diagnostics
	model.Spec, d = types.ObjectValueFrom(ctx, model.Spec.AttributeTypes(ctx), spec)
	if d.HasError() {
		t.Fatalf("encoding spec: %v", d)
	}

	if diags := model.Read(ctx, serverlessIndexWithMetadata()); diags.HasError() {
		t.Fatalf("second Read returned errors: %v", diags)
	}
	if fields := serverlessSchemaFields(t, model); fields != nil {
		t.Errorf("schema fields = %v, want null schema kept from prior state", fields)
	}
}

// TestIndexResourceModelRead_keepsPodMetadataConfig covers the pod metadata config, which describe
// responses no longer report.
func TestIndexResourceModelRead_keepsPodMetadataConfig(t *testing.T) {
	ctx := t.Context()
	index := &pinecone.Index{
		Name:   "my-pod-index",
		Spec:   &pinecone.IndexSpec{Pod: &pinecone.PodSpec{Environment: "us-west4-gcp", PodType: "s1.x1", Replicas: 1, ShardCount: 1, PodCount: 1}},
		Status: &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady},
	}

	indexed := []string{"genre"}
	priorPod, d := NewIndexPodSpecModel(ctx, &pinecone.PodSpec{
		Environment: "us-west4-gcp", PodType: "s1.x1", Replicas: 1, ShardCount: 1, PodCount: 1,
		MetadataConfig: &pinecone.PodSpecMetadataConfig{Indexed: &indexed},
	})
	if d.HasError() {
		t.Fatalf("building prior pod spec: %v", d)
	}
	var model IndexResourceModel
	model.Spec, d = types.ObjectValueFrom(ctx, indexSpecResourceAttrTypes(), IndexSpecModel{Pod: priorPod})
	if d.HasError() {
		t.Fatalf("encoding prior spec: %v", d)
	}

	if diags := model.Read(ctx, index); diags.HasError() {
		t.Fatalf("Read returned errors: %v", diags)
	}

	var spec IndexSpecModel
	if d := model.Spec.As(ctx, &spec, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding spec: %v", d)
	}
	var metadataConfig IndexMetadataConfigModel
	if d := spec.Pod.MetadataConfig.As(ctx, &metadataConfig, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding metadata_config: %v", d)
	}
	var got []string
	if d := metadataConfig.Indexed.ElementsAs(ctx, &got, false); d.HasError() {
		t.Fatalf("decoding indexed: %v", d)
	}
	if len(got) != 1 || got[0] != "genre" {
		t.Errorf("metadata_config.indexed = %v, want [genre]", got)
	}
}

// TestIndexResourceModelRead_embedVectorType covers integrated indexes: 2026-07 derives embed from
// the semantic text field without a vector type, so it's taken from the index.
func TestIndexResourceModelRead_embedVectorType(t *testing.T) {
	ctx := t.Context()
	dim := int32(1024)
	metric := pinecone.IndexMetricCosine
	index := &pinecone.Index{
		Name:       "my-integrated-index",
		Metric:     metric,
		VectorType: "dense",
		Dimension:  &dim,
		Embed: &pinecone.IndexEmbed{
			Model:     "multilingual-e5-large",
			Dimension: &dim,
			Metric:    &metric,
			FieldMap:  &map[string]interface{}{"text": "chunk_text"},
		},
		Spec:   &pinecone.IndexSpec{Serverless: &pinecone.ServerlessSpec{Cloud: pinecone.CloudAWS, Region: "us-west-2"}},
		Status: &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady},
	}

	var model IndexResourceModel
	if diags := model.Read(ctx, index); diags.HasError() {
		t.Fatalf("Read returned errors: %v", diags)
	}
	var embed IndexEmbedResourceModel
	if d := model.Embed.As(ctx, &embed, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding embed: %v", d)
	}
	if got := embed.VectorType.ValueString(); got != "dense" {
		t.Errorf("embed.vector_type = %q, want %q", got, "dense")
	}
	if index.Embed.VectorType != nil {
		t.Error("Read modified the SDK's embed in place")
	}
}
