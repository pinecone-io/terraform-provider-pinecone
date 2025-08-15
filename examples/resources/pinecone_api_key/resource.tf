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

# Create API key with default roles (ProjectEditor)
resource "pinecone_api_key" "example" {
  name       = "example-api-key"
  project_id = "your-project-id"
}

# Create API key with custom roles
resource "pinecone_api_key" "custom_roles" {
  name       = "custom-roles-api-key"
  project_id = "your-project-id"
  roles      = ["ProjectViewer", "DataPlaneViewer"]
}

# Create an API key that can be updated later
resource "pinecone_api_key" "updatable" {
  name       = "example-updatable-key"
  project_id = "your-project-id"
  roles      = ["ProjectEditor"]
}

# Example of how to update the API key
# resource "pinecone_api_key" "updatable" {
#   name  = "updated-name"
#   roles = ["ProjectViewer", "DataPlaneViewer"]
# }

output "api_key_roles" {
  description = "The roles assigned to the API key"
  value       = pinecone_api_key.example.roles
}
