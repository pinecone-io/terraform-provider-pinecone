package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApiKeyResourceModel defines the API key model for the resource.
type ApiKeyResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	Key       types.String `tfsdk:"key"`
} 