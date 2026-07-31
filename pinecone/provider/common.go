package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

// roleBindingScopeError checks a role binding's resource_type against its
// resource_id and returns the diagnostic to raise, or ok when they agree.
// Project scope requires a resource_id; organization scope forbids one (the
// server scopes the binding to the caller's organization automatically).
//
// It keys on whether resource_id is configured at all rather than on its value,
// so an unknown resource_id — a project id from another resource, say —
// validates correctly during plan. An unknown resource_type is skipped and
// re-checked at apply, when its value is known.
func roleBindingScopeError(resourceType, resourceId types.String) (summary, detail string, ok bool) {
	if resourceType.IsNull() || resourceType.IsUnknown() {
		return "", "", true
	}

	hasResourceId := !resourceId.IsNull()

	switch resourceType.ValueString() {
	case string(pinecone.ResourceTypeProject):
		if !hasResourceId {
			return "Missing resource_id", "resource_id is required when resource_type is \"project\".", false
		}
	case string(pinecone.ResourceTypeOrganization):
		if hasResourceId {
			return "Unexpected resource_id", "resource_id must be omitted when resource_type is \"organization\".", false
		}
	}

	return "", "", true
}

// hasStatusCode reports whether err is (or wraps) a *pinecone.PineconeError
// whose HTTP status code equals code. The SDK returns this typed error for all
// non-2xx admin API responses, so matching on the code is more reliable than
// scanning the rendered error string (which embeds the raw response body).
// Non-API errors (transport failures, uuid.Parse failures) do not match, which
// is the desired behavior for the callers below.
func hasStatusCode(err error, code int) bool {
	var pineconeErr *pinecone.PineconeError
	return errors.As(err, &pineconeErr) && pineconeErr.Code == code
}

// isNotFoundErr reports whether an admin API error indicates the resource no
// longer exists, so callers can treat it as a removed resource rather than a
// hard failure.
func isNotFoundErr(err error) bool {
	return hasStatusCode(err, http.StatusNotFound)
}

// isConflictErr reports whether an admin API error is an HTTP 409 Conflict.
func isConflictErr(err error) bool {
	return hasStatusCode(err, http.StatusConflict)
}

// retryDeletion polls describe until it reports the resource is gone. Admin
// deletes are asynchronous (HTTP 202), so callers pass a describe closure; a
// not-found error means deletion completed, any other error is retried, and a
// successful describe means the resource still exists.
func retryDeletion(ctx context.Context, describe func() error) error {
	return retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		err := describe()
		if err != nil {
			if isNotFoundErr(err) {
				return nil
			}
			return retry.RetryableError(fmt.Errorf("deletion verification in progress, retrying: %w", err))
		}
		return retry.RetryableError(fmt.Errorf("resource not deleted yet"))
	})
}

type PineconeDatasource struct {
	client      *pinecone.Client
	adminClient *pinecone.AdminClient
}

func (d *PineconeDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*PineconeProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *PineconeProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
	d.adminClient = providerData.AdminClient
}

type PineconeResource struct {
	client      *pinecone.Client
	adminClient *pinecone.AdminClient
}

func (d *PineconeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*PineconeProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *PineconeProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
	d.adminClient = providerData.AdminClient
}
