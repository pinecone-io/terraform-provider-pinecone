package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UserDataSource{}
var _ datasource.DataSourceWithConfigValidators = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	*PineconeDatasource
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("email"),
		),
	}
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch a single organization user by ID or email. Exactly one of `id` or `email` must be set.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the user. Set this or `email`. Use it as the `principal_id` when creating a `pinecone_role_binding`.",
				Optional:            true,
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The user's email address (case-insensitive). Set this or `id`.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The user's display name. Null if the user has not set one.",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read users.")
		return
	}

	hasId := !data.Id.IsNull() && !data.Id.IsUnknown()
	hasEmail := !data.Email.IsNull() && !data.Email.IsUnknown()

	switch {
	case hasId == hasEmail:
		// Backstop for the plan-time ExactlyOneOf config validator.
		resp.Diagnostics.AddError("Invalid lookup", "Exactly one of `id` or `email` must be set.")
		return
	case hasId:
		user, err := d.adminClient.User.Describe(ctx, data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to describe user", err.Error())
			return
		}
		data.Read(user)
	case hasEmail:
		email := data.Email.ValueString()
		listParams := &pinecone.ListUsersParams{Email: &email}

		// The list email parameter is a server-side filter whose exactness is not
		// guaranteed, so keep only exact case-insensitive matches to ensure a
		// broadened server match can never resolve to the wrong user. Collect
		// across every page for the same reason: a broadened filter could push
		// the exact match off the first page and turn it into a false "not found".
		matches := make([]*pinecone.User, 0, 1)
		for {
			users, err := d.adminClient.User.List(ctx, listParams)
			if err != nil {
				resp.Diagnostics.AddError("Failed to list users", err.Error())
				return
			}

			for _, u := range users.Data {
				if strings.EqualFold(u.Email, email) {
					matches = append(matches, u)
				}
			}

			if users.Pagination == nil || users.Pagination.Next == "" {
				break
			}
			next := users.Pagination.Next
			listParams.PaginationToken = &next
		}

		switch len(matches) {
		case 0:
			resp.Diagnostics.AddError("User not found", fmt.Sprintf("No user found with email %q.", email))
			return
		case 1:
			data.Read(matches[0])
		default:
			resp.Diagnostics.AddError("Multiple users found", fmt.Sprintf("More than one user matched email %q; look up by id instead.", email))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
