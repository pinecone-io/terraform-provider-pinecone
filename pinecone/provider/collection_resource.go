// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/skyscrapr/pinecone-sdk-go/pinecone"
)

const (
	defaultCollectionCreateTimeout time.Duration = 2 * time.Minute
	defaultCollectionDeleteTimeout time.Duration = 2 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CollectionResource{}
var _ resource.ResourceWithImportState = &CollectionResource{}

func NewCollectionResource() resource.Resource {
	return &CollectionResource{}
}

// CollectionResource defines the resource implementation.
type CollectionResource struct {
	*PineconeResource
}

// CollectionResourceModel describes the resource data model.
type CollectionResourceModel struct {
	Id       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	Source   types.String   `tfsdk:"source"`
	Size     types.Int64    `tfsdk:"size"`
	Status   types.String   `tfsdk:"status"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *CollectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

func (r *CollectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Collection resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Collection identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the collection.",
				Required:            true,
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The name of the source index to be used as the source for the collection.",
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the collection in bytes.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the collection.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
					CreateDescription: `Timeout defaults to 2 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
					Delete: true,
					DeleteDescription: `Timeout defaults to 2 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
				},
			),
		},
	}
}

func (r *CollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := pinecone.CreateCollectionParams{
		Name:   data.Name.ValueString(),
		Source: data.Source.ValueString(),
	}

	err := r.client.Collections().CreateCollection(&payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create collection", err.Error())
		return
	}

	// Wait for collection to be ready
	// Create() is passed a default timeout to use if no value
	// has been supplied in the Terraform configuration.
	createTimeout, diags := data.Timeouts.Create(ctx, defaultCollectionCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		collection, err := r.client.Collections().DescribeCollection(data.Name.ValueString())

		readCollectionData(collection, &data)
		// Save current status to state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		if err != nil {
			return retry.NonRetryableError(err)
		}
		if collection.Status != "Ready" {
			return retry.RetryableError(fmt.Errorf("collection not ready. State: %s", collection.Status))
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for collection to become ready.", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collection, err := r.client.Collections().DescribeCollection(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe collection", err.Error())
		return
	}

	readCollectionData(collection, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Collections currently do not support updates
}

func (r *CollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Collections().DeleteCollection(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete collection", err.Error())
		return
	}
	// Wait for collection to be deleted
	// Create() is passed a default timeout to use if no value
	// has been supplied in the Terraform configuration.
	deleteTimeout, diags := data.Timeouts.Create(ctx, defaultIndexDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		collection, err := r.client.Collections().DescribeCollection(data.Id.ValueString())
		tflog.Info(ctx, fmt.Sprintf("Deleting Collection. Status: '%s'", collection.Status))

		if err != nil {
			if pineconeErr, ok := err.(*pinecone.HTTPError); ok && pineconeErr.StatusCode == 404 {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("collection not deleted. State: %s", collection.Status))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for collection to be deleted.", err.Error())
		return
	}
}

func (r *CollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readCollectionData(collection *pinecone.Collection, model *CollectionResourceModel) {
	model.Id = types.StringValue(collection.Name)
	model.Name = types.StringValue(collection.Name)
	model.Source = types.StringValue(model.Source.ValueString())
	model.Size = types.Int64Value(int64(collection.Size))
	model.Status = types.StringValue(collection.Status)
}
