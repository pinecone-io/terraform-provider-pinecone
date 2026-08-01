package models

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

// ServiceAccountResourceModel defines the service account model for the resource.
type ServiceAccountResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ClientId      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	RotateTrigger types.String `tfsdk:"rotate_trigger"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

// ServiceAccountDataSourceModel defines the service account model for the singular data source.
type ServiceAccountDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ClientId  types.String `tfsdk:"client_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// ServiceAccountsDataSourceModel defines the service accounts list model for the data source.
type ServiceAccountsDataSourceModel struct {
	ServiceAccounts []ServiceAccountModel `tfsdk:"service_accounts"`
	Id              types.String          `tfsdk:"id"`
}

// ServiceAccountModel defines a single service account in the service accounts list.
type ServiceAccountModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ClientId  types.String `tfsdk:"client_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// SetServiceAccountResourceModel maps a Pinecone service account onto the Terraform
// resource model. It intentionally does not touch client_secret or rotate_trigger,
// which are managed separately (the secret is only returned at create/rotation).
func SetServiceAccountResourceModel(data *ServiceAccountResourceModel, sa *pinecone.ServiceAccount) {
	data.Id = types.StringValue(sa.Id)
	data.Name = types.StringValue(sa.Name)
	data.ClientId = types.StringValue(sa.ClientId)
	data.CreatedAt = types.StringValue(sa.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(sa.UpdatedAt.Format(time.RFC3339))
}

// Read populates the ServiceAccountDataSourceModel from a pinecone.ServiceAccount.
func (m *ServiceAccountDataSourceModel) Read(sa *pinecone.ServiceAccount) {
	m.Id = types.StringValue(sa.Id)
	m.Name = types.StringValue(sa.Name)
	m.ClientId = types.StringValue(sa.ClientId)
	m.CreatedAt = types.StringValue(sa.CreatedAt.Format(time.RFC3339))
	m.UpdatedAt = types.StringValue(sa.UpdatedAt.Format(time.RFC3339))
}

// NewServiceAccountModel creates a new ServiceAccountModel from a pinecone.ServiceAccount.
func NewServiceAccountModel(sa *pinecone.ServiceAccount) *ServiceAccountModel {
	return &ServiceAccountModel{
		Id:        types.StringValue(sa.Id),
		Name:      types.StringValue(sa.Name),
		ClientId:  types.StringValue(sa.ClientId),
		CreatedAt: types.StringValue(sa.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(sa.UpdatedAt.Format(time.RFC3339)),
	}
}
