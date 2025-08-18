package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProjectResourceModel defines the project model for the resource.
type ProjectResourceModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationId          types.String `tfsdk:"organization_id"`
	ForceEncryptionWithCmek types.Bool   `tfsdk:"force_encryption_with_cmek"`
	MaxPods                 types.Int64  `tfsdk:"max_pods"`
	CreatedAt               types.String `tfsdk:"created_at"`
}
