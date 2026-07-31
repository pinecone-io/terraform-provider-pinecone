package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RoleBindingDataSource{}

func NewRoleBindingDataSource() datasource.DataSource {
	return &RoleBindingDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// RoleBindingDataSource defines the data source implementation.
type RoleBindingDataSource struct {
	*PineconeDatasource
}

func (d *RoleBindingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_binding"
}

func (d *RoleBindingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch a single role binding by ID. The `pinecone_role_bindings` list cannot filter on a binding's own ID, so use this when you already have one. To find the bindings held by a principal instead, use `pinecone_role_bindings` with `principal_type` and `principal_id`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the role binding.",
				Required:            true,
			},
			"principal_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the principal that receives permissions.",
				Computed:            true,
			},
			"principal_type": schema.StringAttribute{
				MarkdownDescription: "The kind of principal that receives permissions: `user`, `service_account`, `api_key`, or `invite`.",
				Computed:            true,
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "The kind of resource scope the binding applies to: `organization` or `project`.",
				Computed:            true,
			},
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization or project the binding is scoped to. Note that this always reports the server value, so an organization-scoped binding returns the organization ID — unlike the `pinecone_role_binding` resource, which keeps `resource_id` null under organization scope to match the config.",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role assigned to the principal at the resource scope.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the role binding was created.",
				Computed:            true,
			},
		},
	}
}

func (d *RoleBindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.RoleBindingDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read role bindings.")
		return
	}

	roleBinding, err := d.adminClient.RoleBinding.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe role binding", err.Error())
		return
	}

	data.Read(roleBinding)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
