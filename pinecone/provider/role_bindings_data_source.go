package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RoleBindingsDataSource{}

func NewRoleBindingsDataSource() datasource.DataSource {
	return &RoleBindingsDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// RoleBindingsDataSource defines the data source implementation.
type RoleBindingsDataSource struct {
	*PineconeDatasource
}

func (d *RoleBindingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_bindings"
}

func (d *RoleBindingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Role bindings data source. Lists role bindings in your organization, optionally filtered by principal, resource, or role.",

		Attributes: map[string]schema.Attribute{
			"principal_type": schema.StringAttribute{
				MarkdownDescription: "Filter by principal type. Required when `principal_id` is set. Valid values are: `user`, `service_account`, `api_key`, `invite`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(pinecone.PrincipalTypeUser),
						string(pinecone.PrincipalTypeServiceAccount),
						string(pinecone.PrincipalTypeAPIKey),
						string(pinecone.PrincipalTypeInvite),
					),
				},
			},
			"principal_id": schema.StringAttribute{
				MarkdownDescription: "Filter by principal ID. Requires `principal_type`.",
				Optional:            true,
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "Filter by resource type. Required when `resource_id` is set. Valid values are: `organization`, `project`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(pinecone.ResourceTypeOrganization),
						string(pinecone.ResourceTypeProject),
					),
				},
			},
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "Filter by resource ID. Requires `resource_type`.",
				Optional:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Filter by role.",
				Optional:            true,
			},
			"role_bindings": schema.ListNestedAttribute{
				MarkdownDescription: "The list of matching role bindings.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique ID of the role binding.",
							Computed:            true,
						},
						"principal_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the principal that receives permissions.",
							Computed:            true,
						},
						"principal_type": schema.StringAttribute{
							MarkdownDescription: "The kind of principal that receives permissions.",
							Computed:            true,
						},
						"resource_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the organization or project the binding is scoped to.",
							Computed:            true,
						},
						"resource_type": schema.StringAttribute{
							MarkdownDescription: "The kind of resource scope the binding applies to.",
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
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Role bindings identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *RoleBindingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.RoleBindingsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to list role bindings.")
		return
	}

	listParams := &pinecone.ListRoleBindingsParams{}
	if !data.PrincipalType.IsNull() {
		pt := pinecone.PrincipalType(data.PrincipalType.ValueString())
		listParams.PrincipalType = &pt
	}
	if !data.PrincipalId.IsNull() {
		pid := data.PrincipalId.ValueString()
		listParams.PrincipalId = &pid
	}
	if !data.ResourceType.IsNull() {
		rt := pinecone.ResourceType(data.ResourceType.ValueString())
		listParams.ResourceType = &rt
	}
	if !data.ResourceId.IsNull() {
		rid := data.ResourceId.ValueString()
		listParams.ResourceId = &rid
	}
	if !data.Role.IsNull() {
		role := data.Role.ValueString()
		listParams.Role = &role
	}

	// Page through all results using the returned pagination token.
	for {
		roleBindings, err := d.adminClient.RoleBinding.List(ctx, listParams)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list role bindings, got error: %s", err))
			return
		}

		for _, rb := range roleBindings.Data {
			data.RoleBindings = append(data.RoleBindings, *models.NewRoleBindingModel(rb))
		}

		if roleBindings.Pagination == nil || roleBindings.Pagination.Next == "" {
			break
		}
		next := roleBindings.Pagination.Next
		listParams.PaginationToken = &next
	}

	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
