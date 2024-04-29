// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectApiKeyDataSource{}

func NewProjectApiKeyDataSource() datasource.DataSource {
	return &ProjectApiKeyDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// ProjectApiKeyDataSource defines the data source implementation.
type ProjectApiKeyDataSource struct {
	*PineconeDatasource
}

func (d *ProjectApiKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_api_key"
}

func (d *ProjectApiKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Project ApiKey data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ApiKey identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "ApiKey Name",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Required:            true,
			},
		},
	}
}

func (d *ProjectApiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ProjectApiKeyModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId, _ := uuid.Parse(data.ProjectId.ValueString())
	apiKeys, err := d.mgmtClient.ListApiKeys(ctx, projectId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list project api keys", err.Error())
		return
	}

	for _, apiKey := range apiKeys {
		if apiKey.Name == *data.Name.ValueStringPointer() {
			data.ReadWithoutSecret(ctx, apiKey)
			break
		}
	}

	if data.Id.IsNull() {
		resp.Diagnostics.AddError("Failed to find project api key with name: ", data.Name.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
