package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
		MarkdownDescription: "Project data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project.",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID where the project is located.",
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
	}
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ProjectDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if admin client is available
	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read projects.")
		return
	}

	project, err := d.adminClient.Project.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to describe project, got error: %s", err))
		return
	}

	// Save data into Terraform state
	data.Read(project)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
