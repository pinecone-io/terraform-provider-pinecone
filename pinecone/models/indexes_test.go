// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
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
		Metric:             pinecone.Cosine,
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
		Status: &pinecone.IndexStatus{Ready: true, State: pinecone.Ready},
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
