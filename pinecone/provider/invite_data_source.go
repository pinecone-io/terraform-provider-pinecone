package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &InviteDataSource{}

func NewInviteDataSource() datasource.DataSource {
	return &InviteDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// InviteDataSource defines the data source implementation.
type InviteDataSource struct {
	*PineconeDatasource
}

func (d *InviteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (d *InviteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch a single invite by ID. Unlike the `pinecone_invites` list, which returns only `pending` and `expired` invites, this returns an invite in any state — including an accepted (`processed`) one. The granted roles are not returned; read them with the `pinecone_role_bindings` data source using `principal_type = \"invite\"`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the invite.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address the invite was sent to.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The lifecycle status of the invite: `pending`, `expired`, or `processed`.",
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
	}
}

func (d *InviteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.InviteDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read invites.")
		return
	}

	invite, err := d.adminClient.Invite.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe invite", err.Error())
		return
	}

	data.Read(invite)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
