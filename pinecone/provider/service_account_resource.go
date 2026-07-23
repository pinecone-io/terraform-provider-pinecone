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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServiceAccountResource{}
var _ resource.ResourceWithImportState = &ServiceAccountResource{}

func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{PineconeResource: &PineconeResource{}}
}

// ServiceAccountResource defines the resource implementation.
type ServiceAccountResource struct {
	*PineconeResource
}

// rotateSecretPlanModifier keeps the stored client_secret stable across plans
// unless rotate_trigger changes. Without it, client_secret (a computed value)
// would plan as "known after apply" on every in-place update, and pairing it
// with UseStateForUnknown would instead make a rotation plan inconsistent (the
// plan would promise the old secret while Update produces a new one).
type rotateSecretPlanModifier struct{}

func (m rotateSecretPlanModifier) Description(ctx context.Context) string {
	return "Preserves the stored client secret unless rotate_trigger changes."
}

func (m rotateSecretPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m rotateSecretPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// On create there is no prior state; leave the value unknown for Create to fill.
	if req.State.Raw.IsNull() {
		return
	}
	// Respect an already-known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	var planTrigger, stateTrigger types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("rotate_trigger"), &planTrigger)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("rotate_trigger"), &stateTrigger)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the current secret unless a rotation is actually requested. When
	// rotation will happen, leave the value unknown so Update can store the new
	// secret; otherwise keep the stored value.
	if !rotateRequested(planTrigger, stateTrigger) {
		resp.PlanValue = req.StateValue
	}
}

// rotateRequested reports whether a rotate_trigger change should rotate the
// secret. Rotation happens only when moving between two concrete values. Adding
// the attribute (null -> value), removing it (value -> null), an unknown value,
// or importing (where the prior value cannot be recovered) must NOT rotate an
// otherwise-valid credential.
func rotateRequested(plan, state types.String) bool {
	if plan.IsNull() || plan.IsUnknown() || state.IsNull() || state.IsUnknown() {
		return false
	}
	return !plan.Equal(state)
}

func (r *ServiceAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `pinecone_service_account` resource lets you create and manage organization service accounts in Pinecone. A service account authenticates with an OAuth client ID and secret; the secret is returned only at creation (or rotation) and cannot be retrieved later. Grant a service account roles with the `pinecone_role_binding` resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Service account identifier. Use this as the `principal_id` when creating a `pinecone_role_binding`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service account. Must be 1-80 characters long.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 80),
				},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth client ID the service account uses to obtain access tokens.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The OAuth client secret. Returned only once, at creation or rotation. Change `rotate_trigger` to rotate it. " +
					"**This value is stored in plaintext in Terraform state.** `Sensitive` redacts it from CLI output and logs " +
					"but does not encrypt it in state — secure your state backend (encryption at rest, restricted access) and " +
					"never commit state to version control.",
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					rotateSecretPlanModifier{},
				},
			},
			"rotate_trigger": schema.StringAttribute{
				MarkdownDescription: "An arbitrary value used to rotate the service account's client secret. Changing it from one non-empty value to another rotates the secret and stores the new value in `client_secret`. Setting it for the first time (or clearing it) establishes a baseline and does not rotate, so an existing credential is never invalidated unintentionally.",
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the service account was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp of the service account's most recent metadata update.",
				Computed:            true,
			},
		},
	}
}

func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ServiceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to create service accounts.")
		return
	}

	saWithSecret, err := r.adminClient.ServiceAccount.Create(ctx, &pinecone.CreateServiceAccountParams{
		Name: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create service account", err.Error())
		return
	}

	models.SetServiceAccountResourceModel(&data, &saWithSecret.ServiceAccount)
	data.ClientSecret = types.StringValue(saWithSecret.ClientSecret)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ServiceAccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to read service accounts.")
		return
	}

	sa, err := r.adminClient.ServiceAccount.Describe(ctx, data.Id.ValueString())
	if err != nil {
		if isNotFoundErr(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to describe service account", err.Error())
		}
		return
	}

	// client_secret and rotate_trigger are preserved from prior state; the
	// describe response never includes the secret.
	models.SetServiceAccountResourceModel(&data, sa)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state models.ServiceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to update service accounts.")
		return
	}

	saId := state.Id.ValueString()
	var sa *pinecone.ServiceAccount

	// Apply a name change.
	if !data.Name.Equal(state.Name) {
		name := data.Name.ValueString()
		updated, err := r.adminClient.ServiceAccount.Update(ctx, saId, &pinecone.UpdateServiceAccountParams{Name: &name})
		if err != nil {
			resp.Diagnostics.AddError("Failed to update service account", err.Error())
			return
		}
		sa = updated
	}

	// Rotate the secret only on a genuine value change of rotate_trigger.
	rotated := false
	if rotateRequested(data.RotateTrigger, state.RotateTrigger) {
		saWithSecret, err := r.adminClient.ServiceAccount.RotateSecret(ctx, saId)
		if err != nil {
			resp.Diagnostics.AddError("Failed to rotate service account secret", err.Error())
			return
		}
		sa = &saWithSecret.ServiceAccount
		data.ClientSecret = types.StringValue(saWithSecret.ClientSecret)
		rotated = true
	}

	// If neither call ran (should not happen for an in-place update), refresh
	// from the API so computed fields stay accurate.
	if sa == nil {
		refreshed, err := r.adminClient.ServiceAccount.Describe(ctx, saId)
		if err != nil {
			resp.Diagnostics.AddError("Failed to describe service account", err.Error())
			return
		}
		sa = refreshed
	}

	models.SetServiceAccountResourceModel(&data, sa)
	if !rotated {
		// The secret is unchanged; carry the stored value forward.
		data.ClientSecret = state.ClientSecret
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ServiceAccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.adminClient == nil {
		resp.Diagnostics.AddError("Admin client not configured", "Admin client credentials (client_id and client_secret) are required to delete service accounts.")
		return
	}

	err := r.adminClient.ServiceAccount.Delete(ctx, data.Id.ValueString())
	if err != nil {
		if !isNotFoundErr(err) {
			resp.Diagnostics.AddError("Failed to delete service account", err.Error())
		}
		return
	}

	// Deletion is asynchronous (HTTP 202); wait until the service account is gone.
	err = retryDeletion(ctx, func() error {
		_, err := r.adminClient.ServiceAccount.Describe(ctx, data.Id.ValueString())
		return err
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for service account to be deleted.", err.Error())
		return
	}
}

func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: service_account_id
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
