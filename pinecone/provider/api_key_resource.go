// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

const (
	defaultAPIKeyCreateTimeout time.Duration = 2 * time.Minute
	defaultAPIKeyUpdateTimeout time.Duration = 2 * time.Minute
	defaultAPIKeyDeleteTimeout time.Duration = 2 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &APIKeyResource{}
var _ resource.ResourceWithImportState = &APIKeyResource{}

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{PineconeResource: &PineconeResource{}}
}

// APIKeyResource defines the resource implementation.
type APIKeyResource struct {
	*PineconeResource
}

func (r *APIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_api_key` resource lets you create and manage API keys within Pinecone projects.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "API key identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the API key to be created.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(100),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project where the API key will be created.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"roles": schema.ListAttribute{
				MarkdownDescription: "The roles assigned to the API key.",
				Computed:            true,
				ElementType:         basetypes.StringType{},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The API key value (only available on creation).",
				Computed:            true,
				Sensitive:           true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.APIKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default timeouts
	createTimeout, diags := data.Timeouts.Create(ctx, defaultAPIKeyCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Create the API key using Admin API
	createParams := &pinecone.CreateAPIKeyParams{
		Name: data.Name.ValueString(),
	}

	apiKey, err := r.adminClient.APIKey.Create(ctx, data.ProjectId.ValueString(), createParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	// Read the created API key into the model
	resp.Diagnostics.Append(data.ReadWithSecret(ctx, apiKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.APIKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the API key using Admin API
	apiKey, err := r.adminClient.APIKey.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read API key", err.Error())
		return
	}

	// Read the API key into the model
	resp.Diagnostics.Append(data.Read(ctx, apiKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.APIKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default timeouts
	updateTimeout, diags := data.Timeouts.Update(ctx, defaultAPIKeyUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Update the API key using Admin API
	name := data.Name.ValueString()
	updateParams := &pinecone.UpdateAPIKeyParams{
		Name: &name,
	}

	apiKey, err := r.adminClient.APIKey.Update(ctx, data.Id.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update API key", err.Error())
		return
	}

	// Read the updated API key into the model
	resp.Diagnostics.Append(data.Read(ctx, apiKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.APIKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default timeouts
	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultAPIKeyDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Delete the API key using Admin API
	err := r.adminClient.APIKey.Delete(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete API key", err.Error())
		return
	}
}

func (r *APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
} 