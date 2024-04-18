// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

type ProjectApiKeysModel struct {
	Id        types.String         `tfsdk:"id"`
	ProjectId types.String         `tfsdk:"project_id"`
	ApiKeys   []ProjectApiKeyModel `tfsdk:"api_keys"`
}

type ProjectApiKeyModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Secret    types.String `tfsdk:"secret"`
	ProjectId types.String `tfsdk:"project_id"`
}

func (model *ProjectApiKeyModel) Read(ctx context.Context, apiKey *pinecone.APIKeyWithSecret) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(apiKey.Id.String())
	model.Name = types.StringValue(apiKey.Name)
	model.Secret = types.StringValue(apiKey.Secret)
	model.ProjectId = types.StringValue(apiKey.ProjectId.String())

	return diags
}

func (model *ProjectApiKeyModel) ReadWithoutSecret(ctx context.Context, apiKey *pinecone.APIKeyWithoutSecret) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(apiKey.Id.String())
	model.Name = types.StringValue(apiKey.Name)
	model.ProjectId = types.StringValue(apiKey.ProjectId.String())

	return diags
}
