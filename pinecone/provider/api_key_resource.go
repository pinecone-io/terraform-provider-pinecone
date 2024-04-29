// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ProjectApiKeyResource{}

// var _ resource.ResourceWithImportState = &ProjectApiKeyResource{}

func NewProjectApiKeyResource() resource.Resource {
	return &ProjectApiKeyResource{PineconeResource: &PineconeResource{}}
}

// ProjectApiKeyResource defines the resource implementation.
type ProjectApiKeyResource struct {
	*PineconeResource
}

func (r *ProjectApiKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_api_key"
}

func (r *ProjectApiKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Project ApiKey resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ApiKey identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the api key to be created.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(7),
				},
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "The api key secret.",
				Computed:            true,
				Sensitive:           true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The id of the project.",
				Required:            true,
			},
		},
	}
}

func (r *ProjectApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ProjectApiKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId, _ := uuid.Parse(data.ProjectId.ValueString())
	apiKey, err := r.mgmtClient.CreateApiKey(ctx, projectId, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create project api key", err.Error())
		return
	}

	resp.Diagnostics.Append(data.Read(ctx, apiKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ProjectApiKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKeyId, _ := uuid.Parse(data.Id.ValueString())
	apiKey, err := r.mgmtClient.FetchApiKey(ctx, apiKeyId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch project api key", err.Error())
		return
	}

	data.ReadWithoutSecret(ctx, apiKey)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// var data models.ProjectApiKeyModel
}

func (r *ProjectApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ProjectApiKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	apiKeyId, _ := uuid.Parse(data.Id.ValueString())
	err := r.mgmtClient.DeleteApiKey(ctx, apiKeyId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete project api key", err.Error())
		return
	}
}

// func (r *ProjectApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
// }
