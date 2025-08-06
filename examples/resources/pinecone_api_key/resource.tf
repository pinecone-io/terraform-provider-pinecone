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

resource "pinecone_api_key" "example" {
  name       = "example-api-key"
  project_id = "your-project-id"
}

output "api_key_roles" {
  description = "The roles assigned to the API key"
  value       = pinecone_api_key.example.roles
} 