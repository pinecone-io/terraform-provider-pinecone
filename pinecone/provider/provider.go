// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/skyscrapr/pinecone-sdk-go/pinecone"
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
	ApiKey      types.String `tfsdk:"api_key"`
	Environment types.String `tfsdk:"environment"`
}

func (p *PineconeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pinecone"
	resp.Version = p.version
}

func (p *PineconeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Pinecone API Key",
				Optional:            true,
				Sensitive:           true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Pinecone Environment",
				Required:            true,
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
	client := pinecone.NewClient(apiKey, data.Environment.ValueString())

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PineconeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIndexResource,
		NewCollectionResource, // Added the collection resource
	}
}

func (p *PineconeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCollectionsDataSource,
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
