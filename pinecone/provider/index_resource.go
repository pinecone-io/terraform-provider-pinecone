// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/skyscrapr/pinecone-sdk-go/pinecone"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IndexResource{}
var _ resource.ResourceWithImportState = &IndexResource{}

func NewIndexResource() resource.Resource {
	return &IndexResource{}
}

// IndexResource defines the resource implementation.
type IndexResource struct {
	client *pinecone.Client
}

// IndexResourceModel describes the resource data model.
type IndexResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Dimension        types.Int64  `tfsdk:"dimension"`
	Metric           types.String `tfsdk:"metric"`
	Pods             types.Int64  `tfsdk:"pods"`
	Replicas         types.Int64  `tfsdk:"replicas"`
	PodType          types.String `tfsdk:"pod_type"`
	SourceCollection types.String `tfsdk:"source_collection"`
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
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(512),
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
			"pods": schema.Int64Attribute{
				MarkdownDescription: "The number of pods for the index to use,including replicas.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"replicas": schema.Int64Attribute{
				MarkdownDescription: "The number of replicas. Replicas duplicate your index. They provide higher availability and throughput.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"pod_type": schema.StringAttribute{
				MarkdownDescription: "The type of pod to use. One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("starter"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(starter|(s1|p1|p2)\.(x1|x2|x4|x8))$`),
						"One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8.",
					),
				},
			},
			// metadata_config   object | null    Configuration for the behavior of Pinecone's internal metadata index. By default, all metadata is indexed; when metadata_config is present, only specified metadata fields are indexed. To specify metadata fields to index, provide a JSON object of the following form: {"indexed": ["example_metadata_field"]}
			"source_collection": schema.StringAttribute{
				MarkdownDescription: "The name of the collection to create an index from.",
				Optional:            true,
			},
		},
	}
}

func (r *IndexResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *IndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IndexResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the payload for the API request
	payload := pinecone.CreateIndexParams{
		Name:      data.Name.ValueString(),
		Dimension: int(data.Dimension.ValueInt64()),
		Metric:    pinecone.IndexMetric(data.Metric.ValueString()),
		Pods:      int(data.Pods.ValueInt64()),
		Replicas:  int(data.Replicas.ValueInt64()),
		PodType:   data.PodType.ValueString(),
	}
	if !data.SourceCollection.IsNull() {
		payload.SourceCollection = data.SourceCollection.ValueStringPointer()
	}

	err := r.client.Databases().CreateIndex(&payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create index", err.Error())
		return
	}

	// Wait for index to be ready
	createTimeout := 1 * time.Hour
	err = retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		index, err := r.client.Databases().DescribeIndex(data.Name.ValueString())

		readIndexData(index, &data)
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IndexResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	index, err := r.client.Databases().DescribeIndex(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe index", err.Error())
		return
	}

	readIndexData(index, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IndexResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the payload for the API request
	payload := pinecone.ConfigureIndexParams{
		Name:     data.Name.ValueString(),
		Replicas: int(data.Replicas.ValueInt64()),
		PodType:  data.PodType.ValueString(),
	}

	err := r.client.Databases().ConfigureIndex(&payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update index", err.Error())
		return
	}

	index, err := r.client.Databases().DescribeIndex(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe index", err.Error())
		return
	}

	readIndexData(index, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IndexResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Databases().DeleteIndex(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete index", err.Error())
		return
	}
	// Wait for index to be deleted
	deleteTimeout := 1 * time.Hour
	err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		index, err := r.client.Databases().DescribeIndex(data.Id.ValueString())
		if err != nil {
			if err.Error() == "404 Not Found: Index not found" {
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

func readIndexData(index *pinecone.Index, model *IndexResourceModel) {
	model.Id = types.StringValue(index.Database.Name)
	model.Name = types.StringValue(index.Database.Name)
	model.Dimension = types.Int64Value(int64(index.Database.Dimension))
	model.Metric = types.StringValue(index.Database.Metric.String())
	model.Pods = types.Int64Value(int64(index.Database.Pods))
	model.Replicas = types.Int64Value(int64(index.Database.Replicas))
	model.PodType = types.StringValue(index.Database.PodType)
}
