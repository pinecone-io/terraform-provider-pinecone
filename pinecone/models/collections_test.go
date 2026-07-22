// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v5/pinecone"
)

// TestCollectionResourceModelRead covers the collection state mapping. Collections
// are a pod-only, legacy feature no longer exercised by acceptance tests, so this
// unit test guards the API-response -> Terraform-state translation that used to be
// covered end-to-end by TestAccCollectionResource.
func TestCollectionResourceModelRead(t *testing.T) {
	collection := &pinecone.Collection{
		Name:        "my-collection",
		Size:        1024,
		Status:      pinecone.CollectionStatusReady,
		Dimension:   1536,
		VectorCount: 42,
		Environment: "us-west4-gcp",
	}

	model := CollectionResourceModel{
		// Source is not returned by the API; Read must preserve the configured value.
		Source: types.StringValue("source-index"),
	}
	model.Read(collection)

	if got := model.Id.ValueString(); got != "my-collection" {
		t.Errorf("Id = %q, want %q", got, "my-collection")
	}
	if got := model.Name.ValueString(); got != "my-collection" {
		t.Errorf("Name = %q, want %q", got, "my-collection")
	}
	if got := model.Source.ValueString(); got != "source-index" {
		t.Errorf("Source = %q, want %q (should be preserved)", got, "source-index")
	}
	if got := model.Status.ValueString(); got != "Ready" {
		t.Errorf("Status = %q, want %q", got, "Ready")
	}
	if got := model.Size.ValueInt64(); got != 1024 {
		t.Errorf("Size = %d, want 1024", got)
	}
	if got := model.Dimension.ValueInt32(); got != 1536 {
		t.Errorf("Dimension = %d, want 1536", got)
	}
	if got := model.VectorCount.ValueInt32(); got != 42 {
		t.Errorf("VectorCount = %d, want 42", got)
	}
	if got := model.Environment.ValueString(); got != "us-west4-gcp" {
		t.Errorf("Environment = %q, want %q", got, "us-west4-gcp")
	}
}
