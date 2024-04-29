package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

type PineconeDatasource struct {
	client     *pinecone.Client
	mgmtClient *pinecone.ManagementClient
}

func (d *PineconeDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*PineconeClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Clients Type",
			fmt.Sprintf("Expected *PineconeClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = clients.Client
	d.mgmtClient = clients.MgmtClient
}

type PineconeResource struct {
	client     *pinecone.Client
	mgmtClient *pinecone.ManagementClient
}

func (d *PineconeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*PineconeClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Clients Type",
			fmt.Sprintf("Expected *PineconeClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = clients.Client
	d.mgmtClient = clients.MgmtClient
}
