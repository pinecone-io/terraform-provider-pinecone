package models

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
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

// ProjectDataSourceModel defines the project model for the data source.
type ProjectDataSourceModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationId          types.String `tfsdk:"organization_id"`
	ForceEncryptionWithCmek types.Bool   `tfsdk:"force_encryption_with_cmek"`
	MaxPods                 types.Int64  `tfsdk:"max_pods"`
	CreatedAt               types.String `tfsdk:"created_at"`
}

// ProjectsDataSourceModel defines the projects list model for the data source.
type ProjectsDataSourceModel struct {
	Projects []ProjectModel `tfsdk:"projects"`
	Id       types.String   `tfsdk:"id"`
}

// ProjectModel defines a single project in the projects list.
type ProjectModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationId          types.String `tfsdk:"organization_id"`
	ForceEncryptionWithCmek types.Bool   `tfsdk:"force_encryption_with_cmek"`
	MaxPods                 types.Int64  `tfsdk:"max_pods"`
	CreatedAt               types.String `tfsdk:"created_at"`
}

// Read populates the ProjectDataSourceModel from a pinecone.Project.
func (m *ProjectDataSourceModel) Read(project *pinecone.Project) {
	m.Id = types.StringValue(project.Id)
	m.Name = types.StringValue(project.Name)
	m.OrganizationId = types.StringValue(project.OrganizationId)
	m.ForceEncryptionWithCmek = types.BoolValue(project.ForceEncryptionWithCmek)
	m.MaxPods = types.Int64Value(int64(project.MaxPods))
	if project.CreatedAt != nil {
		m.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	} else {
		m.CreatedAt = types.StringNull()
	}
}

// NewProjectModel creates a new ProjectModel from a pinecone.Project.
func NewProjectModel(project *pinecone.Project) *ProjectModel {
	model := &ProjectModel{
		Id:                      types.StringValue(project.Id),
		Name:                    types.StringValue(project.Name),
		OrganizationId:          types.StringValue(project.OrganizationId),
		ForceEncryptionWithCmek: types.BoolValue(project.ForceEncryptionWithCmek),
		MaxPods:                 types.Int64Value(int64(project.MaxPods)),
	}

	if project.CreatedAt != nil {
		model.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	return model
}
