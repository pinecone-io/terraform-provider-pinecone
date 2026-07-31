// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

func pineconeErr(code int) *pinecone.PineconeError {
	return &pinecone.PineconeError{Code: code, Msg: errors.New("boom")}
}

func TestHasStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
		want bool
	}{
		{"nil error", nil, http.StatusNotFound, false},
		{"matching code", pineconeErr(http.StatusNotFound), http.StatusNotFound, true},
		{"non-matching code", pineconeErr(http.StatusConflict), http.StatusNotFound, false},
		{"wrapped matching code", fmt.Errorf("describe failed: %w", pineconeErr(http.StatusNotFound)), http.StatusNotFound, true},
		{"non-pinecone error", errors.New("boom"), http.StatusNotFound, false},
		// Regression guard: the old string-matching heuristic would have matched
		// the embedded "404" here; the typed check must not.
		{"plain error containing status text", errors.New(`{"status_code":404,"body":"not found"}`), http.StatusNotFound, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasStatusCode(tt.err, tt.code); got != tt.want {
				t.Errorf("hasStatusCode(%v, %d) = %v, want %v", tt.err, tt.code, got, tt.want)
			}
		})
	}
}

func TestRoleBindingScopeError(t *testing.T) {
	tests := []struct {
		name         string
		resourceType types.String
		resourceId   types.String
		wantOk       bool
		wantSummary  string
	}{
		{"project with id", types.StringValue("project"), types.StringValue("proj-1"), true, ""},
		{"project without id", types.StringValue("project"), types.StringNull(), false, "Missing resource_id"},
		{"organization without id", types.StringValue("organization"), types.StringNull(), true, ""},
		{"organization with id", types.StringValue("organization"), types.StringValue("org-1"), false, "Unexpected resource_id"},
		// An unknown resource_id means the attribute is configured, just not yet
		// resolved, so scope rules key on configured-ness and still apply.
		{"project with unknown id", types.StringValue("project"), types.StringUnknown(), true, ""},
		{"organization with unknown id", types.StringValue("organization"), types.StringUnknown(), false, "Unexpected resource_id"},
		// An unknown or absent resource_type carries no scope to check; Create
		// re-checks once the value is known.
		{"unknown resource_type", types.StringUnknown(), types.StringNull(), true, ""},
		{"null resource_type", types.StringNull(), types.StringValue("proj-1"), true, ""},
		// Unrecognized scopes are left to the schema's OneOf validator and the API.
		{"unrecognized resource_type", types.StringValue("nonsense"), types.StringValue("x"), true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, detail, ok := roleBindingScopeError(tt.resourceType, tt.resourceId)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v (summary %q, detail %q)", ok, tt.wantOk, summary, detail)
			}
			if summary != tt.wantSummary {
				t.Errorf("summary = %q, want %q", summary, tt.wantSummary)
			}
			if !ok && detail == "" {
				t.Error("expected a non-empty detail alongside a failure")
			}
		})
	}
}

func TestIsNotFoundErr(t *testing.T) {
	if !isNotFoundErr(pineconeErr(http.StatusNotFound)) {
		t.Error("expected a 404 PineconeError to be treated as not found")
	}
	if isNotFoundErr(pineconeErr(http.StatusConflict)) {
		t.Error("expected a 409 PineconeError to not be treated as not found")
	}
	if isNotFoundErr(errors.New("not found")) {
		t.Error("expected a plain error containing 'not found' to not be treated as not found")
	}
	if isNotFoundErr(nil) {
		t.Error("expected a nil error to not be treated as not found")
	}
}

func TestIsConflictErr(t *testing.T) {
	if !isConflictErr(pineconeErr(http.StatusConflict)) {
		t.Error("expected a 409 PineconeError to be treated as a conflict")
	}
	if isConflictErr(pineconeErr(http.StatusNotFound)) {
		t.Error("expected a 404 PineconeError to not be treated as a conflict")
	}
	if isConflictErr(errors.New("conflict")) {
		t.Error("expected a plain error containing 'conflict' to not be treated as a conflict")
	}
}

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
	ctx := t.Context()
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
	ctx := t.Context()
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
