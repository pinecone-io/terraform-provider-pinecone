// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
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
var _ resource.ResourceWithModifyPlan = &IndexResource{}

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				MarkdownDescription: "The dimensions of the vectors to be inserted in the index. Required for pod-based and non-integrated serverless indexes. For integrated indexes with an embed model, this is optional and will default to the model's dimension if not specified.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
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
					stringplanmodifier.UseStateForUnknown(),
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
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"dense", "sparse"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.MapAttribute{
				Description: "Custom user tags added to an index. Keys must be 80 characters or less. Values must be 120 characters or less. Keys must be alphanumeric, '', or '-'. Values must be alphanumeric, ';', '@', '', '-', '.', '+', or ' '. To unset a key, set the value to be an empty string.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The URL address where the index is hosted.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"spec": schema.SingleNestedAttribute{
				Description: "Spec",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"pod": schema.SingleNestedAttribute{
						MarkdownDescription: "Configuration of an existing pod-based index. New pod-based indexes can't be created: " +
							"Pinecone API version 2026-07 doesn't support it. Existing pod-based indexes can still be imported, " +
							"scaled with `replicas` and `pod_type`, and deleted.",
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"environment": schema.StringAttribute{
								MarkdownDescription: "The environment where the index is hosted.",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"replicas": schema.Int64Attribute{
								MarkdownDescription: "The number of replicas. Replicas duplicate your index. They provide higher availability and throughput. Replicas can be scaled up or down in place.",
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
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
								MarkdownDescription: "The type of pod to use. One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8. " +
									"The pod size can be increased in place, for example from `p1.x1` to `p1.x2`. It can't be decreased, and the pod family can't be changed.",
								Required: true,
							},
							"pods": schema.Int64Attribute{
								MarkdownDescription: "The number of pods to be used in the index. This should be equal to shards x replicas.'",
								Computed:            true,
							},
							"metadata_config": schema.SingleNestedAttribute{
								Description: "Configuration for the behavior of Pinecone's internal metadata index. By default, all metadata is indexed; when metadata_config is present, only specified metadata fields are indexed. These configurations are only valid for use with pod-based indexes. The API no longer reports this setting, so the value recorded in state is kept.",
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
								MarkdownDescription: "The name of the collection the index was created from. Creating an index from a collection is no longer supported.",
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
							"read_capacity": readCapacitySchema(),
							"schema":        metadataSchemaResourceSchema(),
						},
					},
					"byoc": schema.SingleNestedAttribute{
						Description: "Configuration needed to deploy a BYOC (Bring Your Own Cloud) index.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"environment": schema.StringAttribute{
								MarkdownDescription: "The environment identifier for the BYOC index.",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"read_capacity": readCapacitySchema(),
							"schema":        metadataSchemaResourceSchema(),
						},
					},
				},
			},
			"embed": schema.SingleNestedAttribute{
				Description: `Specify the integrated inference embedding configuration for the index. It can only be set when the index is created: ` + "`model`" + ` and ` + "`field_map`" + ` can't be changed afterwards, and ` + "`embed`" + ` can't be added to or removed from an existing index. ` + "`read_parameters`" + ` and ` + "`write_parameters`" + ` can be updated in place.

Refer to the [model guide](https://docs.pinecone.io/guides/inference/understanding-inference#embedding-models) for available models and details.`,
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					embedNullForNullConfig{},
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"model": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "the name of the embedding model to use for the index.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"field_map": schema.MapAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Identifies the name of the text field from your document model that will be embedded.",
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.Map{
							mapplanmodifier.UseStateForUnknown(),
						},
					},
					"metric": schema.StringAttribute{
						Computed:    true,
						Description: "The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'. If the 'vector_type' is 'sparse', the metric must be 'dotproduct'. If the vector_type is dense, the metric defaults to 'cosine'.",
						PlanModifiers: []planmodifier.String{
							embedComputedStringModifier{},
						},
					},
					"dimension": schema.Int32Attribute{
						Computed:    true,
						Description: "The dimension of the embedding model, specifying the size of the output vector.",
						PlanModifiers: []planmodifier.Int32{
							embedComputedInt32Modifier{},
						},
					},
					"vector_type": schema.StringAttribute{
						Computed:    true,
						Description: "The index vector type associated with the model. If 'dense', the vector dimension must be specified. If 'sparse', the vector dimension will be nil.",
						PlanModifiers: []planmodifier.String{
							embedComputedStringModifier{},
						},
					},
					"read_parameters": schema.MapAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The read parameters for the embedding model.",
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.Map{
							embedComputedStringMapModifier{},
						},
					},
					"write_parameters": schema.MapAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The write parameters for the embedding model.",
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.Map{
							embedComputedStringMapModifier{},
						},
					},
					"effective_read_parameters": schema.MapAttribute{
						Computed: true,
						MarkdownDescription: "The effective read parameters as returned by the API after apply, " +
							"including any server-injected defaults not present in `read_parameters`.",
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.Map{
							embedComputedStringMapModifier{},
						},
					},
					"effective_write_parameters": schema.MapAttribute{
						Computed: true,
						MarkdownDescription: "The effective write parameters as returned by the API after apply, " +
							"including any server-injected defaults not present in `write_parameters`.",
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.Map{
							embedComputedStringMapModifier{},
						},
					},
				},
			},
			"status": schema.SingleNestedAttribute{
				Description: "Status",
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
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
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
					CreateDescription: `Timeout defaults to 5 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
					Update: true,
					UpdateDescription: `How long to wait for a pod-based index to finish scaling. Defaults to 10 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
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

	var embed *models.IndexEmbedResourceModel
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

	// Validate that exactly one spec type is provided.
	specCount := 0
	if spec.Pod != nil {
		specCount++
	}
	if spec.Serverless != nil {
		specCount++
	}
	if spec.BYOC != nil {
		specCount++
	}
	if specCount == 0 {
		resp.Diagnostics.AddError("Invalid configuration", "Exactly one of spec.pod, spec.serverless, or spec.byoc must be specified.")
		return
	}
	if specCount > 1 {
		resp.Diagnostics.AddError("Invalid configuration", "Only one of spec.pod, spec.serverless, or spec.byoc may be specified.")
		return
	}

	// Prepare the payload for the API request. ModifyPlan rejects spec.pod on create; this guards
	// against a plan that skipped it.
	if spec.Pod != nil {
		resp.Diagnostics.AddError(podCreateUnsupportedSummary, podCreateUnsupportedDetail)
		return
	}
	if spec.Serverless != nil {
		metric := pinecone.IndexMetric(data.Metric.ValueString())
		deletionProtection := pinecone.DeletionProtection(data.DeletionProtection.ValueString())

		readCapacityParams, diags := models.ToReadCapacityParams(ctx, spec.Serverless.ReadCapacity)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		schemaParams, diags := models.ToMetadataSchema(ctx, spec.Serverless.Schema)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if embed != nil {
			fieldMap := mapAttrToInterfacePtr(embed.FieldMap)

			embedConfig := pinecone.CreateIndexForModelEmbed{
				Model:           embed.Model.ValueString(),
				FieldMap:        *fieldMap,
				Metric:          &metric,
				ReadParameters:  mapAttrToInterfacePtr(embed.ReadParameters),
				WriteParameters: mapAttrToInterfacePtr(embed.WriteParameters),
			}

			// If dimension is specified at the top level, pass it through to the embed config
			// Otherwise, the API will use the model's default dimension
			if !data.Dimension.IsUnknown() && !data.Dimension.IsNull() {
				dimension := int(data.Dimension.ValueInt32())
				embedConfig.Dimension = &dimension
			}

			indexForModelReq := pinecone.CreateIndexForModelRequest{
				Name:               data.Name.ValueString(),
				Cloud:              pinecone.Cloud(spec.Serverless.Cloud.ValueString()),
				Region:             spec.Serverless.Region.ValueString(),
				Embed:              embedConfig,
				DeletionProtection: &deletionProtection,
				ReadCapacity:       readCapacityParams,
				Schema:             schemaParams,
			}

			if tags != nil {
				indexForModelReq.Tags = &tags
			}

			_, err := r.client.CreateIndexForModel(ctx, &indexForModelReq)
			if err != nil {
				resp.Diagnostics.AddError("Failed to create integrated serverless index", err.Error())
				return
			}
		} else {
			serverlessReq := pinecone.CreateServerlessIndexRequest{
				Name:               data.Name.ValueString(),
				Dimension:          data.Dimension.ValueInt32Pointer(),
				Metric:             &metric,
				DeletionProtection: &deletionProtection,
				Cloud:              pinecone.Cloud(spec.Serverless.Cloud.ValueString()),
				Region:             spec.Serverless.Region.ValueString(),
				ReadCapacity:       readCapacityParams,
				Schema:             schemaParams,
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
	} else if spec.BYOC != nil {
		metric := pinecone.IndexMetric(data.Metric.ValueString())
		deletionProtection := pinecone.DeletionProtection(data.DeletionProtection.ValueString())

		if embed != nil {
			resp.Diagnostics.AddError("Invalid configuration", "BYOC indexes cannot have an embed configuration.")
			return
		}

		readCapacityParams, diags := models.ToReadCapacityParams(ctx, spec.BYOC.ReadCapacity)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		byocSchemaParams, diags := models.ToMetadataSchema(ctx, spec.BYOC.Schema)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		byocReq := pinecone.CreateBYOCIndexRequest{
			Name:               data.Name.ValueString(),
			Environment:        spec.BYOC.Environment.ValueString(),
			Dimension:          data.Dimension.ValueInt32Pointer(),
			Metric:             &metric,
			DeletionProtection: &deletionProtection,
			ReadCapacity:       readCapacityParams,
			Schema:             byocSchemaParams,
		}

		if tags != nil {
			byocReq.Tags = &tags
		}

		if vectorType := data.VectorType.ValueString(); vectorType != "" {
			byocReq.VectorType = &vectorType
		}

		_, err := r.client.CreateBYOCIndex(ctx, &byocReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to create BYOC index", err.Error())
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
		if err != nil {
			errStr := err.Error()
			// Retry if the index is not found, otherwise return a non-retryable error
			if strings.Contains(errStr, "not found") ||
				strings.Contains(errStr, "404") ||
				strings.Contains(errStr, "NOT_FOUND") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		resp.Diagnostics.Append(data.Read(ctx, index)...)
		if resp.Diagnostics.HasError() {
			return retry.NonRetryableError(fmt.Errorf("reading index state: %v", resp.Diagnostics))
		}

		// Restore user-configured read_parameters / write_parameters from the plan
		// so state matches the plan exactly. effective_* retains the full API response
		// (set by NewIndexEmbedResourceModel) which may include server defaults like "truncate".
		if embed != nil {
			resp.Diagnostics.Append(restoreEmbedParams(ctx, embed, &data)...)
			if resp.Diagnostics.HasError() {
				return retry.NonRetryableError(fmt.Errorf("restoring embed params: %v", resp.Diagnostics))
			}
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return retry.NonRetryableError(fmt.Errorf("setting state: %v", resp.Diagnostics))
		}

		return indexReadyRetry(index)
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

	// Capture prior embed to restore user-configured read/write parameters after the
	// API read overwrites them. effective_* will reflect the new full API response.
	var priorEmbedModel *models.IndexEmbedResourceModel
	if !data.Embed.IsNull() && !data.Embed.IsUnknown() {
		priorEmbedModel = &models.IndexEmbedResourceModel{}
		if d := data.Embed.As(ctx, priorEmbedModel, basetypes.ObjectAsOptions{}); d.HasError() {
			priorEmbedModel = nil
		}
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

	resp.Diagnostics.Append(data.Read(ctx, index)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if priorEmbedModel != nil {
		resp.Diagnostics.Append(restoreEmbedParams(ctx, priorEmbedModel, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

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

	// Update the embedding model's read and write parameters. ModifyPlan rejects every other embed
	// change on an existing index, since 2026-07 can't change the model or field map in place.
	if !data.Embed.IsNull() && !data.Embed.IsUnknown() && !newData.Embed.IsNull() && !newData.Embed.IsUnknown() {
		var oldEmbed, newEmbed models.IndexEmbedResourceModel
		resp.Diagnostics.Append(data.Embed.As(ctx, &oldEmbed, basetypes.ObjectAsOptions{})...)
		resp.Diagnostics.Append(newData.Embed.As(ctx, &newEmbed, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		readChanged := !newEmbed.ReadParameters.IsUnknown() && !newEmbed.ReadParameters.Equal(oldEmbed.ReadParameters)
		writeChanged := !newEmbed.WriteParameters.IsUnknown() && !newEmbed.WriteParameters.Equal(oldEmbed.WriteParameters)
		if readChanged || writeChanged {
			fieldName, ok := semanticTextFieldName(oldEmbed.FieldMap)
			if !ok {
				resp.Diagnostics.AddError("Failed to update index", "Couldn't determine the embedded text field from embed.field_map.")
				return
			}
			var field pinecone.ConfigureSemanticTextField
			if readChanged {
				field.ReadParameters = mapAttrToInterfacePtr(newEmbed.ReadParameters)
			}
			if writeChanged {
				field.WriteParameters = mapAttrToInterfacePtr(newEmbed.WriteParameters)
			}
			configureRequest.Schema = &pinecone.ConfigureIndexSchema{
				Fields: map[string]pinecone.ConfigureSemanticTextField{fieldName: field},
			}
		}
	}

	// Scale a pod-based index. ModifyPlan rejects pod_type changes the API can't apply.
	oldPod, newPod := extractPodSpec(ctx, data.Spec, &resp.Diagnostics), extractPodSpec(ctx, newData.Spec, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if oldPod != nil && newPod != nil {
		if !newPod.Replicas.IsUnknown() && !newPod.Replicas.Equal(oldPod.Replicas) {
			configureRequest.Replicas = int32(newPod.Replicas.ValueInt64())
		}
		if !newPod.PodType.IsUnknown() && !newPod.PodType.Equal(oldPod.PodType) {
			configureRequest.PodType = newPod.PodType.ValueString()
		}
	}

	// Update ReadCapacity if it has changed (serverless and BYOC only)
	oldReadCapacity, newReadCapacity := extractReadCapacityFromSpec(ctx, data.Spec, &resp.Diagnostics), extractReadCapacityFromSpec(ctx, newData.Spec, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !oldReadCapacity.Equal(newReadCapacity) {
		// Guard: ReadCapacity is not supported for pod indexes
		var currentSpec models.IndexSpecModel
		resp.Diagnostics.Append(data.Spec.As(ctx, &currentSpec, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		if currentSpec.Pod != nil {
			resp.Diagnostics.AddError("Invalid configuration", "ReadCapacity is not supported for pod-based indexes.")
			return
		}

		readCapacityParams, diags := models.ToReadCapacityParams(ctx, newReadCapacity)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		configureRequest.ReadCapacity = readCapacityParams
	}

	// send configure index request if there are things that have been updated
	if configureRequest.DeletionProtection != "" || configureRequest.Schema != nil || configureRequest.Tags != nil ||
		configureRequest.ReadCapacity != nil || configureRequest.PodType != "" || configureRequest.Replicas != 0 {
		_, err := r.client.ConfigureIndex(ctx, data.Name.ValueString(), configureRequest)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update index", err.Error())
			return
		}
	}

	var index *pinecone.Index
	if configureRequest.PodType != "" || configureRequest.Replicas != 0 {
		updateTimeout, diags := newData.Timeouts.Update(ctx, defaultIndexUpdateTimeout)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		var err error
		index, err = r.waitForPodScaling(ctx, newData.Name.ValueString(), newPod, updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed to wait for index to finish scaling.", err.Error())
			return
		}
	} else {
		var err error
		index, err = r.client.DescribeIndex(ctx, newData.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to describe index", err.Error())
			return
		}
	}

	// Capture the plan embed so we can restore user-configured read/write parameters
	// after the API read overwrites them. effective_* will reflect the new full API response.
	var planEmbedModel *models.IndexEmbedResourceModel
	if !newData.Embed.IsNull() && !newData.Embed.IsUnknown() {
		planEmbedModel = &models.IndexEmbedResourceModel{}
		if d := newData.Embed.As(ctx, planEmbedModel, basetypes.ObjectAsOptions{}); d.HasError() {
			planEmbedModel = nil
		}
	}

	resp.Diagnostics.Append(newData.Read(ctx, index)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planEmbedModel != nil {
		resp.Diagnostics.Append(restoreEmbedParams(ctx, planEmbedModel, &newData)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}

func (r *IndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.IndexResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete() is passed a default timeout to use if no value
	// has been supplied in the Terraform configuration.
	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultIndexDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Bound the whole delete (issue + wait-for-gone) to a single delete
	// timeout. Both retry phases share this deadline, so the operation cannot
	// run for up to twice the configured budget.
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Issue the delete, retrying transient server errors (e.g. a 500 blip).
	// A successful call or "not found" ends the retry immediately, so the
	// request is not re-issued once it has taken effect.
	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		err := r.client.DeleteIndex(ctx, data.Name.ValueString())
		if err == nil || strings.Contains(err.Error(), "not found") {
			return nil
		}
		if isTransientError(err) {
			return retry.RetryableError(fmt.Errorf("delete index request failed, retrying: %w", err))
		}
		return retry.NonRetryableError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete index", err.Error())
		return
	}

	err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		index, err := r.client.DescribeIndex(ctx, data.Id.ValueString())
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil
			}
			if isTransientError(err) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("index not deleted. State: %s", indexState(index)))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for index to be deleted.", err.Error())
		return
	}
}

func (r *IndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const (
	podCreateUnsupportedSummary = "Pod-based indexes can't be created"
	podCreateUnsupportedDetail  = "Pinecone API version 2026-07 doesn't support creating pod-based indexes. " +
		"Existing pod-based indexes can still be imported, scaled, and deleted. Use spec.serverless or spec.byoc to create a new index."
	recreateIndexHint = "To recreate the index with the new configuration, taint it first (`terraform taint <address>`). " +
		"Recreating an index deletes all of its data."
)

// indexReplacePaths lists the attributes whose RequiresReplace plan modifier recreates the index.
// The framework doesn't pass attribute-level replacements to resource ModifyPlan, so ModifyPlan
// checks these itself.
var indexReplacePaths = []path.Path{
	path.Root("name"),
	path.Root("dimension"),
	path.Root("metric"),
	path.Root("spec").AtName("pod").AtName("environment"),
	path.Root("spec").AtName("pod").AtName("shards"),
	path.Root("spec").AtName("pod").AtName("source_collection"),
	path.Root("spec").AtName("serverless").AtName("cloud"),
	path.Root("spec").AtName("serverless").AtName("region"),
	path.Root("spec").AtName("serverless").AtName("schema"),
	path.Root("spec").AtName("byoc").AtName("environment"),
	path.Root("spec").AtName("byoc").AtName("schema"),
}

// ModifyPlan rejects configurations that API version 2026-07 can't apply, so they fail at plan time
// instead of partway through an apply.
func (r *IndexResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var config models.IndexResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.State.Raw.IsNull() {
		resp.Diagnostics.Append(validateIndexCreate(ctx, config)...)
		return
	}

	replace, diags := plansIndexReplacement(ctx, req)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if replace {
		resp.Diagnostics.Append(validateIndexCreate(ctx, config)...)
		return
	}

	var state models.IndexResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateIndexUpdate(ctx, config, state)...)
}

// plansIndexReplacement reports whether the plan changes an attribute that recreates the index.
func plansIndexReplacement(ctx context.Context, req resource.ModifyPlanRequest) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	for _, p := range indexReplacePaths {
		var planned, prior attr.Value
		diags.Append(req.Plan.GetAttribute(ctx, p, &planned)...)
		diags.Append(req.State.GetAttribute(ctx, p, &prior)...)
		if diags.HasError() {
			return false, diags
		}
		if planned.IsUnknown() || !planned.Equal(prior) {
			return true, diags
		}
	}
	return false, diags
}

// validateIndexCreate checks a configuration that creates a new index.
func validateIndexCreate(ctx context.Context, config models.IndexResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if config.Spec.IsNull() || config.Spec.IsUnknown() {
		return diags
	}
	var spec models.IndexSpecModel
	diags.Append(config.Spec.As(ctx, &spec, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return diags
	}

	if spec.Pod != nil {
		diags.AddAttributeError(path.Root("spec").AtName("pod"), podCreateUnsupportedSummary, podCreateUnsupportedDetail)
	}
	metadataSchemaDetail := "Metadata fields are indexed automatically when you upsert data, so they no longer need to be declared. " +
		"Remove the schema attribute. It's only accepted together with embed, for integrated indexes."
	if spec.Serverless != nil && !spec.Serverless.Schema.IsNull() && config.Embed.IsNull() {
		diags.AddAttributeError(path.Root("spec").AtName("serverless").AtName("schema"), "Metadata schema isn't supported", metadataSchemaDetail)
	}
	if spec.BYOC != nil && !spec.BYOC.Schema.IsNull() {
		diags.AddAttributeError(path.Root("spec").AtName("byoc").AtName("schema"), "Metadata schema isn't supported", metadataSchemaDetail)
	}
	return diags
}

// validateIndexUpdate checks a configuration that updates an existing index in place.
func validateIndexUpdate(ctx context.Context, config, state models.IndexResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if !config.Embed.IsUnknown() {
		embedPath := path.Root("embed")
		switch {
		case state.Embed.IsNull() && !config.Embed.IsNull():
			diags.AddAttributeError(embedPath, "embed can't be added to an existing index",
				"Integrated embedding can only be configured when an index is created. "+recreateIndexHint)
		case !state.Embed.IsNull() && config.Embed.IsNull():
			diags.AddAttributeError(embedPath, "embed can't be removed from an existing index",
				"An index created with integrated embedding keeps it. Add the embed block back to match the index. "+recreateIndexHint)
		case !state.Embed.IsNull() && !config.Embed.IsNull():
			var configEmbed, stateEmbed models.IndexEmbedResourceModel
			diags.Append(config.Embed.As(ctx, &configEmbed, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
			diags.Append(state.Embed.As(ctx, &stateEmbed, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			if !configEmbed.Model.IsNull() && !configEmbed.Model.IsUnknown() && !configEmbed.Model.Equal(stateEmbed.Model) {
				diags.AddAttributeError(embedPath.AtName("model"), "embed.model can't be changed",
					fmt.Sprintf("The index uses the embedding model %s, which can't be changed after the index is created. %s", stateEmbed.Model, recreateIndexHint))
			}
			if !configEmbed.FieldMap.IsNull() && !configEmbed.FieldMap.IsUnknown() && !configEmbed.FieldMap.Equal(stateEmbed.FieldMap) {
				diags.AddAttributeError(embedPath.AtName("field_map"), "embed.field_map can't be changed",
					"The embedded text field can't be changed after the index is created. "+recreateIndexHint)
			}
		}
	}

	if !config.Spec.IsNull() && !config.Spec.IsUnknown() && !state.Spec.IsNull() {
		var configSpec, stateSpec models.IndexSpecModel
		diags.Append(config.Spec.As(ctx, &configSpec, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
		diags.Append(state.Spec.As(ctx, &stateSpec, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return diags
		}
		if configSpec.Pod != nil && stateSpec.Pod != nil && !configSpec.Pod.PodType.IsUnknown() {
			if detail, ok := podTypeChange(stateSpec.Pod.PodType.ValueString(), configSpec.Pod.PodType.ValueString()); !ok {
				diags.AddAttributeError(path.Root("spec").AtName("pod").AtName("pod_type"), "pod_type can't be changed this way", detail)
			}
		}
	}

	return diags
}

// podSizes orders the pod sizes a pod type can scale between.
var podSizes = map[string]int{"x1": 1, "x2": 2, "x4": 4, "x8": 8}

// podTypeChange reports whether a pod-based index can scale from one pod type to another in place:
// the size can only grow, within the same pod family. Pod types it can't parse are left for the
// API to judge.
func podTypeChange(from, to string) (detail string, ok bool) {
	if from == to {
		return "", true
	}
	fromFamily, fromSize, fromOk := strings.Cut(from, ".")
	toFamily, toSize, toOk := strings.Cut(to, ".")
	if !fromOk || !toOk || podSizes[fromSize] == 0 || podSizes[toSize] == 0 {
		return "", true
	}
	if fromFamily != toFamily {
		return fmt.Sprintf("The pod family can't be changed (from %s to %s). %s", fromFamily, toFamily, recreateIndexHint), false
	}
	if podSizes[toSize] < podSizes[fromSize] {
		return fmt.Sprintf("The pod size can only be increased (from %s to %s). %s", fromSize, toSize, recreateIndexHint), false
	}
	return "", true
}

// semanticTextFieldName returns the name of the schema field an integrated index embeds, which
// ConfigureIndex addresses the embedding model by. It's the "text" entry of the field map, or the
// only entry when the map has one.
func semanticTextFieldName(fieldMap types.Map) (string, bool) {
	if fieldMap.IsNull() || fieldMap.IsUnknown() {
		return "", false
	}
	elements := fieldMap.Elements()
	value, ok := elements["text"]
	if !ok && len(elements) == 1 {
		for _, v := range elements {
			value, ok = v, true
		}
	}
	name, isString := value.(basetypes.StringValue)
	if !ok || !isString || name.IsNull() || name.IsUnknown() || name.ValueString() == "" {
		return "", false
	}
	return name.ValueString(), true
}

// indexState returns the reported state of an index, or "Unknown" when it has no status yet.
func indexState(index *pinecone.Index) string {
	if index == nil || index.Status == nil {
		return "Unknown"
	}
	return string(index.Status.State)
}

// indexReadyRetry classifies a described index for a wait loop: nil once it's ready, a
// non-retryable error once it reaches a state it won't leave on its own, and a retryable error
// otherwise.
func indexReadyRetry(index *pinecone.Index) *retry.RetryError {
	if index.Status == nil {
		return retry.RetryableError(fmt.Errorf("index status not reported yet"))
	}
	switch index.Status.State {
	case pinecone.IndexStatusStateFailed, pinecone.IndexStatusStateInitializationFailed, pinecone.IndexStatusStateDisabled:
		return retry.NonRetryableError(fmt.Errorf("index entered state %s", index.Status.State))
	}
	if !index.Status.Ready && index.Status.State != pinecone.IndexStatusStateReady {
		return retry.RetryableError(fmt.Errorf("index not ready. State: %s", index.Status.State))
	}
	return nil
}

// waitForPodScaling waits until a pod-based index is ready and reports the target pod type and
// replica count, and returns its final description.
func (r *IndexResource) waitForPodScaling(ctx context.Context, name string, target *models.IndexPodSpecModel, timeout time.Duration) (*pinecone.Index, error) {
	var index *pinecone.Index
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		described, err := r.client.DescribeIndex(ctx, name)
		if err != nil {
			if isTransientError(err) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		index = described
		if retryErr := indexReadyRetry(described); retryErr != nil {
			return retryErr
		}
		if !podDeploymentMatches(described, target) {
			return retry.RetryableError(fmt.Errorf("index still scaling to %s with %d replicas", target.PodType.ValueString(), target.Replicas.ValueInt64()))
		}
		return nil
	})
	return index, err
}

// podDeploymentMatches reports whether a pod-based index reports the target pod type and replicas.
func podDeploymentMatches(index *pinecone.Index, target *models.IndexPodSpecModel) bool {
	if index.Deployment == nil || index.Deployment.Pod == nil {
		return false
	}
	pod := index.Deployment.Pod
	replicas := int64(1)
	if pod.Replicas != nil {
		replicas = int64(*pod.Replicas)
	}
	return pod.PodType == target.PodType.ValueString() && replicas == target.Replicas.ValueInt64()
}

// extractPodSpec pulls the pod spec out of a spec object, returning nil if absent.
func extractPodSpec(ctx context.Context, specObj types.Object, diagnostics *diag.Diagnostics) *models.IndexPodSpecModel {
	if specObj.IsNull() || specObj.IsUnknown() {
		return nil
	}
	var spec models.IndexSpecModel
	diagnostics.Append(specObj.As(ctx, &spec, basetypes.ObjectAsOptions{})...)
	return spec.Pod
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

// extractReadCapacityFromSpec pulls the read_capacity object out of a spec object
// (checking serverless and byoc sub-specs), returning types.ObjectNull if absent.
func extractReadCapacityFromSpec(ctx context.Context, specObj types.Object, diagnostics *diag.Diagnostics) types.Object {
	if specObj.IsNull() || specObj.IsUnknown() {
		return types.ObjectNull(models.IndexReadCapacityResourceModel{}.AttrTypes())
	}
	var spec models.IndexSpecModel
	diagnostics.Append(specObj.As(ctx, &spec, basetypes.ObjectAsOptions{})...)
	if diagnostics.HasError() {
		return types.ObjectNull(models.IndexReadCapacityResourceModel{}.AttrTypes())
	}
	if spec.Serverless != nil {
		return spec.Serverless.ReadCapacity
	}
	if spec.BYOC != nil {
		return spec.BYOC.ReadCapacity
	}
	return types.ObjectNull(models.IndexReadCapacityResourceModel{}.AttrTypes())
}

// readCapacitySchema returns the schema for the read_capacity block,
// shared between spec.serverless and spec.byoc.
//
// Only user-configurable fields are exposed. Status-only fields (state, current_replicas,
// current_shards, error_message) are intentionally omitted: they cannot be configured,
// and their asynchronous, nullable nature makes them difficult to represent in the desired-state
// model without custom plan modifiers. Use the index data source to observe them.
func readCapacitySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Read capacity configuration for the index. Set exactly one of `dedicated` or `on_demand` to select the mode. " +
			"Omitting `read_capacity` entirely on create defaults to OnDemand. " +
			"To switch modes after creation, explicitly set the desired sub-block — removing `read_capacity` from config will not change the mode already recorded in state.",
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"dedicated": schema.SingleNestedAttribute{
				MarkdownDescription: "Dedicated read capacity mode. Set `node_type`, `replicas`, and `shards` to provision fixed compute for this index. " +
					"All three fields are required when first switching to dedicated mode.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"node_type": schema.StringAttribute{
						MarkdownDescription: "The type of machines to use. Available options: 'b1' and 't1'.",
						Optional:            true,
						Computed:            true,
					},
					"replicas": schema.Int32Attribute{
						MarkdownDescription: "The desired number of replicas.",
						Optional:            true,
						Computed:            true,
					},
					"shards": schema.Int32Attribute{
						MarkdownDescription: "The desired number of shards.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"on_demand": schema.SingleNestedAttribute{
				MarkdownDescription: "OnDemand read capacity mode (the default). Specify this block (even empty) to explicitly select OnDemand or to switch back from dedicated mode.",
				Optional:            true,
				Attributes:          map[string]schema.Attribute{},
			},
		},
	}
}

func metadataSchemaResourceSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Schema for the behavior of Pinecone's internal metadata index. " +
			"By default, all metadata is indexed; when `schema` is present, only fields listed in `fields` " +
			"with `filterable: true` are indexed. This field can only be set at index creation time — " +
			"changing it requires replacing the index. New indexes accept it only together with `embed`; " +
			"other indexes index metadata automatically when you upsert data.",
		DeprecationMessage: "Metadata fields are indexed automatically when you upsert data, so they no longer need to be declared. " +
			"This attribute is kept for existing indexes and for integrated indexes created with embed.",
		Optional: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"fields": schema.MapNestedAttribute{
				MarkdownDescription: "Map of metadata field names to their schema configuration. " +
					"Only fields with `filterable: true` are indexed.",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"filterable": schema.BoolAttribute{
							Description: "Whether the field is filterable. Only true is currently supported.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// restoreEmbedParams writes the user-configured read_parameters and write_parameters
// from refEmbed back into model.Embed after a data.Read() call. effective_read_parameters
// and effective_write_parameters retain the full API response already populated by
// NewIndexEmbedResourceModel, exposing server-injected defaults (e.g. "truncate") without
// causing Terraform's plan-consistency check to fail.
//
// Only non-null, non-unknown values are restored. A null value means the attribute was
// not configured by the user (the plan held null because embedComputedStringMapModifier
// set it to unknown, which UseStateForUnknown then didn't copy from a null prior state).
// In that case the API-populated value is kept in state, which keeps import consistent
// with regular applies. An unknown value means the attribute is still being computed and
// must not be written to state.
func restoreEmbedParams(ctx context.Context, refEmbed *models.IndexEmbedResourceModel, model *models.IndexResourceModel) diag.Diagnostics {
	if refEmbed == nil || model.Embed.IsNull() || model.Embed.IsUnknown() {
		return nil
	}
	var embedState models.IndexEmbedResourceModel
	diags := model.Embed.As(ctx, &embedState, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags
	}
	if !refEmbed.ReadParameters.IsNull() && !refEmbed.ReadParameters.IsUnknown() {
		embedState.ReadParameters = refEmbed.ReadParameters
	}
	if !refEmbed.WriteParameters.IsNull() && !refEmbed.WriteParameters.IsUnknown() {
		embedState.WriteParameters = refEmbed.WriteParameters
	}
	var d diag.Diagnostics
	model.Embed, d = types.ObjectValueFrom(ctx, models.IndexEmbedResourceModel{}.AttrTypes(), embedState)
	diags.Append(d...)
	return diags
}

// embedComputedStringModifier is a plan modifier for Computed-only string attributes
// inside the embed nested object. In terraform-plugin-framework ≥1.15, when the parent
// embed transitions from null prior state to configured (e.g. upgrading a plain index to
// an integrated-inference index), Computed-only children are initialised as null in the
// plan instead of unknown. UseStateForUnknown does not correct this because it only fires
// when the plan value is already unknown; null is not unknown, so it no-ops and the plan
// value stays null. Terraform then rejects the apply result because the API populates
// these fields (null → "cosine" triggers "inconsistent result after apply"). This modifier
// fixes that by always returning unknown when prior state is null, allowing the provider to
// return any value during apply. When state is non-null it behaves identically to
// UseStateForUnknown, preserving the current value in the plan for no-change updates.
type embedComputedStringModifier struct{}

func (embedComputedStringModifier) Description(_ context.Context) string {
	return "Returns unknown when prior state is null; otherwise copies state for unknown."
}
func (embedComputedStringModifier) MarkdownDescription(_ context.Context) string {
	return "Returns unknown when prior state is null; otherwise copies state for unknown."
}
func (embedComputedStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		resp.PlanValue = types.StringUnknown()
		return
	}
	if resp.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}

// embedComputedInt32Modifier is the Int32 equivalent of embedComputedStringModifier.
type embedComputedInt32Modifier struct{}

func (embedComputedInt32Modifier) Description(_ context.Context) string {
	return "Returns unknown when prior state is null; otherwise copies state for unknown."
}
func (embedComputedInt32Modifier) MarkdownDescription(_ context.Context) string {
	return "Returns unknown when prior state is null; otherwise copies state for unknown."
}
func (embedComputedInt32Modifier) PlanModifyInt32(_ context.Context, req planmodifier.Int32Request, resp *planmodifier.Int32Response) {
	if req.StateValue.IsNull() {
		resp.PlanValue = types.Int32Unknown()
		return
	}
	if resp.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}

// embedComputedStringMapModifier is the Map[string]string equivalent of embedComputedStringModifier.
// For Optional+Computed map attributes (read_parameters, write_parameters) it also
// preserves explicit config values so a user-provided map is not overwritten.
type embedComputedStringMapModifier struct{}

func (embedComputedStringMapModifier) Description(_ context.Context) string {
	return "Returns unknown when prior state is null; otherwise copies state for unknown."
}
func (embedComputedStringMapModifier) MarkdownDescription(_ context.Context) string {
	return "Returns unknown when prior state is null; otherwise copies state for unknown."
}
func (embedComputedStringMapModifier) PlanModifyMap(_ context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// If the user supplied an explicit config value, leave it untouched.
	if !resp.PlanValue.IsNull() && !resp.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		resp.PlanValue = types.MapUnknown(types.StringType)
		return
	}
	if resp.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}

// embedNullForNullConfig is a plan modifier for the embed SingleNestedAttribute.
//
// When embed is Optional+Computed and the user hasn't configured it (config is null),
// the Terraform Framework generates an unknown plan value because the block contains
// Computed children. UseStateForUnknown cannot suppress this for non-integrated indexes
// because it no-ops when state is null. This modifier explicitly sets plan = null when
// config is null, preventing spurious "(known after apply)" diffs.
type embedNullForNullConfig struct{}

func (embedNullForNullConfig) Description(_ context.Context) string {
	return "Sets planned embed to null when the user has not configured it."
}

func (embedNullForNullConfig) MarkdownDescription(_ context.Context) string {
	return "Sets planned embed to null when the user has not configured it."
}

func (embedNullForNullConfig) PlanModifyObject(_ context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if req.ConfigValue.IsNull() && req.PlanValue.IsUnknown() {
		resp.PlanValue = req.ConfigValue
	}
}

// isTransientError reports whether err looks like a temporary, retryable
// failure from the Pinecone API or the network rather than a permanent one.
// It is intentionally string-based: the SDK surfaces API errors as formatted
// strings (e.g. `{"status_code":500,"body":"Internal Error. Please try again later."}`)
// rather than typed errors.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	transientMarkers := []string{
		`"status_code":5`, // any 5xx from the API
		"Internal Error",
		"try again",
		"timeout",
		"connection reset",
		"EOF",
	}
	for _, marker := range transientMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
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
