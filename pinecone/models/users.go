package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

// UserResourceModel defines the user model for the (delete-only) resource.
type UserResourceModel struct {
	Id    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Name  types.String `tfsdk:"name"`
}

// UserDataSourceModel defines the user model for the singular data source.
type UserDataSourceModel struct {
	Id    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Name  types.String `tfsdk:"name"`
}

// UsersDataSourceModel defines the users list model for the data source.
type UsersDataSourceModel struct {
	Email types.String `tfsdk:"email"`
	Users []UserModel  `tfsdk:"users"`
	Id    types.String `tfsdk:"id"`
}

// UserModel defines a single user in the users list.
type UserModel struct {
	Id    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Name  types.String `tfsdk:"name"`
}

// SetUserResourceModel maps a Pinecone user onto the Terraform resource model.
func SetUserResourceModel(data *UserResourceModel, user *pinecone.User) {
	data.Id = types.StringValue(user.Id)
	data.Email = types.StringValue(user.Email)
	data.Name = optionalString(user.Name)
}

// Read populates the UserDataSourceModel from a pinecone.User.
func (m *UserDataSourceModel) Read(user *pinecone.User) {
	m.Id = types.StringValue(user.Id)
	m.Email = types.StringValue(user.Email)
	m.Name = optionalString(user.Name)
}

// NewUserModel creates a new UserModel from a pinecone.User.
func NewUserModel(user *pinecone.User) *UserModel {
	return &UserModel{
		Id:    types.StringValue(user.Id),
		Email: types.StringValue(user.Email),
		Name:  optionalString(user.Name),
	}
}

// optionalString renders a nullable string pointer as a value or null.
func optionalString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
