// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
)

// ProjectResourceModel defines the Project model for the resource.
type ProjectResourceModel struct {
	Id                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	OrganizationId          types.String   `tfsdk:"organization_id"`
	CreatedAt               types.String   `tfsdk:"created_at"`
	ForceEncryptionWithCmek types.Bool     `tfsdk:"force_encryption_with_cmek"`
	MaxPods                 types.Int64    `tfsdk:"max_pods"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

func (model *ProjectResourceModel) Read(ctx context.Context, project *pinecone.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(project.Id)
	model.Name = types.StringValue(project.Name)
	model.OrganizationId = types.StringValue(project.OrganizationId)
	model.ForceEncryptionWithCmek = types.BoolValue(project.ForceEncryptionWithCmek)
	model.MaxPods = types.Int64Value(int64(project.MaxPods))

	if project.CreatedAt != nil {
		model.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	return diags
}

// ProjectDatasourceModel defines the Project model for the datasource.
type ProjectDatasourceModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationId          types.String `tfsdk:"organization_id"`
	CreatedAt               types.String `tfsdk:"created_at"`
	ForceEncryptionWithCmek types.Bool   `tfsdk:"force_encryption_with_cmek"`
	MaxPods                 types.Int64  `tfsdk:"max_pods"`
}

func (model *ProjectDatasourceModel) Read(ctx context.Context, project *pinecone.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(project.Id)
	model.Name = types.StringValue(project.Name)
	model.OrganizationId = types.StringValue(project.OrganizationId)
	model.ForceEncryptionWithCmek = types.BoolValue(project.ForceEncryptionWithCmek)
	model.MaxPods = types.Int64Value(int64(project.MaxPods))

	if project.CreatedAt != nil {
		model.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	return diags
}

// ProjectsDataSourceModel defines the model for listing projects.
type ProjectsDataSourceModel struct {
	Projects []ProjectModel `tfsdk:"projects"`
	Id       types.String   `tfsdk:"id"`
}

type ProjectModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationId          types.String `tfsdk:"organization_id"`
	CreatedAt               types.String `tfsdk:"created_at"`
	ForceEncryptionWithCmek types.Bool   `tfsdk:"force_encryption_with_cmek"`
	MaxPods                 types.Int64  `tfsdk:"max_pods"`
}

func (model *ProjectModel) Read(ctx context.Context, project *pinecone.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Id = types.StringValue(project.Id)
	model.Name = types.StringValue(project.Name)
	model.OrganizationId = types.StringValue(project.OrganizationId)
	model.ForceEncryptionWithCmek = types.BoolValue(project.ForceEncryptionWithCmek)
	model.MaxPods = types.Int64Value(int64(project.MaxPods))

	if project.CreatedAt != nil {
		model.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	return diags
}
