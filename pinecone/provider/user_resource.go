package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{PineconeResource: &PineconeResource{}}
}

// UserResource defines the resource implementation. Users cannot be created or
// updated through the admin API — they join via invites — so this resource only
// supports importing an existing user, reading it, and deleting it (which removes
// the user from the organization).
type UserResource struct {
	*PineconeResource
}

// immutableUserIdPlanModifier blocks changing a managed user's id in place. It
// intentionally does NOT use RequiresReplace: replacing the resource would delete
// the original user (removing them from the organization) before failing to
// recreate one, so an id change is surfaced as a hard error instead.
type immutableUserIdPlanModifier struct{}

func (m immutableUserIdPlanModifier) Description(ctx context.Context) string {
	return "Prevents changing a managed user's id in place, which would otherwise remove the original user from the organization."
}

func (m immutableUserIdPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableUserIdPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// No prior state on create; nothing to protect.
	if req.State.Raw.IsNull() {
		return
	}
	// Only guard a change between two concrete values.
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}
	if !req.StateValue.Equal(req.PlanValue) {
		resp.Diagnostics.AddError(
			"User id cannot be changed",
			"A managed pinecone_user's id is immutable because changing it would remove the original user from the organization. "+
				"To manage a different user, remove this resource from state (`terraform state rm`) and import the intended user.",
		)
	}
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_user` resource manages an existing organization member's membership. Users cannot be created or updated through Terraform — they join the organization by accepting a `pinecone_invite`. Bring an existing user under management with `terraform import`, then destroying the resource removes the user from the organization. To change a user's roles, use `pinecone_role_binding`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the user to manage. Use `terraform import` to populate it from an existing user. This value is immutable — changing it is rejected (changing it would remove the original user from the organization); to manage a different user, `terraform state rm` this resource and import the intended one.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					immutableUserIdPlanModifier{},
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The user's email address.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The user's display name. Null if the user has not set one.",
				Computed:            true,
			},
		},
	}
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Users cannot be created",
		"The pinecone_user resource cannot create users; users join the organization by accepting a pinecone_invite. "+
			"To manage an existing user, import it with `terraform import pinecone_user.<name> <user_id>`.",
	)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read users.")
		return
	}

	user, err := r.adminClient.User.Describe(ctx, data.Id.ValueString())
	if err != nil {
		if isNotFoundErr(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to describe user", err.Error())
		}
		return
	}

	models.SetUserResourceModel(&data, user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op. The only configurable attribute (id) forces replacement and
// no other attribute is writable, so a plan never produces an in-place update.
func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to delete users.")
		return
	}

	err := r.adminClient.User.Delete(ctx, data.Id.ValueString())
	if err != nil {
		if !isNotFoundErr(err) {
			resp.Diagnostics.AddError("Failed to delete user", err.Error())
		}
		return
	}

	// Deletion is asynchronous (HTTP 202); wait until the user is gone.
	err = retryDeletion(ctx, func() error {
		_, err := r.adminClient.User.Describe(ctx, data.Id.ValueString())
		return err
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for user to be deleted.", err.Error())
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: user_id
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
