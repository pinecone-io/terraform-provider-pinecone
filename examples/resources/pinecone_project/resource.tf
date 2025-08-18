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

# Create a basic project
resource "pinecone_project" "example" {
  name = "example-project"
}

# Create a project with CMEK encryption enabled
resource "pinecone_project" "encrypted" {
  name                       = "encrypted-project"
  force_encryption_with_cmek = true
}

# Create a project with custom max pods
resource "pinecone_project" "custom_pods" {
  name     = "custom-pods-project"
  max_pods = 10
}

# Create a project with all options
resource "pinecone_project" "full_featured" {
  name                       = "full-featured-project"
  force_encryption_with_cmek = false
  max_pods                   = 5
}

output "project_id" {
  description = "The ID of the created project"
  value       = pinecone_project.example.id
}

output "project_name" {
  description = "The name of the created project"
  value       = pinecone_project.example.name
}

output "organization_id" {
  description = "The organization ID of the project"
  value       = pinecone_project.example.organization_id
}
