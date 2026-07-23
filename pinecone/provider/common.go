package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

// isNotFoundErr reports whether an admin API error indicates the resource no
// longer exists, so callers can treat it as a removed resource rather than a
// hard failure.
func isNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "NOT_FOUND") ||
		strings.Contains(msg, "404")
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
