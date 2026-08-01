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

# Create a service account. The client_secret is returned only once and stored
# in state — treat it as a sensitive credential.
resource "pinecone_service_account" "example" {
  name = "example-service-account"
}

# Grant the service account an organization-scoped role. Role bindings are
# managed separately via the pinecone_role_binding resource.
resource "pinecone_role_binding" "example_org_member" {
  principal_id   = pinecone_service_account.example.id
  principal_type = "service_account"
  resource_type  = "organization"
  role           = "OrgMember"
}

# Rotate the client secret by changing rotate_trigger to any new value. The new
# secret replaces the value stored in client_secret.
resource "pinecone_service_account" "rotatable" {
  name           = "rotatable-service-account"
  rotate_trigger = "2026-01-01"
}

output "example_client_id" {
  value = pinecone_service_account.example.client_id
}

output "example_client_secret" {
  value     = pinecone_service_account.example.client_secret
  sensitive = true
}
