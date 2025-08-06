terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

# Provider configuration for admin operations (API key management)
provider "pinecone" {
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}

# Example API key resource
resource "pinecone_api_key" "example" {
  name       = "example-api-key"
  project_id = "your-project-id"
} 