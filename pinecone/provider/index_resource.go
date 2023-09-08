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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	pinecone "github.com/nekomeowww/go-pinecone"
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
	Id      types.String `tfsdk:"id"`
	Name	types.String `tfsdk:"name"`
	Dimension types.Int64 `tfsdk:"dimension"`
	Metric types.String `tfsdk:"metric"`
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
				MarkdownDescription: "Index name",
				Required: true,
			},
			"dimension": schema.Int64Attribute{
				MarkdownDescription: "Index dimension",
				Required: true,
			},
			"metric": schema.StringAttribute{
				MarkdownDescription: "Index metric",
				Required: true,
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
		// Metric: mo.Option["cosine"],
	}

	err := r.client.CreateIndex(ctx, payload)

	if err != nil {
		// Handle the error, maybe set a diagnostic in the response
		resp.Diagnostics.AddError("Failed to create index", err.Error())
		return
	}

	// Decode the API response
	// var data map[string]interface{}
	// err = json.NewDecoder(response.Body).Decode(&data)
	// if err != nil {
	// 	// Handle the error, maybe set a diagnostic in the response
	// 	resp.Diagnostics.AddError("Failed to decode API response", err.Error())
	// 	return
	// }

	data.Id = data.Name

	tflog.Trace(ctx, "created an index")

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

	index, err := r.client.DescribeIndex(ctx, data.Name.ValueString())

	if err != nil {
		// Handle the error, maybe set a diagnostic in the response
		resp.Diagnostics.AddError("Failed to create index", err.Error())
		return
	}

	data.Id = data.Name
	data.Name = types.StringValue(index.Database.Name)
	data.Dimension = types.Int64Value(int64(index.Database.Dimension))

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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Index, got error: %s", err))
	//     return
	// }

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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Index, got error: %s", err))
	//     return
	// }
}

func (r *IndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
