// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/skyscrapr/pinecone-sdk-go/pinecone"
)

type ProjectApiKeysModel struct {
	Id        types.String         `tfsdk:"id"`
	ProjectId types.String         `tfsdk:"project_id"`
	ApiKeys   []ProjectApiKeyModel `tfsdk:"api_keys"`
}

type ProjectApiKeyModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
}

func (model *ProjectApiKeyModel) Read(ctx context.Context, project *pinecone.ProjectApiKey) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(project.ID)
	model.Name = types.StringValue(project.Name)
	model.ProjectId = types.StringValue(project.ProjectID)

	return diags
}
