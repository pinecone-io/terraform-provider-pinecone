// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	pinecone "github.com/nekomeowww/go-pinecone"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IndexesDataSource{}

func NewIndexesDataSource() datasource.DataSource {
	return &IndexesDataSource{}
}

// IndexesDataSource defines the data source implementation.
type IndexesDataSource struct {
	client *pinecone.Client
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

func (d *IndexesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IndexesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IndexesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	indexes, err := d.client.ListIndexes()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Indexes, got error: %s", err))
		return
	}

	data.Indexes = indexes

	// For the purposes of this Indexes code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue("Indexes-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	// tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
