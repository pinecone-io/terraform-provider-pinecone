// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IndexesDataSource{}

func NewIndexDataSource() datasource.DataSource {
	return &IndexDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// IndexDataSource defines the data source implementation.
type IndexDataSource struct {
	*PineconeDatasource
}

// IndexDataSource describes the data source data model.
type IndexDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Dimension types.Int64  `tfsdk:"dimension"`
	Metric    types.String `tfsdk:"metric"`
}

func (d *IndexDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index"
}

func (d *IndexDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Index data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Index identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Index name",
				Required:            true,
			},
			"dimension": schema.Int64Attribute{
				MarkdownDescription: "Index dimension",
				Computed:            true,
			},
			"metric": schema.StringAttribute{
				MarkdownDescription: "Index metric",
				Computed:            true,
			},
		},
	}
}

func (d *IndexDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IndexDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	index, err := d.client.Databases().DescribeIndex(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Index, got error: %s", err))
		return
	}

	data.Name = types.StringValue(index.Database.Name)
	data.Id = types.StringValue(index.Database.Name)
	data.Dimension = types.Int64Value(int64(index.Database.Dimension))
	data.Metric = types.StringValue(index.Database.Metric.String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
