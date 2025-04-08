// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

const (
	defaultIndexCreateTimeout time.Duration = 10 * time.Minute
	defaultIndexUpdateTimeout time.Duration = 10 * time.Minute
	defaultIndexDeleteTimeout time.Duration = 10 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IndexResource{}
var _ resource.ResourceWithImportState = &IndexResource{}

func NewIndexResource() resource.Resource {
	return &IndexResource{PineconeResource: &PineconeResource{}}
}

// IndexResource defines the resource implementation.
type IndexResource struct {
	*PineconeResource
}

func (r *IndexResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index"
}

func (r *IndexResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The `pinecone_index` resource lets you create and manage indexes in Pinecone. Learn more about indexes in the [docs](https://docs.pinecone.io/guides/indexes/understanding-indexes).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Index identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the index to be created. The maximum length is 45 characters.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(45),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dimension": schema.Int32Attribute{
				MarkdownDescription: "The dimensions of the vectors to be inserted in the index",
				Required:            true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"metric": schema.StringAttribute{
				MarkdownDescription: "The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'. If the 'vector_type' is 'sparse', the metric must be 'dotproduct'. If the vector_type is dense, the metric defaults to 'cosine'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("cosine"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"euclidean", "cosine", "dotproduct"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deletion_protection": schema.StringAttribute{
				MarkdownDescription: "Whether deletion protection for the index is enabled. You can use 'enabled', or 'disabled'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("disabled"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"enabled", "disabled"}...),
				},
			},
			"vector_type": schema.StringAttribute{
				MarkdownDescription: "The index vector type. You can use 'dense' or 'sparse'. If 'dense', the vector dimension must be specified. If 'sparse', the vector dimension should not be specified.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dense"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"dense", "sparse"}...),
				},
			},
			"tags": schema.MapAttribute{
				Description: "Custom user tags added to an index. Keys must be 80 characters or less. Values must be 120 characters or less. Keys must be alphanumeric, '', or '-'. Values must be alphanumeric, ';', '@', '', '-', '.', '+', or ' '. To unset a key, set the value to be an empty string.",
				Optional:    true,
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
				Attributes: map[string]schema.Attribute{
					"pod": schema.SingleNestedAttribute{
						Description: "Configuration needed to deploy a pod-based index.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"environment": schema.StringAttribute{
								MarkdownDescription: "The environment where the index is hosted.",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"replicas": schema.Int64Attribute{
								MarkdownDescription: "The number of replicas. Replicas duplicate your index. They provide higher availability and throughput. Replicas can be scaled up or down as your needs change.",
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.RequiresReplace(),
								},
							},
							"shards": schema.Int64Attribute{
								MarkdownDescription: "The number of shards. Shards split your data across multiple pods so you can fit more data into an index.",
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.RequiresReplace(),
								},
							},
							"pod_type": schema.StringAttribute{
								MarkdownDescription: "The type of pod to use. One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8.",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
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
										Required:    true,
										ElementType: types.StringType,
									},
								},
							},
							"source_collection": schema.StringAttribute{
								MarkdownDescription: "The name of the collection to create an index from.",
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"serverless": schema.SingleNestedAttribute{
						Description: "Configuration needed to deploy a serverless index.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"cloud": schema.StringAttribute{
								Description: "The public cloud where you would like your index hosted. [gcp|aws|azure]",
								Required:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"region": schema.StringAttribute{
								MarkdownDescription: "The region where you would like your index to be created.",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"embed": schema.SingleNestedAttribute{
				Description: `Specify the integrated inference embedding configuration for the index. Once set, the model cannot be changed. However, you can later update the embedding configuration—including field map, read parameters, and write parameters.

Refer to the [model guide](https://docs.pinecone.io/guides/inference/understanding-inference#embedding-models) for available models and details.`,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"model": schema.StringAttribute{
						Required:    true,
						Description: "the name of the embedding model to use for the index.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"field_map": schema.MapAttribute{
						Required:    true,
						Description: "Identifies the name of the text field from your document model that will be embedded.",
						ElementType: types.StringType,
					},
					"metric": schema.StringAttribute{
						Optional:    true,
						Description: "The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'. If the 'vector_type' is 'sparse', the metric must be 'dotproduct'. If the vector_type is dense, the metric defaults to 'cosine'.",
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"euclidean", "cosine", "dotproduct"}...),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
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
						Optional:    true,
						Description: "The read parameters for the embedding model.",
						ElementType: types.StringType,
					},
					"write_parameters": schema.MapAttribute{
						Optional:    true,
						Description: "The write parameters for the embedding model.",
						ElementType: types.StringType,
					},
				},
			},
			"status": schema.SingleNestedAttribute{
				Description: "Status",
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
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
					CreateDescription: `Timeout defaults to 5 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
					Delete: true,
					DeleteDescription: `Timeout defaults to 5 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
				},
			),
		},
	}
}

func (r *IndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.IndexResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var spec models.IndexSpecModel
	resp.Diagnostics.Append(data.Spec.As(ctx, &spec, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var embed *models.IndexEmbedModel
	if !data.Embed.IsUnknown() && !data.Embed.IsNull() {
		resp.Diagnostics.Append(data.Embed.As(ctx, &embed, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Extract tags
	tagsMapValue, diags := data.Tags.ToMapValue(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	tagsMap, diags := toStringMap(ctx, tagsMapValue)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	tags := pinecone.IndexTags(tagsMap)

	// Prepare the payload for the API request
	if spec.Pod != nil {
		metric := pinecone.IndexMetric(data.Metric.ValueString())
		deletionProtection := pinecone.DeletionProtection(data.DeletionProtection.ValueString())
		podReq := pinecone.CreatePodIndexRequest{
			Name:               data.Name.ValueString(),
			Dimension:          data.Dimension.ValueInt32(),
			Metric:             &metric,
			DeletionProtection: &deletionProtection,
			Environment:        spec.Pod.Environment.ValueString(),
			PodType:            spec.Pod.PodType.ValueString(),
			Shards:             int32(spec.Pod.ShardCount.ValueInt64()),
			Replicas:           int32(spec.Pod.Replicas.ValueInt64()),
		}

		if tags != nil {
			podReq.Tags = &tags
		}

		if !spec.Pod.SourceCollection.IsUnknown() {
			podReq.SourceCollection = spec.Pod.SourceCollection.ValueStringPointer()
		}

		var metadataConfig *pinecone.PodSpecMetadataConfig
		if !spec.Pod.MetadataConfig.IsUnknown() {
			resp.Diagnostics.Append(spec.Pod.MetadataConfig.As(ctx, &metadataConfig, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		podReq.MetadataConfig = metadataConfig

		_, err := r.client.CreatePodIndex(ctx, &podReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to create pod index", err.Error())
			return
		}
	}

	if spec.Serverless != nil {
		metric := pinecone.IndexMetric(*data.Metric.ValueStringPointer())
		deletionProtection := pinecone.DeletionProtection(data.DeletionProtection.ValueString())

		if embed != nil {
			if spec.Pod != nil || spec.Serverless == nil {
				resp.Diagnostics.AddError("Invalid configuration", "Integrated indexes must have a serverless spec.")
				return
			}

			fieldMap := mapAttrToInterfacePtr(embed.FieldMap)
			if fieldMap == nil {
				resp.Diagnostics.AddError("Invalid configuration", "Integrated indexes must have a field_map")
				return
			}

			indexForModelReq := pinecone.CreateIndexForModelRequest{
				Name:   data.Name.ValueString(),
				Cloud:  pinecone.Cloud(spec.Serverless.Cloud.ValueString()),
				Region: spec.Serverless.Region.ValueString(),
				Embed: pinecone.CreateIndexForModelEmbed{
					Model:           embed.Model.ValueString(),
					FieldMap:        *fieldMap,
					Metric:          &metric,
					ReadParameters:  mapAttrToInterfacePtr(embed.ReadParameters),
					WriteParameters: mapAttrToInterfacePtr(embed.WriteParameters),
				},
				DeletionProtection: &deletionProtection,
			}

			if tags != nil {
				indexForModelReq.Tags = &tags
			}

			_, err := r.client.CreateIndexForModel(ctx, &indexForModelReq)
			if err != nil {
				resp.Diagnostics.AddError("Failed to create integrated serverless index", err.Error())
				return
			}
		}

		serverlessReq := pinecone.CreateServerlessIndexRequest{
			Name:               data.Name.ValueString(),
			Dimension:          data.Dimension.ValueInt32Pointer(),
			Metric:             &metric,
			DeletionProtection: &deletionProtection,
			Cloud:              pinecone.Cloud(spec.Serverless.Cloud.ValueString()),
			Region:             spec.Serverless.Region.ValueString(),
		}

		if tags != nil {
			serverlessReq.Tags = &tags
		}

		if vectorType := data.VectorType.ValueString(); vectorType != "" {
			serverlessReq.VectorType = &vectorType
		}

		_, err := r.client.CreateServerlessIndex(ctx, &serverlessReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to create serverless index", err.Error())
			return
		}
	}

	// Wait for index to be ready
	// Create() is passed a default timeout to use if no value
	// has been supplied in the Terraform configuration.
	createTimeout, diags := data.Timeouts.Create(ctx, defaultIndexCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		index, err := r.client.DescribeIndex(ctx, data.Name.ValueString())

		resp.Diagnostics.Append(data.Read(ctx, index)...)

		// Save current status to state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		if err != nil {
			return retry.NonRetryableError(err)
		}
		if !index.Status.Ready && index.Status.State != "Ready" {
			return retry.RetryableError(fmt.Errorf("index not ready. State: %s", index.Status.State))
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for index to become ready.", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.IndexResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	index, err := r.client.DescribeIndex(ctx, data.Id.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to describe index", err.Error())
		}
		return
	}

	data.Read(ctx, index)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.IndexResourceModel
	var newData models.IndexResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read new data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configureRequest pinecone.ConfigureIndexParams

	// Update DeletionProtection if it has changed
	if data.DeletionProtection != newData.DeletionProtection {
		configureRequest.DeletionProtection = pinecone.DeletionProtection(newData.DeletionProtection.ValueString())
	}

	// Update Tags if they have changed
	newTags, diags := newData.Tags.ToMapValue(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	newTagsMap, diags := toStringMap(ctx, newTags)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	if newTagsMap != nil {
		oldTags, diags := data.Tags.ToMapValue(ctx)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}
		oldTagsMap, diags := toStringMap(ctx, oldTags)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		configureRequest.Tags = mergeTags(oldTagsMap, newTagsMap)
	}

	if configureRequest.DeletionProtection != "" || configureRequest.Embed != nil || configureRequest.Tags != nil {
		_, err := r.client.ConfigureIndex(ctx, data.Name.ValueString(), configureRequest)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update index", err.Error())
			return
		}
	}

	index, err := r.client.DescribeIndex(ctx, newData.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe index", err.Error())
		return
	}

	newData.Read(ctx, index)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}

func (r *IndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.IndexResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIndex(ctx, data.Name.ValueString())
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			resp.Diagnostics.AddError("Failed to delete index", err.Error())
		}
		return
	}

	// Wait for index to be deleted
	// Create() is passed a default timeout to use if no value
	// has been supplied in the Terraform configuration.
	deleteTimeout, diags := data.Timeouts.Create(ctx, defaultIndexDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		index, err := r.client.DescribeIndex(ctx, data.Id.ValueString())
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("index not deleted. State: %s", index.Status.State))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for index to be deleted.", err.Error())
		return
	}
}

func (r *IndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mergeTags(oldTags, newTags map[string]string) map[string]string {
	mergedTags := make(map[string]string)

	for k, newVal := range newTags {
		if oldVal, ok := oldTags[k]; !ok || oldVal != newVal {
			mergedTags[k] = newVal
		}
	}

	for k := range oldTags {
		if _, ok := newTags[k]; !ok {
			mergedTags[k] = ""
		}
	}

	return mergedTags
}

func toStringMap(ctx context.Context, value basetypes.MapValue) (map[string]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var result map[string]string
	diags := value.ElementsAs(ctx, &result, false)

	return result, diags
}

func mapAttrToInterfacePtr(attr types.Map) *map[string]interface{} {
	if attr.IsUnknown() || attr.IsNull() {
		return nil
	}

	raw := make(map[string]interface{}, len(attr.Elements()))
	for k, v := range attr.Elements() {
		if sv, ok := v.(basetypes.StringValue); ok {
			raw[k] = sv.ValueString()
		} else {
			raw[k] = v.String()
		}
	}
	return &raw
}
