// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"maps"

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
				MarkdownDescription: "Index metric can be one of 'cosine', 'dotproduct', or 'euclidean'.",
				Computed:            true,
			},
			"deletion_protection": schema.StringAttribute{
				MarkdownDescription: "Index deletion protection can be one of 'enabled' or 'disabled'.",
				Computed:            true,
			},
			"vector_type": schema.StringAttribute{
				MarkdownDescription: "Index vector type, for example 'dense' or 'sparse'.",
				Computed:            true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Custom user tags added to an index, at most 20 per index. Keys must be 80 characters or less and contain only letters, digits, `_`, or `-`. Values must be 120 characters or less and consist of printable ASCII characters or spaces.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The URL address where the index is hosted.",
				Computed:            true,
			},
			"spec": schema.SingleNestedAttribute{
				Description: "Where and how the index runs, in the form used by `spec` on the `pinecone_index` resource. The same information is in `deployment`.",
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
								Description: "Configuration for the behavior of Pinecone's internal metadata index. The API no longer reports this setting, so `indexed` is always null. Indexed metadata fields are listed in the top-level `schema`.",
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
								MarkdownDescription: "The name of the collection the index was created from, if any.",
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
								Description: "The public cloud where the index is hosted.",
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
						Description: "Configuration for a BYOC (Bring Your Own Cloud) index.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"environment": schema.StringAttribute{
								MarkdownDescription: "The environment identifier for the BYOC index.",
								Computed:            true,
							},
							"read_capacity": readCapacityDSSchema(),
							"schema":        metadataSchemaComputedSchema(),
						},
					},
				},
			},
			"embed": schema.SingleNestedAttribute{
				Description: `Specify the integrated inference embedding configuration for the index. The model and field map are fixed when the index is created; the read and write parameters can be updated.

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
						Computed:    true,
						Description: "The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'. If the 'vector_type' is 'sparse', the metric must be 'dotproduct'. If the vector_type is dense, the metric defaults to 'cosine'.",
					},
					"dimension": schema.Int64Attribute{
						Computed:    true,
						Description: "The dimension of the embedding model, specifying the size of the output vector.",
					},
					"vector_type": schema.StringAttribute{
						Computed:    true,
						Description: "The index vector type associated with the model. If 'dense', the vector dimension must be specified. If 'sparse', the vector dimension will be nil.",
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
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"ready": schema.BoolAttribute{
						Description: "Ready.",
						Computed:    true,
					},
					"state": schema.StringAttribute{
						MarkdownDescription: "Initializing InitializationFailed ScalingUp ScalingDown ScalingUpPodSize Terminating Ready Failed Disabled",
						Computed:            true,
					},
				},
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, sharedIndexDSAttributes())
}

// withSharedIndexDSAttributes adds sharedIndexDSAttributes to attributes and returns them.
func withSharedIndexDSAttributes(attributes map[string]schema.Attribute) map[string]schema.Attribute {
	maps.Copy(attributes, sharedIndexDSAttributes())
	return attributes
}

// sharedIndexDSAttributes returns index attributes declared once for both the index and indexes
// data sources.
func sharedIndexDSAttributes() map[string]schema.Attribute {
	description := func(what string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: what, Computed: true}
	}
	filterable := schema.BoolAttribute{
		MarkdownDescription: "Whether the field is indexed for metadata filtering.",
		Computed:            true,
	}
	fieldDescription := description("The field's description, if one was set.")
	metadataField := func(what string) schema.SingleNestedAttribute {
		return schema.SingleNestedAttribute{
			MarkdownDescription: what + " The API adds these fields as data is upserted.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"filterable":  filterable,
				"description": fieldDescription,
			},
		}
	}

	return map[string]schema.Attribute{
		"schema": schema.SingleNestedAttribute{
			MarkdownDescription: "The index's schema: the typed fields its records can contain.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"fields": schema.MapNestedAttribute{
					MarkdownDescription: "The schema's fields, keyed by field name. Exactly one attribute of each field is set, naming its type. " +
						"Vector indexes report their vectors as `_values` (dense) and `_sparse_values` (sparse); indexes that store dense vectors report both.",
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"dense_vector": schema.SingleNestedAttribute{
								MarkdownDescription: "A dense vector field.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"dimension":   schema.Int32Attribute{MarkdownDescription: "The number of dimensions in the field's vectors.", Computed: true},
									"metric":      description("The distance metric used for similarity search: `cosine`, `dotproduct`, or `euclidean`."),
									"description": fieldDescription,
								},
							},
							"sparse_vector": schema.SingleNestedAttribute{
								MarkdownDescription: "A sparse vector field.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"description": fieldDescription,
								},
							},
							"semantic_text": schema.SingleNestedAttribute{
								MarkdownDescription: "A text field embedded by an integrated embedding model, as on an index created with `embed`.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"model":     description("The embedding model."),
									"dimension": schema.Int32Attribute{MarkdownDescription: "The dimension of the vectors the model produces. Null for models that produce sparse vectors.", Computed: true},
									"metric":    description("The distance metric used for similarity search."),
									"read_parameters": schema.MapAttribute{
										MarkdownDescription: "The model parameters applied at query time.",
										Computed:            true,
										ElementType:         types.StringType,
									},
									"write_parameters": schema.MapAttribute{
										MarkdownDescription: "The model parameters applied at write time.",
										Computed:            true,
										ElementType:         types.StringType,
									},
									"description": fieldDescription,
								},
							},
							"string": schema.SingleNestedAttribute{
								MarkdownDescription: "A string field, either declared for full-text search or added by the API as data is upserted.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"full_text_search": schema.SingleNestedAttribute{
										MarkdownDescription: "The field's full-text search configuration. Null unless the field was declared for full-text search.",
										Computed:            true,
										Attributes: map[string]schema.Attribute{
											"language":   description("The language used for text analysis."),
											"stemming":   schema.BoolAttribute{MarkdownDescription: "Whether words are reduced to their root form.", Computed: true},
											"stop_words": schema.BoolAttribute{MarkdownDescription: "Whether common words such as \"the\" are filtered out.", Computed: true},
											"ngram": schema.SingleNestedAttribute{
												MarkdownDescription: "Character n-gram tokenization for substring or prefix matching.",
												Computed:            true,
												Attributes: map[string]schema.Attribute{
													"min_gram":    schema.Int64Attribute{MarkdownDescription: "The minimum n-gram length.", Computed: true},
													"max_gram":    schema.Int64Attribute{MarkdownDescription: "The maximum n-gram length.", Computed: true},
													"prefix_only": schema.BoolAttribute{MarkdownDescription: "Whether only n-grams anchored at the start of each token are generated.", Computed: true},
												},
											},
										},
									},
									"filterable":  filterable,
									"description": fieldDescription,
								},
							},
							"string_list": metadataField("A string list metadata field."),
							"boolean":     metadataField("A boolean metadata field."),
							"float":       metadataField("A floating-point metadata field."),
							"integer":     metadataField("An integer metadata field."),
							"legacy_metadata": schema.SingleNestedAttribute{
								MarkdownDescription: "A metadata field on an index created with a metadata schema before API version 2026-07.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"filterable": filterable,
								},
							},
						},
					},
				},
			},
		},
		"deployment": schema.SingleNestedAttribute{
			MarkdownDescription: "Where the index runs. Exactly one of `managed`, `pod`, or `byoc` is set.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"managed": schema.SingleNestedAttribute{
					MarkdownDescription: "A serverless index.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"cloud":       description("The public cloud where the index is hosted."),
						"region":      description("The region where the index is hosted."),
						"environment": description("The Pinecone environment hosting the index."),
					},
				},
				"pod": schema.SingleNestedAttribute{
					MarkdownDescription: "A pod-based index.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"environment": description("The environment where the index is hosted."),
						"pod_type":    description("The pod type, such as `p1.x1`."),
						"replicas":    schema.Int32Attribute{MarkdownDescription: "The number of replicas.", Computed: true},
						"shards":      schema.Int32Attribute{MarkdownDescription: "The number of shards.", Computed: true},
					},
				},
				"byoc": schema.SingleNestedAttribute{
					MarkdownDescription: "A BYOC (Bring Your Own Cloud) index.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"environment": description("The BYOC environment where the index is hosted."),
					},
				},
			},
		},
		"read_capacity":     readCapacityDSSchema(),
		"private_host":      description("The private endpoint URL of the index, if any."),
		"cmek_id":           description("The ID of the customer-managed encryption key used to encrypt the index, if any."),
		"source_backup_id":  description("The ID of the backup the index was restored from, if any."),
		"source_collection": description("The name of the collection the index was created from, if any."),
	}
}

// readCapacityDSSchema returns the computed-only read_capacity schema for data sources.
func readCapacityDSSchema() schema.Attribute {
	statusAttrs := map[string]schema.Attribute{
		"state": schema.StringAttribute{
			MarkdownDescription: "The overall status of the read capacity configuration.",
			Computed:            true,
		},
		"current_replicas": schema.Int32Attribute{
			MarkdownDescription: "The current number of replicas.",
			Computed:            true,
		},
		"current_shards": schema.Int32Attribute{
			MarkdownDescription: "The current number of shards.",
			Computed:            true,
		},
		"error_message": schema.StringAttribute{
			MarkdownDescription: "An optional error message if there are issues with the read capacity configuration.",
			Computed:            true,
		},
	}

	dedicatedAttrs := map[string]schema.Attribute{
		"node_type": schema.StringAttribute{
			MarkdownDescription: "The type of machines in use.",
			Computed:            true,
		},
		"replicas": schema.Int32Attribute{
			MarkdownDescription: "The desired number of replicas.",
			Computed:            true,
		},
		"shards": schema.Int32Attribute{
			MarkdownDescription: "The desired number of shards.",
			Computed:            true,
		},
	}
	for k, v := range statusAttrs {
		dedicatedAttrs[k] = v
	}

	return schema.SingleNestedAttribute{
		MarkdownDescription: "Read capacity configuration for the index.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"dedicated": schema.SingleNestedAttribute{
				MarkdownDescription: "Dedicated read capacity configuration.",
				Computed:            true,
				Attributes:          dedicatedAttrs,
			},
			"on_demand": schema.SingleNestedAttribute{
				MarkdownDescription: "OnDemand read capacity configuration.",
				Computed:            true,
				Attributes:          statusAttrs,
			},
		},
	}
}

func metadataSchemaComputedSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Schema for the behavior of Pinecone's internal metadata index. " +
			"When present, only fields listed in `fields` with `filterable: true` are indexed.",
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"fields": schema.MapNestedAttribute{
				MarkdownDescription: "Map of metadata field names to their schema configuration.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"filterable": schema.BoolAttribute{
							Description: "Whether the field is filterable.",
							Computed:    true,
						},
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

	resp.Diagnostics.Append(data.Read(ctx, index)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
