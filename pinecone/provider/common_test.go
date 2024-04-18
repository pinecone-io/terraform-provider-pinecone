// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"testing"
)

func TestDatasource_Configure(t *testing.T) {
	// Create test pinecone clients
	testClients := &PineconeClients{
		Client:     &pinecone.Client{},
		MgmtClient: &pinecone.ManagementClient{},
	}
	// Create a mock context and request
	ctx := context.Background()
	req := datasource.ConfigureRequest{
		ProviderData: &testClients,
	}
	resp := &datasource.ConfigureResponse{}

	r := &PineconeDatasource{}

	// Call the Configure function with the test data
	r.Configure(ctx, req, resp)

	// Check if the client field in r has been correctly set
	if r.client != nil && r.client != testClients.Client {
		t.Errorf("Expected r.client to be set to the test client, got: %v", r.client)
	}
	// Check if the mgmtClient field in r has been correctly set
	if r.mgmtClient != nil && r.mgmtClient != testClients.MgmtClient {
		t.Errorf("Expected r.mgmtClient to be set to the test client, got: %v", r.mgmtClient)
	}

	// Now, let's test the case where req.ProviderData is not ProviderClients
	invalidReq := datasource.ConfigureRequest{
		ProviderData: "not a *pinecone.Client", // Pass a non-*pinecone.Client value
	}
	invalidResp := &datasource.ConfigureResponse{}

	// Call the Configure function with the invalid data
	r.Configure(ctx, invalidReq, invalidResp)

	// Check if the Diagnostics field in the response contains an error
	if !invalidResp.Diagnostics.HasError() {
		t.Error("Expected an error in resp.Diagnostics.Errors, but found none")
	} else {
		// Check the error message
		expectedErrorMessage := "Expected *pinecone.Client, got: string. Please report this issue to the provider developers."
		actualErrorMessage := invalidResp.Diagnostics.Errors()[0].Detail()
		if actualErrorMessage != expectedErrorMessage {
			t.Errorf("Expected error message: %s, got: %s", expectedErrorMessage, actualErrorMessage)
		}
	}
}

func TestResource_Configure(t *testing.T) {
	// Create test pinecone clients
	testClients := &PineconeClients{
		Client:     &pinecone.Client{},
		MgmtClient: &pinecone.ManagementClient{},
	}
	// Create a mock context and request
	ctx := context.Background()
	req := resource.ConfigureRequest{
		ProviderData: &testClients,
	}
	resp := &resource.ConfigureResponse{}

	r := &PineconeResource{}

	// Call the Configure function with the test data
	r.Configure(ctx, req, resp)

	// Check if the client fields in r have been correctly set
	if r.client != nil && r.client != testClients.Client {
		t.Errorf("Expected r.client to be set to the test client, got: %v", r.client)
	}
	if r.mgmtClient != nil && r.mgmtClient != testClients.MgmtClient {
		t.Errorf("Expected r.mgmtClient to be set to the test mgmt client, got: %v", r.mgmtClient)
	}

	// Now, let's test the case where req.ProviderData is not *pinecone.Client
	invalidReq := resource.ConfigureRequest{
		ProviderData: "not a *pinecone.Client", // Pass a non-*pinecone.Client value
	}
	invalidResp := &resource.ConfigureResponse{}

	// Call the Configure function with the invalid data
	r.Configure(ctx, invalidReq, invalidResp)

	// Check if the Diagnostics field in the response contains an error
	if !invalidResp.Diagnostics.HasError() {
		t.Error("Expected an error in resp.Diagnostics.Errors, but found none")
	} else {
		// Check the error message
		expectedErrorMessage := "Expected *PineconeClients, got: string. Please report this issue to the provider developers."
		actualErrorMessage := invalidResp.Diagnostics.Errors()[0].Detail()
		if actualErrorMessage != expectedErrorMessage {
			t.Errorf("Expected error message: %s, got: %s", expectedErrorMessage, actualErrorMessage)
		}
	}
}
