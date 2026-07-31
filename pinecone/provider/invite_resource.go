package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// emailRegex is a deliberately permissive sanity check that catches obvious
// mistakes (missing "@" or domain) before a round trip. The API performs
// authoritative validation.
var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InviteResource{}
var _ resource.ResourceWithImportState = &InviteResource{}
var _ resource.ResourceWithConfigValidators = &InviteResource{}

// inviteRoleBindingScopeValidator surfaces a resource_id/resource_type mismatch
// in any nested role binding during plan instead of waiting for the apply-time
// check in Create.
type inviteRoleBindingScopeValidator struct{}

func (v inviteRoleBindingScopeValidator) Description(ctx context.Context) string {
	return "Checks that each role binding sets resource_id for project scope and omits it for organization scope."
}

func (v inviteRoleBindingScopeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v inviteRoleBindingScopeValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	roleBindingsPath := path.Root("role_bindings")

	var roleBindings types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, roleBindingsPath, &roleBindings)...)
	if resp.Diagnostics.HasError() || roleBindings.IsNull() || roleBindings.IsUnknown() {
		return
	}

	// Convert one element at a time rather than with ElementsAs: a wholly unknown
	// element cannot be read into a struct, and it carries no scope to check
	// anyway, so it is skipped and left to the apply-time check in Create.
	for i, element := range roleBindings.Elements() {
		if element.IsNull() || element.IsUnknown() {
			continue
		}

		object, isObject := element.(types.Object)
		if !isObject {
			continue
		}

		var binding models.InviteRoleBindingModel
		diags := object.As(ctx, &binding, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			continue
		}

		if summary, detail, ok := roleBindingScopeError(binding.ResourceType, binding.ResourceId); !ok {
			resp.Diagnostics.AddAttributeError(
				roleBindingsPath.AtListIndex(i).AtName("resource_id"),
				summary,
				fmt.Sprintf("role_bindings[%d]: %s", i, detail),
			)
		}
	}
}

func NewInviteResource() resource.Resource {
	return &InviteResource{PineconeResource: &PineconeResource{}}
}

// InviteResource defines the resource implementation.
type InviteResource struct {
	*PineconeResource
}

func (r *InviteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (r *InviteResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		inviteRoleBindingScopeValidator{},
	}
}

func (r *InviteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_invite` resource manages an organization invitation — not the resulting membership. " +
			"Creating it sends an invite to the given email with a set of initial roles; deleting it revokes a still-pending invite. " +
			"Once the invitee accepts, the invitation is complete (`status = processed`) and Terraform stops acting on it: destroying " +
			"an accepted invite is a no-op, and that user's roles should be managed from then on with `pinecone_role_binding`. " +
			"Invites are immutable — changing `email` or `role_bindings` replaces the resource by sending a new invite; do not change " +
			"either once the invite is accepted (`status = processed`), because the forced replacement re-sends an invite to an address " +
			"that already belongs to a member and the create fails. Because the API never returns the granted roles, `role_bindings` is " +
			"set only at creation and is not drift-detected or recoverable on import.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the invite.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address to invite. Changing it replaces the resource by sending a new invite. Do not change `email` or `role_bindings` after the invite is accepted (`status = processed`) — the replacement re-invites an address that already belongs to a member and fails; manage that user with `pinecone_role_binding` instead.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 254),
					stringvalidator.RegexMatches(emailRegex, "must be a valid email address"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_bindings": schema.ListNestedAttribute{
				MarkdownDescription: "The initial roles to grant the invitee. Must include at least one organization-scoped binding that grants membership (`OrgOwner`, `OrgManager`, `OrgBillingAdmin`, or `OrgMember`); project-scoped bindings are optional. These are applied only when the invite is created and are not returned by the API, so they cannot be drift-detected or imported. Changing this list replaces the resource; doing so after the invite is accepted fails because it re-invites an existing member. After the invite is accepted, manage the user's roles with `pinecone_role_binding`.",
				Required:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_type": schema.StringAttribute{
							MarkdownDescription: "The kind of resource scope the binding applies to. Valid values are: `organization`, `project`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(pinecone.ResourceTypeOrganization),
									string(pinecone.ResourceTypeProject),
								),
							},
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The role to assign at the resource scope. Organization-scoped values: `OrgOwner`, `OrgManager`, `OrgBillingAdmin`, `OrgMember`. Project-scoped values: `ProjectOwner`, `ProjectManager`, `ProjectMember`, `ProjectEditor`, `ProjectViewer`, `ControlPlaneEditor`, `ControlPlaneViewer`, `DataPlaneEditor`, `DataPlaneViewer`.",
							Required:            true,
						},
						"resource_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the project the binding applies to. Required when `resource_type` is `project`; must be omitted when `resource_type` is `organization`.",
							Optional:            true,
						},
					},
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The lifecycle status of the invite: `pending`, `expired`, or `processed`.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the invite was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the invite expires if not accepted. The default TTL is 7 days.",
				Computed:            true,
			},
			"processed_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the invite was accepted. Null while the invite is still pending or expired.",
				Computed:            true,
			},
		},
	}
}

func (r *InviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.InviteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to create invites.")
		return
	}

	roleBindings := make([]pinecone.RoleBindingInput, 0, len(data.RoleBindings))
	for i, rb := range data.RoleBindings {
		resourceType := rb.ResourceType.ValueString()
		hasResourceId := !rb.ResourceId.IsNull() && !rb.ResourceId.IsUnknown()

		// Backstop for the plan-time inviteRoleBindingScopeValidator, which does
		// not run on every path into Create and skips unknown bindings.
		if summary, detail, ok := roleBindingScopeError(rb.ResourceType, rb.ResourceId); !ok {
			resp.Diagnostics.AddError(summary, fmt.Sprintf("role_bindings[%d]: %s", i, detail))
			return
		}

		input := pinecone.RoleBindingInput{
			ResourceType: pinecone.ResourceType(resourceType),
			Role:         rb.Role.ValueString(),
		}
		if hasResourceId {
			resourceId := rb.ResourceId.ValueString()
			input.ResourceId = &resourceId
		}
		roleBindings = append(roleBindings, input)
	}

	invite, err := r.adminClient.Invite.Create(ctx, &pinecone.CreateInviteParams{
		Email:        data.Email.ValueString(),
		RoleBindings: roleBindings,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create invite", err.Error())
		return
	}

	models.SetInviteResourceModel(&data, invite)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.InviteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read invites.")
		return
	}

	invite, err := r.adminClient.Invite.Describe(ctx, data.Id.ValueString())
	if err != nil {
		if isNotFoundErr(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to describe invite", err.Error())
		}
		return
	}

	// role_bindings are preserved from prior state; the API never returns them.
	models.SetInviteResourceModel(&data, invite)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op. Every configurable attribute forces replacement, so a plan
// never produces an in-place update. The method exists only to satisfy the
// resource.Resource interface.
func (r *InviteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.InviteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.InviteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to delete invites.")
		return
	}

	inviteId := data.Id.ValueString()

	// An accepted invite has already converted to a membership; there is nothing
	// to revoke, so deleting it is a no-op success. Manage that user's roles via
	// pinecone_role_binding instead.
	if r.inviteAcceptedOrGone(ctx, inviteId) {
		return
	}

	err := r.adminClient.Invite.Delete(ctx, inviteId)
	if err != nil {
		if isNotFoundErr(err) {
			return
		}
		// Acceptance can race with destroy: an invite accepted between the check
		// above and this call can no longer be deleted and returns a conflict.
		// If it is now processed (or already gone), treat the delete as a no-op.
		if isConflictErr(err) && r.inviteAcceptedOrGone(ctx, inviteId) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete invite", err.Error())
		return
	}

	// Deletion is asynchronous (HTTP 202); wait until the invite is gone.
	err = retryDeletion(ctx, func() error {
		_, err := r.adminClient.Invite.Describe(ctx, inviteId)
		return err
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for invite to be deleted.", err.Error())
		return
	}
}

// inviteAcceptedOrGone reports whether the invite no longer needs to be revoked:
// it has been accepted (status processed) or no longer exists. A describe error
// other than not-found is treated as "still present" so the caller surfaces the
// original delete failure.
func (r *InviteResource) inviteAcceptedOrGone(ctx context.Context, inviteId string) bool {
	invite, err := r.adminClient.Invite.Describe(ctx, inviteId)
	if err != nil {
		return isNotFoundErr(err)
	}
	return invite.Status == pinecone.InviteStatusProcessed
}

func (r *InviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: invite_id. Note: role_bindings cannot be imported because the
	// API does not return them; the next plan will show a replacement until they
	// are set to match the original invite.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
