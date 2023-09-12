// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/skyscrapr/pinecone-sdk-go/pinecone"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CollectionResource{}
var _ resource.ResourceWithImportState = &CollectionResource{}

func NewCollectionResource() resource.Resource {
	return &CollectionResource{}
}

// CollectionResource defines the resource implementation.
type CollectionResource struct {
	client *pinecone.Client
}

// CollectionResourceModel describes the resource data model.
type CollectionResourceModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Source types.String `tfsdk:"source"`
	Size   types.Int64  `tfsdk:"size"`
	Status types.String `tfsdk:"status"`
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
	}
}

func (r *CollectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pinecone.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *pinecone.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collection, err := r.client.Collections().DescribeCollection(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe collection", err.Error())
		return
	}

	data.Id = types.StringValue(collection.Name)
	data.Name = types.StringValue(collection.Name)
	data.Size = types.Int64Value(int64(collection.Size))
	data.Status = types.StringValue(collection.Status)

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
}

func (r *CollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
