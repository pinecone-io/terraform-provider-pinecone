package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/testing/v2"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCollectionResource(t *testing.T) {
	t.Parallel()

	resourceType := NewCollectionResource()

	t.Run("Plan", func(t *testing.T) {
		req := resource.PlanResourceRequest{
			Config: tfsdk.NewAttributePathValue(tfsdk.NewAttributePath().WithAttributeName("name"), types.String{Value: "test-collection"}),
			State:  tfsdk.NewAttributePathValue(tfsdk.NewAttributePath().WithAttributeName("id"), types.String{}),
		}
		resp := resource.PlanResourceResponse{}

		resourceType.Plan(context.Background(), req, &resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Unexpected diagnostics: %s", resp.Diagnostics)
		}
	})

	t.Run("Apply", func(t *testing.T) {
		req := resource.ApplyResourceRequest{
			Config: tfsdk.NewAttributePathValue(tfsdk.NewAttributePath().WithAttributeName("name"), types.String{Value: "test-collection"}),
			State:  tfsdk.NewAttributePathValue(tfsdk.NewAttributePath().WithAttributeName("id"), types.String{}),
		}
		resp := resource.ApplyResourceResponse{}

		resourceType.Apply(context.Background(), req, &resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Unexpected diagnostics: %s", resp.Diagnostics)
		}
	})

	// Add more tests as needed, such as Read, Update, and Delete.
}

func TestCollectionResource_Acceptance(t *testing.T) {
	// This is a placeholder for acceptance tests which would involve real API calls.
	// You'd typically set up the environment, create real resources, and then tear them down.

	t.Skip("Acceptance tests are skipped for now. Remove this line to run them.")

	resourceType := NewCollectionResource()

	// Example acceptance test structure
	t.Run("Create and delete collection", func(t *testing.T) {
		// Setup, API calls, assertions
	})

	// Add more acceptance tests as needed.
}
