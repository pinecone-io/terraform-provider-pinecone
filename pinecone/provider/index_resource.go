// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"github.com/pinecone-io/go-pinecone/pinecone"
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
		MarkdownDescription: "Index resource",

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
			},
			"dimension": schema.Int64Attribute{
				MarkdownDescription: "The dimensions of the vectors to be inserted in the index",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1536),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"metric": schema.StringAttribute{
				MarkdownDescription: "The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("cosine"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"euclidean", "cosine", "dotproduct"}...),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The URL address where the index is hosted.",
				Computed:            true,
			},
			"spec": schema.SingleNestedAttribute{
				Description: "Spec",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"pod": schema.SingleNestedAttribute{
						Description: "Configuration needed to deploy a pod-based index.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"environment": schema.StringAttribute{
								MarkdownDescription: "The environment where the index is hosted.",
								Required:            true,
							},
							"replicas": schema.Int64Attribute{
								MarkdownDescription: "The number of replicas. Replicas duplicate your index. They provide higher availability and throughput. Replicas can be scaled up or down as your needs change.",
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
							},
							"shards": schema.Int64Attribute{
								MarkdownDescription: "The number of shards. Shards split your data across multiple pods so you can fit more data into an index.",
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
							},
							"pod_type": schema.StringAttribute{
								MarkdownDescription: "The type of pod to use. One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8.",
								Required:            true,
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
								Computed:            true,
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
							},
							"region": schema.StringAttribute{
								MarkdownDescription: "The region where you would like your index to be created.",
								Required:            true,
							},
						},
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

	// Prepare the payload for the API request
	if spec.Pod != nil {
		podReq := pinecone.CreatePodIndexRequest{
			Name:        data.Name.ValueString(),
			Dimension:   int32(data.Dimension.ValueInt64()),
			Metric:      pinecone.IndexMetric(data.Metric.ValueString()),
			Environment: spec.Pod.Environment.ValueString(),
			PodType:     spec.Pod.PodType.ValueString(),
			Shards:      int32(spec.Pod.ShardCount.ValueInt64()),
			Replicas:    int32(spec.Pod.Replicas.ValueInt64()),
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
		serverlessReq := pinecone.CreateServerlessIndexRequest{
			Name:      data.Name.ValueString(),
			Dimension: int32(data.Dimension.ValueInt64()),
			Metric:    pinecone.IndexMetric(data.Metric.ValueString()),
			Cloud:     pinecone.Cloud(spec.Serverless.Cloud.ValueString()),
			Region:    spec.Serverless.Region.ValueString(),
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
		if !index.Status.Ready {
			return retry.RetryableError(fmt.Errorf("index not ready. State: %s", index.Status.State))
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for index to become ready.", err.Error())
		return
	}

	// resp.Diagnostics.Append(data.Read(ctx, index)...)

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
		resp.Diagnostics.AddError("Failed to describe index", err.Error())
		return
	}

	data.Read(ctx, index)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update not supported.
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
		resp.Diagnostics.AddError("Failed to delete index", err.Error())
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
