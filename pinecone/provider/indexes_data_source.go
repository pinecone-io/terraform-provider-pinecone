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
		MarkdownDescription: "The `pinecone_indexes` data source lists the indexes in the project. Each index is reported both ways the `pinecone_index` resource describes indexes: " +
			"with `schema` and `deployment`, and with `dimension`, `metric`, `vector_type`, `spec`, and `embed`. The second set is derived from the first.",

		Attributes: map[string]schema.Attribute{
			"indexes": schema.ListNestedAttribute{
				MarkdownDescription: "The indexes in the project.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: withSharedIndexDSAttributes(map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The index's name.",
							Computed:            true,
						},
						"dimension": schema.Int32Attribute{
							MarkdownDescription: "The dimension of the index's dense vectors. Derived from the index's vector fields in `schema`: null when they don't determine a single value, as for an index with only full-text-search fields.",
							Computed:            true,
						},
						"metric": schema.StringAttribute{
							MarkdownDescription: "The distance metric used for similarity search: `cosine`, `dotproduct`, or `euclidean`. Derived from the index's vector fields in `schema`: null when they don't determine a single value, as for an index with only full-text-search fields.",
							Computed:            true,
						},
						"deletion_protection": schema.StringAttribute{
							MarkdownDescription: "Whether deletion protection is enabled: `enabled` or `disabled`.",
							Computed:            true,
						},
						"tags": schema.MapAttribute{
							MarkdownDescription: "Custom user tags added to an index, at most 20 per index. Keys must be 80 characters or less and contain only letters, digits, `_`, or `-`. Values must be 120 characters or less and consist of printable ASCII characters or spaces.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"vector_type": schema.StringAttribute{
							MarkdownDescription: "The type of vectors the index stores: `dense` or `sparse`. Derived from the index's vector fields in `schema`: null when they don't determine a single value, as for an index with only full-text-search fields.",
							Computed:            true,
						},
						"host": schema.StringAttribute{
							MarkdownDescription: "The URL address where the index is hosted.",
							Computed:            true,
						},
						"spec": schema.SingleNestedAttribute{
							Description: "Where and how the index runs, in the form used by `spec` on the `pinecone_index` resource. The same information is in `deployment`.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"pod": schema.SingleNestedAttribute{
									MarkdownDescription: "A pod-based index.",
									Computed:            true,
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
											MarkdownDescription: "The number of pods the index uses: `shards` × `replicas`.",
											Computed:            true,
										},
										"metadata_config": schema.SingleNestedAttribute{
											Description: "Configuration for the behavior of Pinecone's internal metadata index. The API no longer reports this setting, so `indexed` is always null. Indexed metadata fields are listed in the top-level `schema`.",
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
											MarkdownDescription: "The name of the collection the index was created from, if any.",
											Computed:            true,
										},
									},
								},
								"serverless": schema.SingleNestedAttribute{
									MarkdownDescription: "A serverless index.",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"cloud": schema.StringAttribute{
											Description: "The public cloud where the index is hosted: `aws`, `gcp`, or `azure`.",
											Computed:    true,
										},
										"region": schema.StringAttribute{
											MarkdownDescription: "The region where the index is hosted.",
											Computed:            true,
										},
										"read_capacity": readCapacityDSSchema(),
										"schema":        metadataSchemaComputedSchema(),
									},
								},
								"byoc": schema.SingleNestedAttribute{
									MarkdownDescription: "A BYOC (Bring Your Own Cloud) index.",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"environment": schema.StringAttribute{
											MarkdownDescription: "The BYOC environment where the index is hosted.",
											Computed:            true,
										},
										"read_capacity": readCapacityDSSchema(),
										"schema":        metadataSchemaComputedSchema(),
									},
								},
							},
						},
						"embed": schema.SingleNestedAttribute{
							Description: `The embedding of a legacy integrated index, created with embed. Null for other indexes. Integrated embedding on a document index is reported on its ` + "`schema`" + `'s ` + "`string`" + ` fields.
			
Refer to the [model guide](https://docs.pinecone.io/models/overview) for available models and details.`,
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"model": schema.StringAttribute{
									Computed:    true,
									Description: "The name of the embedding model.",
								},
								"field_map": schema.MapAttribute{
									Computed:            true,
									MarkdownDescription: "Maps `text` to the name of the record field that's embedded.",
									ElementType:         types.StringType,
								},
								"metric": schema.StringAttribute{
									Computed:    true,
									Description: "The distance metric the index uses with the model.",
								},
								"dimension": schema.Int64Attribute{
									Computed:    true,
									Description: "The dimension of the vectors the model produces. Null for models that produce sparse vectors.",
								},
								"vector_type": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The type of vectors the model produces: `dense` or `sparse`.",
								},
								"read_parameters": schema.MapAttribute{
									Computed:    true,
									Description: "The read parameters for the embedding model.",
									ElementType: types.StringType,
								},
								"write_parameters": schema.MapAttribute{
									Computed:    true,
									Description: "The write parameters for the embedding model.",
									ElementType: types.StringType,
								},
							},
						},
						"status": schema.SingleNestedAttribute{
							Description: "The index's status.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"ready": schema.BoolAttribute{
									Description: "Whether the index is ready to use.",
									Computed:    true,
								},
								"state": schema.StringAttribute{
									MarkdownDescription: indexStateDescription,
									Computed:            true,
								},
							},
						},
					}),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source's identifier.",
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
