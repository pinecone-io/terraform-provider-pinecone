// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
)

// APIKeyResourceModel defines the API Key model for the resource.
type APIKeyResourceModel struct {
	Id          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	ProjectId   types.String   `tfsdk:"project_id"`
	Roles       types.List     `tfsdk:"roles"`
	Value       types.String   `tfsdk:"value"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (model *APIKeyResourceModel) Read(ctx context.Context, apiKey *pinecone.APIKey) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(apiKey.Id)
	model.Name = types.StringValue(apiKey.Name)
	model.ProjectId = types.StringValue(apiKey.ProjectId)
	
	// Convert roles to list
	roles, diags := types.ListValueFrom(ctx, types.StringType, apiKey.Roles)
	if diags.HasError() {
		return diags
	}
	model.Roles = roles

	return diags
}

func (model *APIKeyResourceModel) ReadWithSecret(ctx context.Context, apiKeyWithSecret *pinecone.APIKeyWithSecret) diag.Diagnostics {
	var diags diag.Diagnostics

	// Read the regular API key details
	diags.Append(model.Read(ctx, &apiKeyWithSecret.Key)...)
	
	// Set the secret value
	model.Value = types.StringValue(apiKeyWithSecret.Value)

	return diags
}

// APIKeyDatasourceModel defines the API Key model for the datasource.
type APIKeyDatasourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	Roles     types.List   `tfsdk:"roles"`
}

func (model *APIKeyDatasourceModel) Read(ctx context.Context, apiKey *pinecone.APIKey) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(apiKey.Id)
	model.Name = types.StringValue(apiKey.Name)
	model.ProjectId = types.StringValue(apiKey.ProjectId)
	
	// Convert roles to list
	roles, diags := types.ListValueFrom(ctx, types.StringType, apiKey.Roles)
	if diags.HasError() {
		return diags
	}
	model.Roles = roles

	return diags
}

// APIKeysDataSourceModel defines the model for listing API keys.
type APIKeysDataSourceModel struct {
	APIKeys []APIKeyModel `tfsdk:"api_keys"`
	Id      types.String  `tfsdk:"id"`
}

type APIKeyModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	Roles     types.List   `tfsdk:"roles"`
}

func (model *APIKeyModel) Read(ctx context.Context, apiKey *pinecone.APIKey) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(apiKey.Id)
	model.Name = types.StringValue(apiKey.Name)
	model.ProjectId = types.StringValue(apiKey.ProjectId)
	
	// Convert roles to list
	roles, diags := types.ListValueFrom(ctx, types.StringType, apiKey.Roles)
	if diags.HasError() {
		return diags
	}
	model.Roles = roles

	return diags
} 