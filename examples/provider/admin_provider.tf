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

# List all available projects
data "pinecone_projects" "all" {}

# Example API key resource using the first available project
resource "pinecone_api_key" "example" {
  name       = "example-api-key"
  project_id = data.pinecone_projects.all.projects[0].id
} 