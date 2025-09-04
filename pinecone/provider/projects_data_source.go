// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		MarkdownDescription: "Projects data source",

		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				MarkdownDescription: "List of the projects in your organization",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique ID of the project.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the project.",
							Computed:            true,
						},
						"organization_id": schema.StringAttribute{
							MarkdownDescription: "The unique ID of the organization that the project belongs to.",
							Computed:            true,
						},
						"force_encryption_with_cmek": schema.BoolAttribute{
							MarkdownDescription: "Whether encryption with a customer-managed encryption key (CMEK) is forced.",
							Computed:            true,
						},
						"max_pods": schema.Int64Attribute{
							MarkdownDescription: "The maximum number of Pods that can be created in the project.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "The timestamp when the project was created.",
							Computed:            true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Projects identifier",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ProjectsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if admin client is available
	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to list projects.")
		return
	}

	projects, err := d.adminClient.Project.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list projects, got error: %s", err))
		return
	}

	// Convert projects to models and append to the list
	for _, p := range projects {
		data.Projects = append(data.Projects, *models.NewProjectModel(p))
	}

	// Save data into Terraform state
	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
