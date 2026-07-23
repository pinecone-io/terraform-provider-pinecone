package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RoleBindingResource{}
var _ resource.ResourceWithImportState = &RoleBindingResource{}

func NewRoleBindingResource() resource.Resource {
	return &RoleBindingResource{PineconeResource: &PineconeResource{}}
}

// RoleBindingResource defines the resource implementation.
type RoleBindingResource struct {
	*PineconeResource
}

func (r *RoleBindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_binding"
}

func (r *RoleBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_role_binding` resource lets you grant a role to a principal (user, service account, API key, or invite) at organization or project scope. Role bindings are immutable: changing any attribute forces the binding to be recreated. Learn more about roles in the [docs](https://docs.pinecone.io/guides/organizations/understanding-organizations#roles).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Role binding identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"principal_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the principal to grant the role to. A UUID for all principal types (the user, service account, API key, or invite ID).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal_type": schema.StringAttribute{
				MarkdownDescription: "The kind of principal that receives permissions. Valid values are: `user`, `service_account`, `api_key`, `invite`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(pinecone.PrincipalTypeUser),
						string(pinecone.PrincipalTypeServiceAccount),
						string(pinecone.PrincipalTypeAPIKey),
						string(pinecone.PrincipalTypeInvite),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "The kind of resource scope the binding applies to. Valid values are: `organization`, `project`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(pinecone.ResourceTypeOrganization),
						string(pinecone.ResourceTypeProject),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project the binding applies to. Required when `resource_type` is `project`; must be omitted when `resource_type` is `organization` (an organization binding is scoped to the caller's organization automatically).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role to assign to the principal at the resource scope. Organization-scoped values: `OrgOwner`, `OrgManager`, `OrgBillingAdmin`, `OrgMember`. Project-scoped values: `ProjectOwner`, `ProjectManager`, `ProjectMember`, `ProjectEditor`, `ProjectViewer`, `ControlPlaneEditor`, `ControlPlaneViewer`, `DataPlaneEditor`, `DataPlaneViewer`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the role binding was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *RoleBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.RoleBindingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to create role bindings.")
		return
	}

	resourceType := data.ResourceType.ValueString()
	hasResourceId := !data.ResourceId.IsNull() && !data.ResourceId.IsUnknown()

	// Validate resource_id against resource_type. For project scope it is
	// required; for organization scope it must be omitted (the server assigns
	// the organization ID, which we surface as a computed value).
	switch resourceType {
	case string(pinecone.ResourceTypeProject):
		if !hasResourceId {
			resp.Diagnostics.AddError("Missing resource_id", "resource_id is required when resource_type is \"project\".")
			return
		}
	case string(pinecone.ResourceTypeOrganization):
		if hasResourceId {
			resp.Diagnostics.AddError("Unexpected resource_id", "resource_id must be omitted when resource_type is \"organization\".")
			return
		}
	}

	createParams := &pinecone.CreateRoleBindingParams{
		PrincipalId:   data.PrincipalId.ValueString(),
		PrincipalType: pinecone.PrincipalType(data.PrincipalType.ValueString()),
		ResourceType:  pinecone.ResourceType(resourceType),
		Role:          data.Role.ValueString(),
	}
	if hasResourceId {
		resourceId := data.ResourceId.ValueString()
		createParams.ResourceId = &resourceId
	}

	roleBinding, err := r.adminClient.RoleBinding.Create(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create role binding", err.Error())
		return
	}

	models.SetRoleBindingResourceModel(&data, roleBinding)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.RoleBindingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read role bindings.")
		return
	}

	roleBinding, err := r.adminClient.RoleBinding.Describe(ctx, data.Id.ValueString())
	if err != nil {
		if isNotFoundErr(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to describe role binding", err.Error())
		}
		return
	}

	models.SetRoleBindingResourceModel(&data, roleBinding)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op. Every configurable attribute forces replacement, so a plan
// never produces an in-place update for this resource. The method exists only to
// satisfy the resource.Resource interface.
func (r *RoleBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.RoleBindingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.RoleBindingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to delete role bindings.")
		return
	}

	err := r.adminClient.RoleBinding.Delete(ctx, data.Id.ValueString())
	if err != nil {
		if isNotFoundErr(err) {
			return
		}
		// The API refuses some deletes (e.g. the last OrgOwner binding, or the
		// last org-membership binding for a user/invite that has other bindings)
		// with an HTTP 409 whose message is already specific and actionable, so
		// surface it directly rather than guessing at the cause.
		resp.Diagnostics.AddError("Failed to delete role binding", err.Error())
		return
	}

	// Deletion is asynchronous (HTTP 202); wait until the role binding is gone.
	err = retryDeletion(ctx, func() error {
		_, err := r.adminClient.RoleBinding.Describe(ctx, data.Id.ValueString())
		return err
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for role binding to be deleted.", err.Error())
		return
	}
}

func (r *RoleBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: role_binding_id
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
