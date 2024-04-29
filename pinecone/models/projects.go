// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

type ProjectsModel struct {
	Id       types.String   `tfsdk:"id"`
	Projects []ProjectModel `tfsdk:"projects"`
}

type ProjectModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (model *ProjectModel) Read(ctx context.Context, project *pinecone.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(project.Id.String())
	model.Name = types.StringValue(project.Name)

	return diags
}
