// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &APIKeyDataSource{}

func NewAPIKeyDataSource() datasource.DataSource {
	return &APIKeyDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// APIKeyDataSource defines the data source implementation.
type APIKeyDataSource struct {
	*PineconeDatasource
}

func (d *APIKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *APIKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_api_key` data source lets you retrieve information about a specific API key in Pinecone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "API key identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the API key",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project that owns this API key",
				Computed:            true,
			},
			"roles": schema.ListAttribute{
				MarkdownDescription: "The roles assigned to the API key",
				Computed:            true,
				ElementType:         basetypes.StringType{},
			},
		},
	}
}

func (d *APIKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// Try to get the admin client
	if adminClient, ok := req.ProviderData.(*pinecone.AdminClient); ok {
		d.adminClient = adminClient
		return
	}

	// Try to get the combined client structure
	if clientData, ok := req.ProviderData.(map[string]interface{}); ok {
		if adminClient, ok := clientData["adminClient"].(*pinecone.AdminClient); ok {
			d.adminClient = adminClient
			return
		}
	}

	resp.Diagnostics.AddError(
		"Unexpected Client Type",
		"Expected *pinecone.AdminClient. Please report this issue to the provider developers.",
	)
}

func (d *APIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.APIKeyDatasourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the API key using Admin API
	apiKey, err := d.adminClient.APIKey.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read API key", err.Error())
		return
	}

	// Read the API key into the model
	resp.Diagnostics.Append(data.Read(ctx, apiKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
} 