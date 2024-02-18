// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/skyscrapr/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CollectionDataSource{}

func NewCollectionDataSource() datasource.DataSource {
	return &CollectionDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// CollectionDataSource defines the data source implementation.
type CollectionDataSource struct {
	*PineconeDatasource
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
			"dimension": schema.Int64Attribute{
				MarkdownDescription: "The dimension of the vectors stored in each record held in the collection.",
				Computed:            true,
			},
			"vector_count": schema.Int64Attribute{
				MarkdownDescription: "The number of records stored in the collection.",
				Computed:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The environment where the collection is hosted.",
				Computed:            true,
			},
		},
	}
}

func (d *CollectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.CollectionDataSourceModel

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
	data.Read(collection)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
