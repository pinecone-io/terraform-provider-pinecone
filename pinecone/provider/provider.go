// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
)

// Ensure PineconeProvider satisfies various provider interfaces.
var _ provider.Provider = &PineconeProvider{}

// PineconeProvider defines the provider implementation.
type PineconeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PineconeProviderModel describes the provider data model.
type PineconeProviderModel struct {
	ApiKey       types.String `tfsdk:"api_key"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// PineconeProviderData holds the provider data including both regular and admin clients.
type PineconeProviderData struct {
	Client      *pinecone.Client
	AdminClient *pinecone.AdminClient
}

func (p *PineconeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {

	resp.TypeName = "pinecone"
	resp.Version = p.version
}

func (p *PineconeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `You can use the this Terraform provider to manage resources supported 
by [Pinecone](https://www.pinecone.io/). The provider must be configured with the proper 
credentials before use. You can provide credentials via the PINECONE_API_KEY environment variable.`,

		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Pinecone API Key. Can be configured by setting PINECONE_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Pinecone Client ID for admin operations. Can be configured by setting PINECONE_CLIENT_ID environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Pinecone Client Secret for admin operations. Can be configured by setting PINECONE_CLIENT_SECRET environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *PineconeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PineconeProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default to environment variables, but override
	// with Terraform configuration value if set.
	apiKey := os.Getenv("PINECONE_API_KEY")
	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	}

	clientId := os.Getenv("PINECONE_CLIENT_ID")
	if !data.ClientId.IsNull() {
		clientId = data.ClientId.ValueString()
	}

	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")
	if !data.ClientSecret.IsNull() {
		clientSecret = data.ClientSecret.ValueString()
	}

	// Create provider data structure
	providerData := &PineconeProviderData{}

	// Create regular client only if API key is provided
	if apiKey != "" {
		client, err := pinecone.NewClient(pinecone.NewClientParams{
			ApiKey:    apiKey,
			SourceTag: "terraform",
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create pinecone client", err.Error())
			return
		}
		providerData.Client = client
	}

	// Create admin client only if admin credentials are provided
	if clientId != "" && clientSecret != "" {
		adminClient, err := pinecone.NewAdminClient(pinecone.NewAdminClientParams{
			ClientId:     clientId,
			ClientSecret: clientSecret,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create pinecone admin client", err.Error())
			return
		}
		providerData.AdminClient = adminClient
	}

	// Check if at least one client is available
	if providerData.Client == nil && providerData.AdminClient == nil {
		resp.Diagnostics.AddError("No credentials provided", "Either API key (for regular operations) or client_id/client_secret (for admin operations) must be provided.")
		return
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *PineconeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCollectionResource,
		NewIndexResource,
		NewApiKeyResource,
		NewProjectResource,
	}
}

func (p *PineconeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCollectionsDataSource,
		NewCollectionDataSource,
		NewIndexesDataSource,
		NewIndexDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PineconeProvider{
			version: version,
		}
	}
}
