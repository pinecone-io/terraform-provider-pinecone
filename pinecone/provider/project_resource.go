// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

const (
	defaultProjectCreateTimeout time.Duration = 5 * time.Minute
	defaultProjectUpdateTimeout time.Duration = 5 * time.Minute
	defaultProjectDeleteTimeout time.Duration = 5 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{PineconeResource: &PineconeResource{}}
}

// ProjectResource defines the resource implementation.
type ProjectResource struct {
	*PineconeResource
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_project` resource lets you create and manage projects in Pinecone. Projects are organizational units that contain indexes and collections.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project to be created.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(100),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID that the project belongs to.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the project was created.",
				Computed:            true,
			},
			"force_encryption_with_cmek": schema.BoolAttribute{
				MarkdownDescription: "Whether to force encryption with a customer-managed encryption key (CMEK).",
				Computed:            true,
			},
			"max_pods": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of Pods that can be created in the project.",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default timeouts
	createTimeout, diags := data.Timeouts.Create(ctx, defaultProjectCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Create the project using Admin API
	createParams := &pinecone.CreateProjectParams{
		Name: data.Name.ValueString(),
	}

	project, err := r.adminClient.Project.Create(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create project", err.Error())
		return
	}

	// Read the created project into the model
	resp.Diagnostics.Append(data.Read(ctx, project)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the project using Admin API
	project, err := r.adminClient.Project.Describe(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read project", err.Error())
		return
	}

	// Read the project into the model
	resp.Diagnostics.Append(data.Read(ctx, project)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default timeouts
	updateTimeout, diags := data.Timeouts.Update(ctx, defaultProjectUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Update the project using Admin API
	updateParams := &pinecone.UpdateProjectParams{}

	project, err := r.adminClient.Project.Update(ctx, data.Id.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update project", err.Error())
		return
	}

	// Read the updated project into the model
	resp.Diagnostics.Append(data.Read(ctx, project)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default timeouts
	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultProjectDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Delete the project using Admin API
	err := r.adminClient.Project.Delete(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete project", err.Error())
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
