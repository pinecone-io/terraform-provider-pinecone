package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

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
var _ resource.ResourceWithConfigValidators = &RoleBindingResource{}

// roleBindingPrincipalTypes are the principal types whose bindings Terraform can
// track for the lifetime of the resource.
//
// "invite" is deliberately absent. When an invitee accepts, the server re-points
// their existing bindings to the new user principal — the same binding id and
// created_at, with principal_type flipped from "invite" to "user". A managed
// invite binding would therefore refresh into a principal that no longer matches
// its config, and because principal_type and principal_id both force replacement,
// the next apply would delete the binding (revoking the user's real role) and try
// to recreate it against an already-processed invite.
//
// Reading invite bindings is unaffected: the pinecone_role_bindings data source
// still accepts principal_type = "invite" as a filter.
var roleBindingPrincipalTypes = []string{
	string(pinecone.PrincipalTypeUser),
	string(pinecone.PrincipalTypeServiceAccount),
	string(pinecone.PrincipalTypeAPIKey),
}

// roleBindingPrincipalTypeValidator restricts principal_type to the trackable
// types. It exists instead of a plain stringvalidator.OneOf so that "invite" —
// which the API accepts and which users will reasonably reach for — gets an
// explanation and a redirect rather than being silently missing from a list.
type roleBindingPrincipalTypeValidator struct{}

func (v roleBindingPrincipalTypeValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Must be one of: %s. \"invite\" is not supported because the server re-points invite bindings to the user principal on acceptance.", strings.Join(roleBindingPrincipalTypes, ", "))
}

func (v roleBindingPrincipalTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v roleBindingPrincipalTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if slices.Contains(roleBindingPrincipalTypes, value) {
		return
	}

	if value == string(pinecone.PrincipalTypeInvite) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Unsupported principal_type",
			"principal_type \"invite\" is not supported by pinecone_role_binding. When the invitee accepts, the server moves this binding to their new user principal, "+
				"which Terraform would see as a changed principal and resolve by deleting the binding — revoking the role it had just granted.\n\n"+
				"Grant roles at invite time with `pinecone_invite.role_bindings`, and manage them with `pinecone_role_binding` using principal_type = \"user\" once the invite is accepted. "+
				"To read an invite's bindings, use the `pinecone_role_bindings` data source, which does accept principal_type = \"invite\".",
		)
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid principal_type",
		fmt.Sprintf("principal_type must be one of: %s. Got: %q.", strings.Join(roleBindingPrincipalTypes, ", "), value),
	)
}

// roleBindingScopeValidator surfaces a resource_id/resource_type mismatch during
// plan instead of waiting for the apply-time check in Create.
type roleBindingScopeValidator struct{}

func (v roleBindingScopeValidator) Description(ctx context.Context) string {
	return "Checks that resource_id is set for project-scoped bindings and omitted for organization-scoped ones."
}

func (v roleBindingScopeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v roleBindingScopeValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data models.RoleBindingResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if summary, detail, ok := roleBindingScopeError(data.ResourceType, data.ResourceId); !ok {
		resp.Diagnostics.AddAttributeError(path.Root("resource_id"), summary, detail)
	}
}

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

func (r *RoleBindingResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		roleBindingScopeValidator{},
	}
}

func (r *RoleBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_role_binding` resource lets you grant a role to a principal (user, service account, or API key) at organization or project scope. Role bindings are immutable: changing any attribute forces the binding to be recreated. Bindings for a pending invite are not managed here — grant those with `pinecone_invite.role_bindings`, which the server moves to the user principal when the invite is accepted. Learn more about roles in the [docs](https://docs.pinecone.io/guides/organizations/understanding-organizations#roles).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Role binding identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"principal_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the principal to grant the role to. A UUID for all principal types (the user, service account, or API key ID).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal_type": schema.StringAttribute{
				MarkdownDescription: "The kind of principal that receives permissions. Valid values are: `user`, `service_account`, `api_key`. `invite` is not accepted here — an invite's bindings move to the user principal when the invite is accepted, so Terraform cannot manage them across that transition. Grant roles at invite time with `pinecone_invite.role_bindings`, then manage them with `principal_type = \"user\"` afterwards.",
				Required:            true,
				Validators: []validator.String{
					roleBindingPrincipalTypeValidator{},
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
				MarkdownDescription: "The ID of the project the binding applies to. Required when `resource_type` is `project`; must be omitted when `resource_type` is `organization` (an organization binding is scoped to the caller's organization automatically). Under organization scope Terraform keeps this null in state to match your config, so it never appears in plan output; the `pinecone_role_binding` data source reports the organization ID instead.",
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

	// Backstop for the plan-time roleBindingScopeValidator, which does not run on
	// every path into Create (an import followed by an apply, for example).
	if summary, detail, ok := roleBindingScopeError(data.ResourceType, data.ResourceId); !ok {
		resp.Diagnostics.AddError(summary, detail)
		return
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
