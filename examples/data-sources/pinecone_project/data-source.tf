terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  client_id     = var.client_id
  client_secret = var.client_secret
}

# Read a specific project by ID
data "pinecone_project" "example" {
  id = var.project_id
}

# Output the project details
output "project_name" {
  description = "The name of the project"
  value       = data.pinecone_project.example.name
}

output "project_organization_id" {
  description = "The organization ID of the project"
  value       = data.pinecone_project.example.organization_id
}

output "project_force_encryption_with_cmek" {
  description = "Whether CMEK encryption is forced"
  value       = data.pinecone_project.example.force_encryption_with_cmek
}

output "project_max_pods" {
  description = "The maximum number of pods allowed"
  value       = data.pinecone_project.example.max_pods
}

output "project_created_at" {
  description = "When the project was created"
  value       = data.pinecone_project.example.created_at
}
