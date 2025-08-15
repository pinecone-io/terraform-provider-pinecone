// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
)

// NewTestClient returns a new Pinecone API client instance
// to be used in acceptance tests.
func NewTestClient() (*pinecone.Client, error) {
	apiKey := os.Getenv("PINECONE_API_KEY")

	return pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
}

func TestDatasource_Configure(t *testing.T) {
	// Create a test PineconeProviderData
	testProviderData := &PineconeProviderData{
		Client:      &pinecone.Client{},
		AdminClient: nil,
	}

	// Create a mock context and request
	ctx := context.Background()
	req := datasource.ConfigureRequest{
		ProviderData: testProviderData,
	}
	resp := &datasource.ConfigureResponse{}

	r := &PineconeDatasource{}

	// Call the Configure function with the test data
	r.Configure(ctx, req, resp)

	// Check if the client field in r has been correctly set
	if r.client != nil && r.client != testProviderData.Client {
		t.Errorf("Expected r.client to be set to the test client, got: %v", r.client)
	}

	// Now, let's test the case where req.ProviderData is not *PineconeProviderData
	invalidReq := datasource.ConfigureRequest{
		ProviderData: "not a *PineconeProviderData", // Pass a non-*PineconeProviderData value
	}
	invalidResp := &datasource.ConfigureResponse{}

	// Call the Configure function with the invalid data
	r.Configure(ctx, invalidReq, invalidResp)

	// Check if the Diagnostics field in the response contains an error
	if !invalidResp.Diagnostics.HasError() {
		t.Error("Expected an error in resp.Diagnostics.Errors, but found none")
	} else {
		// Check the error message
		expectedErrorMessage := "Expected *PineconeProviderData, got: string. Please report this issue to the provider developers."
		actualErrorMessage := invalidResp.Diagnostics.Errors()[0].Detail()
		if actualErrorMessage != expectedErrorMessage {
			t.Errorf("Expected error message: %s, got: %s", expectedErrorMessage, actualErrorMessage)
		}
	}
}

func TestResource_Configure(t *testing.T) {
	// Create a test PineconeProviderData
	testProviderData := &PineconeProviderData{
		Client:      &pinecone.Client{},
		AdminClient: nil,
	}

	// Create a mock context and request
	ctx := context.Background()
	req := resource.ConfigureRequest{
		ProviderData: testProviderData,
	}
	resp := &resource.ConfigureResponse{}

	r := &PineconeResource{}

	// Call the Configure function with the test data
	r.Configure(ctx, req, resp)

	// Check if the client field in r has been correctly set
	if r.client != nil && r.client != testProviderData.Client {
		t.Errorf("Expected r.client to be set to the test client, got: %v", r.client)
	}

	// Now, let's test the case where req.ProviderData is not *PineconeProviderData
	invalidReq := resource.ConfigureRequest{
		ProviderData: "not a *PineconeProviderData", // Pass a non-*PineconeProviderData value
	}
	invalidResp := &resource.ConfigureResponse{}

	// Call the Configure function with the invalid data
	r.Configure(ctx, invalidReq, invalidResp)

	// Check if the Diagnostics field in the response contains an error
	if !invalidResp.Diagnostics.HasError() {
		t.Error("Expected an error in resp.Diagnostics.Errors, but found none")
	} else {
		// Check the error message
		expectedErrorMessage := "Expected *PineconeProviderData, got: string. Please report this issue to the provider developers."
		actualErrorMessage := invalidResp.Diagnostics.Errors()[0].Detail()
		if actualErrorMessage != expectedErrorMessage {
			t.Errorf("Expected error message: %s, got: %s", expectedErrorMessage, actualErrorMessage)
		}
	}
}
