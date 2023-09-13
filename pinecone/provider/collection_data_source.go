// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/skyscrapr/pinecone-sdk-go/pinecone"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CollectionDataSource{}

func NewCollectionDataSource() datasource.DataSource {
	return &CollectionDataSource{}
}

// CollectionDataSource defines the data source implementation.
type CollectionDataSource struct {
	client *pinecone.Client
}

// CollectionDataSourceModel describes the data source data model.
type CollectionDataSourceModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Size   types.Int64  `tfsdk:"size"`
	Status types.String `tfsdk:"status"`
}

func (d *CollectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

func (d *CollectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Collection data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Collection identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the collection.",
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the collection in bytes.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the collection.",
				Computed:            true,
			},
		},
	}
}

func (d *CollectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pinecone.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *CollectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CollectionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	collection, err := d.client.Collections().DescribeCollection(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to describe collection, got error: %s", err))
		return
	}

	// Save data into Terraform state
	data.Id = types.StringValue(collection.Name)
	data.Name = types.StringValue(collection.Name)
	data.Size = types.Int64Value(int64(collection.Size))
	data.Status = types.StringValue(collection.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
