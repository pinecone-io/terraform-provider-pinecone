package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ServiceAccountDataSource{}

func NewServiceAccountDataSource() datasource.DataSource {
	return &ServiceAccountDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// ServiceAccountDataSource defines the data source implementation.
type ServiceAccountDataSource struct {
	*PineconeDatasource
}

func (d *ServiceAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *ServiceAccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch a single service account by ID. The client secret is never returned.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the service account.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service account.",
				Computed:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth client ID the service account uses to obtain access tokens.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the service account was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp of the service account's most recent metadata update.",
				Computed:            true,
			},
		},
	}
}

func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ServiceAccountDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read service accounts.")
		return
	}

	sa, err := d.adminClient.ServiceAccount.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe service account", err.Error())
		return
	}

	data.Read(sa)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
