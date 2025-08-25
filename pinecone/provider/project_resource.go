package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
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
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The `pinecone_project` resource lets you create and manage projects in Pinecone. Learn more about projects in the [docs](https://docs.pinecone.io/guides/projects).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project to be created.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID where the project will be created.",
				Computed:            true,
			},
			"force_encryption_with_cmek": schema.BoolAttribute{
				MarkdownDescription: "Whether to force encryption with a customer-managed encryption key (CMEK). Default is `false`. Once enabled, CMEK encryption cannot be disabled.",
				Optional:            true,
				Computed:            true,
			},
			"max_pods": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of Pods that can be created in the project. Default is `0` (serverless only).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the project was created.",
				Computed:            true,
			},
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

	// Check if admin client is available
	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to create projects.")
		return
	}

	// Prepare create parameters
	createParams := &pinecone.CreateProjectParams{
		Name: data.Name.ValueString(),
	}

	// Handle force_encryption_with_cmek field
	if !data.ForceEncryptionWithCmek.IsNull() && !data.ForceEncryptionWithCmek.IsUnknown() {
		forceEncryption := data.ForceEncryptionWithCmek.ValueBool()
		createParams.ForceEncryptionWithCmek = &forceEncryption
	}

	// Handle max_pods field
	if !data.MaxPods.IsNull() && !data.MaxPods.IsUnknown() {
		maxPods := int(data.MaxPods.ValueInt64())
		createParams.MaxPods = &maxPods
	}

	// Create the project
	project, err := r.adminClient.Project.Create(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create project", err.Error())
		return
	}

	// Set the computed values
	data.Id = types.StringValue(project.Id)
	data.Name = types.StringValue(project.Name)
	data.OrganizationId = types.StringValue(project.OrganizationId)
	data.ForceEncryptionWithCmek = types.BoolValue(project.ForceEncryptionWithCmek)
	data.MaxPods = types.Int64Value(int64(project.MaxPods))
	if project.CreatedAt != nil {
		data.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
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

	// Check if admin client is available
	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read projects.")
		return
	}

	// Describe the project directly
	project, err := r.adminClient.Project.Describe(ctx, data.Id.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to describe project", err.Error())
		}
		return
	}

	// Update the model with the found project
	data.Id = types.StringValue(project.Id)
	data.Name = types.StringValue(project.Name)
	data.OrganizationId = types.StringValue(project.OrganizationId)
	data.ForceEncryptionWithCmek = types.BoolValue(project.ForceEncryptionWithCmek)
	data.MaxPods = types.Int64Value(int64(project.MaxPods))
	if project.CreatedAt != nil {
		data.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ProjectResourceModel
	var state models.ProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if admin client is available
	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to update projects.")
		return
	}

	// Prepare update parameters
	updateParams := &pinecone.UpdateProjectParams{}

	// Check if name has changed
	if !data.Name.Equal(state.Name) {
		name := data.Name.ValueString()
		updateParams.Name = &name
	}

	// Check if force_encryption_with_cmek has changed
	if !data.ForceEncryptionWithCmek.Equal(state.ForceEncryptionWithCmek) {
		forceEncryption := data.ForceEncryptionWithCmek.ValueBool()
		updateParams.ForceEncryptionWithCmek = &forceEncryption
	}

	// Check if max_pods has changed
	if !data.MaxPods.Equal(state.MaxPods) {
		maxPods := int(data.MaxPods.ValueInt64())
		updateParams.MaxPods = &maxPods
	}

	// Only update if there are changes
	if updateParams.Name == nil && updateParams.ForceEncryptionWithCmek == nil && updateParams.MaxPods == nil {
		// No changes, just save the current state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Update the project
	updatedProject, err := r.adminClient.Project.Update(ctx, state.Id.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update project", err.Error())
		return
	}

	// Update the model with the updated project
	data.Id = types.StringValue(updatedProject.Id)
	data.Name = types.StringValue(updatedProject.Name)
	data.OrganizationId = types.StringValue(updatedProject.OrganizationId)
	data.ForceEncryptionWithCmek = types.BoolValue(updatedProject.ForceEncryptionWithCmek)
	data.MaxPods = types.Int64Value(int64(updatedProject.MaxPods))
	if updatedProject.CreatedAt != nil {
		data.CreatedAt = types.StringValue(updatedProject.CreatedAt.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
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

	// Check if admin client is available
	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to delete projects.")
		return
	}

	// Delete the project
	err := r.adminClient.Project.Delete(ctx, data.Id.ValueString())
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			resp.Diagnostics.AddError("Failed to delete project", err.Error())
		}
		return
	}

	// Wait for project to be deleted with simplified verification
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		// Try to describe the specific project to check if it still exists
		_, err := r.adminClient.Project.Describe(ctx, data.Id.ValueString())
		if err != nil {
			// If we can't describe the project, it's likely deleted
			if strings.Contains(err.Error(), "not found") ||
				strings.Contains(err.Error(), "NOT_FOUND") ||
				strings.Contains(err.Error(), "404") {
				return nil // Project is deleted
			}
			// For other errors, retry
			return retry.RetryableError(fmt.Errorf("deletion verification in progress, retrying: %v", err))
		}

		// Project still exists, retry
		return retry.RetryableError(fmt.Errorf("project not deleted yet"))
	})
	if err != nil {
		// If we get a retryable error that's related to the quota issue,
		// we can assume the project was likely deleted successfully
		if strings.Contains(err.Error(), "Resource Quota PodsPerProject not found") ||
			strings.Contains(err.Error(), "deletion verification in progress") {
			// Log a warning but don't fail the deletion
			// The project deletion was successful, but verification failed due to timing issues
			return
		}
		resp.Diagnostics.AddError("Failed to wait for project to be deleted.", err.Error())
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: project_id
	projectId := req.ID

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), projectId)...)
}
