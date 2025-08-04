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
	AccessToken  types.String `tfsdk:"access_token"`
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
				MarkdownDescription: "Pinecone OAuth Client ID. Can be configured by setting PINECONE_CLIENT_ID environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Pinecone OAuth Client Secret. Can be configured by setting PINECONE_CLIENT_SECRET environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Pinecone OAuth Access Token. Can be configured by setting PINECONE_ACCESS_TOKEN environment variable.",
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

	// Create regular client for index and collection operations
	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey:    apiKey,
		SourceTag: "terraform",
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pinecone client", err.Error())
		return
	}

	// Create admin client for project and API key operations
	// Note: Admin API requires OAuth credentials, not API key
	// For now, we'll create the admin client only if OAuth credentials are available
	var adminClient *pinecone.AdminClient

	// Check for OAuth credentials from provider config first, then environment variables
	clientId := data.ClientId.ValueString()
	clientSecret := data.ClientSecret.ValueString()
	accessToken := data.AccessToken.ValueString()

	if clientId == "" {
		clientId = os.Getenv("PINECONE_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("PINECONE_CLIENT_SECRET")
	}
	if accessToken == "" {
		accessToken = os.Getenv("PINECONE_ACCESS_TOKEN")
	}

	if clientId != "" && clientSecret != "" {
		sourceTag := "terraform"
		adminClient, err = pinecone.NewAdminClient(pinecone.NewAdminClientParams{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			SourceTag:    &sourceTag,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create pinecone admin client", err.Error())
			return
		}
	} else if accessToken != "" {
		sourceTag := "terraform"
		adminClient, err = pinecone.NewAdminClient(pinecone.NewAdminClientParams{
			AccessToken: accessToken,
			SourceTag:   &sourceTag,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create pinecone admin client", err.Error())
			return
		}
	}

	// Create a combined client structure for resources that might need both
	clientData := map[string]interface{}{
		"client":      client,
		"adminClient": adminClient,
	}

	resp.DataSourceData = client
	resp.ResourceData = clientData
}

func (p *PineconeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCollectionResource,
		NewIndexResource,
		NewProjectResource,
		NewAPIKeyResource,
	}
}

func (p *PineconeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCollectionsDataSource,
		NewCollectionDataSource,
		NewIndexesDataSource,
		NewIndexDataSource,
		NewProjectsDataSource,
		NewProjectDataSource,
		NewAPIKeysDataSource,
		NewAPIKeyDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PineconeProvider{
			version: version,
		}
	}
}
