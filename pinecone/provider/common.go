package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

type PineconeDatasource struct {
	client *pinecone.Client
}

func (d *PineconeDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pinecone.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Client Type",
			fmt.Sprintf("Expected *pinecone.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

type PineconeResource struct {
	client *pinecone.Client
}

func (d *PineconeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pinecone.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Client Type",
			fmt.Sprintf("Expected *pinecone.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}
