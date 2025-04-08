// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IndexDataSource{}

func NewIndexDataSource() datasource.DataSource {
	return &IndexDataSource{PineconeDatasource: &PineconeDatasource{}}
}

// IndexDataSource defines the data source implementation.
type IndexDataSource struct {
	*PineconeDatasource
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
			"dimension": schema.Int32Attribute{
				MarkdownDescription: "Index dimension",
				Computed:            true,
			},
			"metric": schema.StringAttribute{
				MarkdownDescription: "Index metric",
				Computed:            true,
			},
			"deletion_protection": schema.StringAttribute{
				MarkdownDescription: "Index deletion protection configuration",
				Computed:            true,
			},
			"vector_type": schema.StringAttribute{
				MarkdownDescription: "Index vector type",
				Computed:            true,
			},
			"tags": schema.MapAttribute{
				Description: "Custom user tags added to an index. Keys must be 80 characters or less. Values must be 120 characters or less. Keys must be alphanumeric, '', or '-'. Values must be alphanumeric, ';', '@', '', '-', '.', '+', or ' '. To unset a key, set the value to be an empty string.",
				Computed:    true,
				ElementType: types.StringType,
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
			"embed": schema.SingleNestedAttribute{
				Description: `Specify the integrated inference embedding configuration for the index. Once set, the model cannot be changed. However, you can later update the embedding configuration—including field map, read parameters, and write parameters.

Refer to the [model guide](https://docs.pinecone.io/guides/inference/understanding-inference#embedding-models) for available models and details.`,
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"model": schema.StringAttribute{
						Computed:    true,
						Description: "the name of the embedding model to use for the index.",
					},
					"field_map": schema.MapAttribute{
						Computed:    true,
						Description: "Identifies the name of the text field from your document model that will be embedded.",
						ElementType: types.StringType,
					},
					"metric": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'. If the 'vector_type' is 'sparse', the metric must be 'dotproduct'. If the vector_type is dense, the metric defaults to 'cosine'.",
					},
					"dimension": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The dimension of the embedding model, specifying the size of the output vector.",
					},
					"vector_type": schema.StringAttribute{
						Computed:    true,
						Description: "The index vector type associated with the model. If 'dense', the vector dimension must be specified. If 'sparse', the vector dimension will be nil.",
					},
					"read_parameters": schema.MapAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The read parameters for the embedding model.",
						ElementType: types.StringType,
					},
					"write_parameters": schema.MapAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The write parameters for the embedding model.",
						ElementType: types.StringType,
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
	}
}

func (d *IndexDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.IndexDatasourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	index, err := d.client.DescribeIndex(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe index", err.Error())
		return
	}

	data.Read(ctx, index)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
