package models

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

// InviteResourceModel defines the invite model for the resource.
type InviteResourceModel struct {
	Id           types.String             `tfsdk:"id"`
	Email        types.String             `tfsdk:"email"`
	RoleBindings []InviteRoleBindingModel `tfsdk:"role_bindings"`
	Status       types.String             `tfsdk:"status"`
	CreatedAt    types.String             `tfsdk:"created_at"`
	ExpiresAt    types.String             `tfsdk:"expires_at"`
	ProcessedAt  types.String             `tfsdk:"processed_at"`
}

// InviteRoleBindingModel defines a single role binding granted to the invitee at
// creation time. These are not returned by the API and are preserved from config.
type InviteRoleBindingModel struct {
	ResourceType types.String `tfsdk:"resource_type"`
	Role         types.String `tfsdk:"role"`
	ResourceId   types.String `tfsdk:"resource_id"`
}

// InvitesDataSourceModel defines the invites list model for the data source.
type InvitesDataSourceModel struct {
	Invites []InviteModel `tfsdk:"invites"`
	Id      types.String  `tfsdk:"id"`
}

// InviteModel defines a single invite in the invites list.
type InviteModel struct {
	Id          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	ProcessedAt types.String `tfsdk:"processed_at"`
}

// SetInviteResourceModel maps a Pinecone invite onto the Terraform resource model.
// It intentionally does not touch role_bindings, which are create-time input that
// the API never returns, so they are preserved from prior config/state.
func SetInviteResourceModel(data *InviteResourceModel, invite *pinecone.Invite) {
	data.Id = types.StringValue(invite.Id)
	data.Email = types.StringValue(invite.Email)
	data.Status = types.StringValue(string(invite.Status))
	data.CreatedAt = types.StringValue(invite.CreatedAt.Format(time.RFC3339))
	data.ExpiresAt = formatOptionalTime(invite.ExpiresAt)
	data.ProcessedAt = formatOptionalTime(invite.ProcessedAt)
}

// NewInviteModel creates a new InviteModel from a pinecone.Invite.
func NewInviteModel(invite *pinecone.Invite) *InviteModel {
	return &InviteModel{
		Id:          types.StringValue(invite.Id),
		Email:       types.StringValue(invite.Email),
		Status:      types.StringValue(string(invite.Status)),
		CreatedAt:   types.StringValue(invite.CreatedAt.Format(time.RFC3339)),
		ExpiresAt:   formatOptionalTime(invite.ExpiresAt),
		ProcessedAt: formatOptionalTime(invite.ProcessedAt),
	}
}

// formatOptionalTime renders a nullable timestamp as an RFC3339 string or null.
func formatOptionalTime(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}
