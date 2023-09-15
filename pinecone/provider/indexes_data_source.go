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
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IndexesDataSource{}

func NewIndexesDataSource() datasource.DataSource {
	return &IndexesDataSource{}
}

// IndexesDataSource defines the data source implementation.
type IndexesDataSource struct {
	*PineconeDatasource
}

// IndexesDataSourceModel describes the data source data model.
type IndexesDataSourceModel struct {
	Indexes []string     `tfsdk:"indexes"`
	Id      types.String `tfsdk:"id"`
}

func (d *IndexesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_indexes"
}

func (d *IndexesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Indexes data source",

		Attributes: map[string]schema.Attribute{
			"indexes": schema.ListAttribute{
				MarkdownDescription: "Indexes",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Indexes identifier",
				Computed:            true,
			},
		},
	}
}

func (d *IndexesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IndexesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	indexes, err := d.client.Databases().ListIndexes()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListIndexes, got error: %s", err))
		return
	}

	data.Indexes = indexes

	// For the purposes of this Indexes code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	// tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
