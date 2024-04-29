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
	"github.com/pinecone-io/go-pinecone/pinecone"
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
	ApiKey     types.String `tfsdk:"api_key"`
	MgmtApiKey types.String `tfsdk:"mgmt_api_key"`
}

func (p *PineconeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pinecone"
	resp.Version = p.version
}

func (p *PineconeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Pinecone API Key. Can be configured by setting PINECONE_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"mgmt_api_key": schema.StringAttribute{
				MarkdownDescription: "Pinecone Management API Key. Can be configured by setting PINECONE_MGMT_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

type PineconeClients struct {
	Client     *pinecone.Client
	MgmtClient *pinecone.ManagementClient
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

	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey:    apiKey,
		SourceTag: "terraform",
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pinecone client", err.Error())
		return
	}

	mgmtApiKey := os.Getenv("PINECONE_MGMT_API_KEY")
	if !data.MgmtApiKey.IsNull() {
		mgmtApiKey = data.MgmtApiKey.ValueString()
	}

	mgmtClient, err := pinecone.NewManagementClient(pinecone.NewManagementClientParams{
		ApiKey:    mgmtApiKey,
		SourceTag: "terraform",
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pinecone mangement client", err.Error())
		return
	}

	clients := &PineconeClients{
		Client:     client,
		MgmtClient: mgmtClient,
	}

	resp.DataSourceData = clients
	resp.ResourceData = clients
}

func (p *PineconeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCollectionResource,
		NewIndexResource,
		NewProjectResource,
		NewProjectApiKeyResource,
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
		NewProjectApiKeysDataSource,
		NewProjectApiKeyDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PineconeProvider{
			version: version,
		}
	}
}
