package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
)

type PineconeDatasource struct {
	client      *pinecone.Client
	adminClient *pinecone.AdminClient
}

func (d *PineconeDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// Try to get the regular client first
	if client, ok := req.ProviderData.(*pinecone.Client); ok {
		d.client = client
		return
	}

	// Try to get the admin client
	if adminClient, ok := req.ProviderData.(*pinecone.AdminClient); ok {
		d.adminClient = adminClient
		return
	}

	// Try to get the combined client structure
	if clientData, ok := req.ProviderData.(map[string]interface{}); ok {
		if client, ok := clientData["client"].(*pinecone.Client); ok {
			d.client = client
		}
		if adminClient, ok := clientData["adminClient"].(*pinecone.AdminClient); ok {
			d.adminClient = adminClient
		}
		return
	}

	resp.Diagnostics.AddError(
		"Unexpected Client Type",
		fmt.Sprintf("Expected *pinecone.Client or *pinecone.AdminClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
	)
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

	// Try to get the regular client first
	if client, ok := req.ProviderData.(*pinecone.Client); ok {
		d.client = client
		return
	}

	// Try to get the admin client
	if adminClient, ok := req.ProviderData.(*pinecone.AdminClient); ok {
		d.adminClient = adminClient
		return
	}

	// Try to get the combined client structure
	if clientData, ok := req.ProviderData.(map[string]interface{}); ok {
		if client, ok := clientData["client"].(*pinecone.Client); ok {
			d.client = client
		}
		if adminClient, ok := clientData["adminClient"].(*pinecone.AdminClient); ok {
			d.adminClient = adminClient
		}
		return
	}

	resp.Diagnostics.AddError(
		"Unexpected Client Type",
		fmt.Sprintf("Expected *pinecone.Client or *pinecone.AdminClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
	)
}
