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

# List all role bindings in the organization.
data "pinecone_role_bindings" "all" {}

# List role bindings for a specific principal.
# principal_type is required when principal_id is set.
data "pinecone_role_bindings" "for_service_account" {
  principal_type = "service_account"
  principal_id   = "your-service-account-id"
}

# List role bindings scoped to a specific project.
# resource_type is required when resource_id is set.
data "pinecone_role_bindings" "for_project" {
  resource_type = "project"
  resource_id   = "your-project-id"
}
