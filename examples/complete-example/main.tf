# Complete example demonstrating Pinecone Terraform provider capabilities
# This example shows:
# 1. Using data sources to list and read existing projects
# 2. Creating a new project resource
# 3. Creating an API key within the project
# 4. Outputting various project and API key information

terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

# Configure the provider with admin credentials
# Set PINECONE_CLIENT_ID and PINECONE_CLIENT_SECRET environment variables
provider "pinecone" {}

# List all available projects (data source)
data "pinecone_projects" "all" {}

# Create a test project
resource "pinecone_project" "test" {
  name = "terraform-test-project"
}

# Read the created project using data source
data "pinecone_project" "test" {
  id = pinecone_project.test.id
}

# Create an API key within the project
resource "pinecone_api_key" "test" {
  name       = "terraform-test-api-key"
  project_id = pinecone_project.test.id

  depends_on = [pinecone_project.test]
}

# Output the results
output "project_name" {
  value = pinecone_project.test.name
}

output "project_id" {
  value = pinecone_project.test.id
}

output "api_key_id" {
  value = pinecone_api_key.test.id
}

output "api_key_name" {
  value = pinecone_api_key.test.name
}

output "api_key_value" {
  value     = pinecone_api_key.test.key
  sensitive = true
}

# Output data source results
output "total_projects_count" {
  description = "Total number of projects in the organization"
  value       = length(data.pinecone_projects.all.projects)
}

output "project_from_data_source" {
  description = "Project details from data source"
  value = {
    id         = data.pinecone_project.test.id
    name       = data.pinecone_project.test.name
    created_at = data.pinecone_project.test.created_at
  }
}

output "all_project_names" {
  description = "Names of all projects in the organization"
  value       = [for project in data.pinecone_projects.all.projects : project.name]
}
