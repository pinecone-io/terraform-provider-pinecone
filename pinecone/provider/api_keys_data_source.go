// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectApiKeysDataSource{}

func NewProjectApiKeysDataSource() datasource.DataSource {
	return &ProjectApiKeysDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// ProjectsDataSource defines the data source implementation.
type ProjectApiKeysDataSource struct {
	*PineconeDatasource
}

func (d *ProjectApiKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_api_keys"
}

func (d *ProjectApiKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Project ApiKeys data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ApiKeys identifier",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Required:            true,
			},
			"api_keys": schema.ListNestedAttribute{
				MarkdownDescription: "List of the api keys in your project",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Project ApiKey name",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "Project ApiKey identifier",
							Computed:            true,
						},
						"project_id": schema.StringAttribute{
							MarkdownDescription: "Project identifier",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ProjectApiKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ProjectApiKeysModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId, _ := uuid.Parse(data.ProjectId.ValueString())
	apiKeys, err := d.mgmtClient.ListApiKeys(ctx, projectId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListApiKeys, got error: %s", err))
		return
	}

	for _, i := range apiKeys {
		apiKey := models.ProjectApiKeyModel{}
		resp.Diagnostics.Append(apiKey.ReadWithoutSecret(ctx, i)...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ApiKeys = append(data.ApiKeys, apiKey)
	}

	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
