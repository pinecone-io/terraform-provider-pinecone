// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// ProjectDataSource defines the data source implementation.
type ProjectDataSource struct {
	*PineconeDatasource
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The `pinecone_project` data source lets you retrieve information about a specific project in Pinecone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID that the project belongs to",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the project was created",
				Computed:            true,
			},
			"force_encryption_with_cmek": schema.BoolAttribute{
				MarkdownDescription: "Whether to force encryption with a customer-managed encryption key (CMEK)",
				Computed:            true,
			},
			"max_pods": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of Pods that can be created in the project",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ProjectDatasourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the project using Admin API
	project, err := d.adminClient.Project.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read project", err.Error())
		return
	}

	// Read the project into the model
	resp.Diagnostics.Append(data.Read(ctx, project)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
