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
var _ datasource.DataSource = &ServiceAccountsDataSource{}

func NewServiceAccountsDataSource() datasource.DataSource {
	return &ServiceAccountsDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// ServiceAccountsDataSource defines the data source implementation.
type ServiceAccountsDataSource struct {
	*PineconeDatasource
}

func (d *ServiceAccountsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

func (d *ServiceAccountsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Service accounts data source. Lists all service accounts in your organization. Client secrets are never returned.",

		Attributes: map[string]schema.Attribute{
			"service_accounts": schema.ListNestedAttribute{
				MarkdownDescription: "The list of service accounts in your organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique ID of the service account.",
							Computed:            true,
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
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Service accounts identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *ServiceAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ServiceAccountsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to list service accounts.")
		return
	}

	listParams := &pinecone.ListServiceAccountsParams{}
	for {
		serviceAccounts, err := d.adminClient.ServiceAccount.List(ctx, listParams)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list service accounts, got error: %s", err))
			return
		}

		for _, sa := range serviceAccounts.Data {
			data.ServiceAccounts = append(data.ServiceAccounts, *models.NewServiceAccountModel(sa))
		}

		if serviceAccounts.Pagination == nil || serviceAccounts.Pagination.Next == "" {
			break
		}
		next := serviceAccounts.Pagination.Next
		listParams.PaginationToken = &next
	}

	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
