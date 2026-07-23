package models

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

// RoleBindingResourceModel defines the role binding model for the resource.
type RoleBindingResourceModel struct {
	Id            types.String `tfsdk:"id"`
	PrincipalId   types.String `tfsdk:"principal_id"`
	PrincipalType types.String `tfsdk:"principal_type"`
	ResourceType  types.String `tfsdk:"resource_type"`
	ResourceId    types.String `tfsdk:"resource_id"`
	Role          types.String `tfsdk:"role"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

// RoleBindingsDataSourceModel defines the role bindings list model for the data source.
type RoleBindingsDataSourceModel struct {
	PrincipalType types.String       `tfsdk:"principal_type"`
	PrincipalId   types.String       `tfsdk:"principal_id"`
	ResourceType  types.String       `tfsdk:"resource_type"`
	ResourceId    types.String       `tfsdk:"resource_id"`
	Role          types.String       `tfsdk:"role"`
	RoleBindings  []RoleBindingModel `tfsdk:"role_bindings"`
	Id            types.String       `tfsdk:"id"`
}

// RoleBindingModel defines a single role binding in the role bindings list.
type RoleBindingModel struct {
	Id            types.String `tfsdk:"id"`
	PrincipalId   types.String `tfsdk:"principal_id"`
	PrincipalType types.String `tfsdk:"principal_type"`
	ResourceId    types.String `tfsdk:"resource_id"`
	ResourceType  types.String `tfsdk:"resource_type"`
	Role          types.String `tfsdk:"role"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

// SetRoleBindingResourceModel maps a Pinecone role binding onto the Terraform resource model.
func SetRoleBindingResourceModel(data *RoleBindingResourceModel, rb *pinecone.RoleBinding) {
	data.Id = types.StringValue(rb.Id)
	data.PrincipalId = types.StringValue(rb.PrincipalId)
	data.PrincipalType = types.StringValue(string(rb.PrincipalType))
	data.ResourceType = types.StringValue(string(rb.ResourceType))
	// Only project-scoped bindings carry a user-supplied resource_id. For
	// organization scope the server returns the organization ID, but the config
	// omits resource_id, so keep state null to avoid perpetual drift and to keep
	// replaces from re-sending it under organization scope.
	if rb.ResourceType == pinecone.ResourceTypeProject {
		data.ResourceId = types.StringValue(rb.ResourceId)
	} else {
		data.ResourceId = types.StringNull()
	}
	data.Role = types.StringValue(rb.Role)
	data.CreatedAt = types.StringValue(rb.CreatedAt.Format(time.RFC3339))
}

// NewRoleBindingModel creates a new RoleBindingModel from a pinecone.RoleBinding.
func NewRoleBindingModel(rb *pinecone.RoleBinding) *RoleBindingModel {
	return &RoleBindingModel{
		Id:            types.StringValue(rb.Id),
		PrincipalId:   types.StringValue(rb.PrincipalId),
		PrincipalType: types.StringValue(string(rb.PrincipalType)),
		ResourceId:    types.StringValue(rb.ResourceId),
		ResourceType:  types.StringValue(string(rb.ResourceType)),
		Role:          types.StringValue(rb.Role),
		CreatedAt:     types.StringValue(rb.CreatedAt.Format(time.RFC3339)),
	}
}
