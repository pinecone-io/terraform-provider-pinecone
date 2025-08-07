package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApiKeyResource{}
var _ resource.ResourceWithImportState = &ApiKeyResource{}

func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{PineconeResource: &PineconeResource{}}
}

// ApiKeyResource defines the resource implementation.
type ApiKeyResource struct {
	*PineconeResource
}

func (r *ApiKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *ApiKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The `pinecone_api_key` resource lets you create and manage API keys in Pinecone. Learn more about API keys in the [docs](https://docs.pinecone.io/guides/authentication/api-keys).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "API key identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the API key to be created. Must be 1-80 characters long.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 80),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The project ID where the API key will be created.",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The generated API key value.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"roles": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The roles assigned to the API key. Valid values are: ProjectEditor, ProjectViewer, ControlPlaneEditor, ControlPlaneViewer, DataPlaneEditor, DataPlaneViewer. Defaults to [\"ProjectEditor\"].",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ApiKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if admin client is available
	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to create API keys.")
		return
	}

	// Prepare create parameters
	createParams := &pinecone.CreateAPIKeyParams{
		Name: data.Name.ValueString(),
	}

	// Handle roles field
	if !data.Roles.IsNull() && !data.Roles.IsUnknown() {
		var roles []string
		resp.Diagnostics.Append(data.Roles.ElementsAs(ctx, &roles, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createParams.Roles = &roles
	}

	// Create the API key
	apiKeyWithSecret, err := r.adminClient.APIKey.Create(ctx, data.ProjectId.ValueString(), createParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	// Set the computed values
	data.Id = types.StringValue(apiKeyWithSecret.Key.Id)
	data.Key = types.StringValue(apiKeyWithSecret.Value)

	// Convert roles from []string to types.Set
	rolesSet, _ := types.SetValueFrom(ctx, types.StringType, apiKeyWithSecret.Key.Roles)
	data.Roles = rolesSet

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ApiKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if admin client is available
	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read API keys.")
		return
	}

	// List API keys to find the one with matching ID
	apiKeys, err := r.adminClient.APIKey.List(ctx, data.ProjectId.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to list API keys", err.Error())
		}
		return
	}

	// Find the API key with matching ID
	var foundApiKey *pinecone.APIKey
	for _, key := range apiKeys {
		if key.Id == data.Id.ValueString() {
			foundApiKey = key
			break
		}
	}

	if foundApiKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the model with the found API key
	data.Name = types.StringValue(foundApiKey.Name)
	// Note: The API key value is not returned in the list operation for security reasons
	// So we keep the existing key value from state

	// Convert roles from []string to types.Set
	rolesSet, _ := types.SetValueFrom(ctx, types.StringType, foundApiKey.Roles)
	data.Roles = rolesSet

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ApiKeyResourceModel
	var state models.ApiKeyResourceModel

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
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to update API keys.")
		return
	}

	// Prepare update parameters
	updateParams := &pinecone.UpdateAPIKeyParams{}

	// Check if name has changed
	if !data.Name.Equal(state.Name) {
		name := data.Name.ValueString()
		updateParams.Name = &name
	}

	// Check if roles have changed
	if !data.Roles.Equal(state.Roles) {
		if !data.Roles.IsNull() && !data.Roles.IsUnknown() {
			var roles []string
			resp.Diagnostics.Append(data.Roles.ElementsAs(ctx, &roles, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			updateParams.Roles = &roles
		} else {
			// If roles is null/unknown, set to empty slice
			emptyRoles := []string{}
			updateParams.Roles = &emptyRoles
		}
	}

	// Only update if there are changes
	if updateParams.Name == nil && updateParams.Roles == nil {
		// No changes, just save the current state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Update the API key
	updatedApiKey, err := r.adminClient.APIKey.Update(ctx, data.Id.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update API key", err.Error())
		return
	}

	// Update the model with the updated API key
	data.Name = types.StringValue(updatedApiKey.Name)

	// Convert roles from []string to types.Set
	rolesSet, _ := types.SetValueFrom(ctx, types.StringType, updatedApiKey.Roles)
	data.Roles = rolesSet

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ApiKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if admin client is available
	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to delete API keys.")
		return
	}

	// Delete the API key
	err := r.adminClient.APIKey.Delete(ctx, data.Id.ValueString())
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			resp.Diagnostics.AddError("Failed to delete API key", err.Error())
		}
		return
	}

	// Wait for API key to be deleted
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		// List API keys to check if the key still exists
		apiKeys, err := r.adminClient.APIKey.List(ctx, data.ProjectId.ValueString())
		if err != nil {
			return retry.NonRetryableError(err)
		}

		// Check if the API key still exists
		for _, key := range apiKeys {
			if key.Id == data.Id.ValueString() {
				return retry.RetryableError(fmt.Errorf("API key not deleted yet"))
			}
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for API key to be deleted.", err.Error())
		return
	}
}

func (r *ApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: project_id:api_key_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import format", "Expected format: project_id:api_key_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
