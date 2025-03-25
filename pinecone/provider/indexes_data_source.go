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
var _ datasource.DataSource = &IndexesDataSource{}

func NewIndexesDataSource() datasource.DataSource {
	return &IndexesDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// IndexesDataSource defines the data source implementation.
type IndexesDataSource struct {
	*PineconeDatasource
}

func (d *IndexesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_indexes"
}

func (d *IndexesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Indexes data source",

		Attributes: map[string]schema.Attribute{
			"indexes": schema.ListNestedAttribute{
				MarkdownDescription: "List of the indexes in your project",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Index name",
							Computed:            true,
						},
						"dimension": schema.Int32Attribute{
							MarkdownDescription: "Index dimension",
							Computed:            true,
						},
						"metric": schema.StringAttribute{
							MarkdownDescription: "Index metric",
							Computed:            true,
						},
						"host": schema.StringAttribute{
							MarkdownDescription: "The URL address where the index is hosted.",
							Computed:            true,
						},
						"spec": schema.SingleNestedAttribute{
							Description: "Spec",
							Optional:    true,
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"pod": schema.SingleNestedAttribute{
									Description: "Configuration needed to deploy a pod-based index.",
									Optional:    true,
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"environment": schema.StringAttribute{
											MarkdownDescription: "The environment where the index is hosted.",
											Computed:            true,
										},
										"replicas": schema.Int64Attribute{
											MarkdownDescription: "The number of replicas. Replicas duplicate your index. They provide higher availability and throughput. Replicas can be scaled up or down as your needs change.",
											Computed:            true,
										},
										"shards": schema.Int64Attribute{
											MarkdownDescription: "The number of shards. Shards split your data across multiple pods so you can fit more data into an index.",
											Computed:            true,
										},
										"pod_type": schema.StringAttribute{
											MarkdownDescription: "The type of pod to use. One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8.",
											Computed:            true,
										},
										"pods": schema.Int64Attribute{
											MarkdownDescription: "The number of pods to be used in the index. This should be equal to shards x replicas.'",
											Computed:            true,
										},
										"metadata_config": schema.SingleNestedAttribute{
											Description: "Configuration for the behavior of Pinecone's internal metadata index. By default, all metadata is indexed; when metadata_config is present, only specified metadata fields are indexed. These configurations are only valid for use with pod-based indexes.",
											Optional:    true,
											Computed:    true,
											Attributes: map[string]schema.Attribute{
												"indexed": schema.ListAttribute{
													Description: "The indexed fields.",
													Computed:    true,
													ElementType: types.StringType,
												},
											},
										},
										"source_collection": schema.StringAttribute{
											MarkdownDescription: "The name of the collection to create an index from.",
											Computed:            true,
										},
									},
								},
								"serverless": schema.SingleNestedAttribute{
									Description: "Configuration needed to deploy a serverless index.",
									Optional:    true,
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"cloud": schema.StringAttribute{
											Description: "Ready.",
											Computed:    true,
										},
										"region": schema.StringAttribute{
											MarkdownDescription: "Initializing InitializationFailed ScalingUp ScalingDown ScalingUpPodSize ScalingDownPodSize Upgrading Terminating Ready",
											Computed:            true,
										},
									},
								},
							},
						},
						"status": schema.SingleNestedAttribute{
							Description: "Configuration for the behavior of Pinecone's internal metadata index. By default, all metadata is indexed; when metadata_config is present, only specified metadata fields are indexed. To specify metadata fields to index, provide an array of the following form: [example_metadata_field]",
							Optional:    true,
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"ready": schema.BoolAttribute{
									Description: "Ready.",
									Computed:    true,
								},
								"state": schema.StringAttribute{
									MarkdownDescription: "Initializing InitializationFailed ScalingUp ScalingDown ScalingUpPodSize ScalingDownPodSize Upgrading Terminating Ready",
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Indexes identifier",
				Computed:            true,
			},
		},
	}
}

func (d *IndexesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.IndexesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	indexes, err := d.client.ListIndexes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListIndexes, got error: %s", err))
		return
	}

	for _, i := range indexes {
		index := models.IndexModel{}
		resp.Diagnostics.Append(index.Read(ctx, i)...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Indexes = append(data.Indexes, index)
	}

	// For the purposes of this Indexes code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	// tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
