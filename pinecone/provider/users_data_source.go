package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// UsersDataSource defines the data source implementation.
type UsersDataSource struct {
	*PineconeDatasource
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Users data source. Lists the members of your organization, optionally filtered by email.",

		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				MarkdownDescription: "Case-insensitive filter on the user's email address.",
				Optional:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "The list of users in your organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique ID of the user.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The user's email address.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The user's display name. Null if the user has not set one.",
							Computed:            true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Users identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.UsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to list users.")
		return
	}

	listParams := &pinecone.ListUsersParams{}
	if !data.Email.IsNull() {
		email := data.Email.ValueString()
		listParams.Email = &email
	}

	for {
		users, err := d.adminClient.User.List(ctx, listParams)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list users, got error: %s", err))
			return
		}

		for _, u := range users.Data {
			data.Users = append(data.Users, *models.NewUserModel(u))
		}

		if users.Pagination == nil || users.Pagination.Next == "" {
			break
		}
		next := users.Pagination.Next
		listParams.PaginationToken = &next
	}

	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
