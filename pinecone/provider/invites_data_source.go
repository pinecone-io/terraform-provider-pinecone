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
var _ datasource.DataSource = &InvitesDataSource{}

func NewInvitesDataSource() datasource.DataSource {
	return &InvitesDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// InvitesDataSource defines the data source implementation.
type InvitesDataSource struct {
	*PineconeDatasource
}

func (d *InvitesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invites"
}

func (d *InvitesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invites data source. Lists the organization's outstanding invites. Only `pending` and `expired` invites are returned; accepted (`processed`) invites are not listed.",

		Attributes: map[string]schema.Attribute{
			"invites": schema.ListNestedAttribute{
				MarkdownDescription: "The list of pending and expired invites in your organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique ID of the invite.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The email address the invite was sent to.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The lifecycle status of the invite: `pending` or `expired`.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "The timestamp when the invite was created.",
							Computed:            true,
						},
						"expires_at": schema.StringAttribute{
							MarkdownDescription: "The timestamp when the invite expires if not accepted.",
							Computed:            true,
						},
						"processed_at": schema.StringAttribute{
							MarkdownDescription: "The timestamp when the invite was accepted. Null for pending or expired invites.",
							Computed:            true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Invites identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *InvitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.InvitesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to list invites.")
		return
	}

	listParams := &pinecone.ListInvitesParams{}
	for {
		invites, err := d.adminClient.Invite.List(ctx, listParams)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list invites, got error: %s", err))
			return
		}

		for _, inv := range invites.Data {
			data.Invites = append(data.Invites, *models.NewInviteModel(inv))
		}

		if invites.Pagination == nil || invites.Pagination.Next == "" {
			break
		}
		next := invites.Pagination.Next
		listParams.PaginationToken = &next
	}

	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
