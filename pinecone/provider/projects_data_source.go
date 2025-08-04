// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectsDataSource{}

func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// ProjectsDataSource defines the data source implementation.
type ProjectsDataSource struct {
	*PineconeDatasource
}

func (d *ProjectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *ProjectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The `pinecone_projects` data source lets you retrieve information about all projects in Pinecone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"projects": schema.ListNestedAttribute{
				MarkdownDescription: "List of projects",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Project identifier",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *ProjectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ProjectsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all projects using Admin API
	projects, err := d.adminClient.Project.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read projects", err.Error())
		return
	}

	// Convert projects to models
	var projectModels []models.ProjectModel
	for _, project := range projects {
		var projectModel models.ProjectModel
		resp.Diagnostics.Append(projectModel.Read(ctx, project)...)
		if resp.Diagnostics.HasError() {
			return
		}
		projectModels = append(projectModels, projectModel)
	}

	data.Projects = projectModels
	data.Id = types.StringValue("projects")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
