// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CollectionsDataSource{}

func NewCollectionsDataSource() datasource.DataSource {
	return &CollectionsDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// CollectionsDataSource defines the data source implementation.
type CollectionsDataSource struct {
	*PineconeDatasource
}

func (d *CollectionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collections"
}

func (d *CollectionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Collections data source",

		Attributes: map[string]schema.Attribute{
			"collections": schema.ListNestedAttribute{
				MarkdownDescription: "List of the collections in your project",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
						// "vector_count": schema.Int64Attribute{
						// 	MarkdownDescription: "The number of records stored in the collection.",
						// 	Computed:            true,
						// },
						"environment": schema.StringAttribute{
							MarkdownDescription: "The environment where the collection is hosted.",
							Computed:            true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Collections identifier",
				Computed:            true,
			},
		},
	}
}

func (d *CollectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.CollectionsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	collections, err := d.client.ListCollections(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListCollections, got error: %s", err))
		return
	}

	for _, c := range collections {
		data.Collections = append(data.Collections, *models.NewCollectionModel(c))
	}

	// Save data into Terraform state
	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
