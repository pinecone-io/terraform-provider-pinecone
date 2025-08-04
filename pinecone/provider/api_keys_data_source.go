// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &APIKeysDataSource{}

func NewAPIKeysDataSource() datasource.DataSource {
	return &APIKeysDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// APIKeysDataSource defines the data source implementation.
type APIKeysDataSource struct {
	*PineconeDatasource
}

func (d *APIKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *APIKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_api_keys` data source lets you retrieve information about all API keys in a Pinecone project.",

		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project to list API keys for",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"api_keys": schema.ListNestedAttribute{
				MarkdownDescription: "List of API keys",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "API key identifier",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *APIKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.APIKeysDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project ID from the data source
	var projectId types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("project_id"), &projectId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all API keys for the project using Admin API
	apiKeys, err := d.adminClient.APIKey.List(ctx, projectId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read API keys", err.Error())
		return
	}

	// Convert API keys to models
	var apiKeyModels []models.APIKeyModel
	for _, apiKey := range apiKeys {
		var apiKeyModel models.APIKeyModel
		resp.Diagnostics.Append(apiKeyModel.Read(ctx, apiKey)...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiKeyModels = append(apiKeyModels, apiKeyModel)
	}

	data.APIKeys = apiKeyModels
	data.Id = types.StringValue("api_keys")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
} 