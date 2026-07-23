terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}

# Grant a service account an organization-scoped role.
# For organization scope, omit resource_id (it is computed to the organization ID).
resource "pinecone_role_binding" "org_member" {
  principal_id   = "your-service-account-id"
  principal_type = "service_account"
  resource_type  = "organization"
  role           = "OrgMember"
}

# Grant a service account a project-scoped role.
# For project scope, resource_id is required and must be the project ID.
resource "pinecone_role_binding" "project_editor" {
  principal_id   = "your-service-account-id"
  principal_type = "service_account"
  resource_type  = "project"
  resource_id    = "your-project-id"
  role           = "ProjectEditor"
}
