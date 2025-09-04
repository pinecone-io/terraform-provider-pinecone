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

# Read all available projects
data "pinecone_projects" "all" {}

# Output the count of projects
output "project_count" {
  description = "Total number of projects"
  value       = length(data.pinecone_projects.all.projects)
}

# Output all project names
output "project_names" {
  description = "Names of all projects"
  value       = [for project in data.pinecone_projects.all.projects : project.name]
}

# Output all project IDs
output "project_ids" {
  description = "IDs of all projects"
  value       = [for project in data.pinecone_projects.all.projects : project.id]
}

# Output projects with CMEK encryption enabled
output "cmek_projects" {
  description = "Projects with CMEK encryption enabled"
  value       = [for project in data.pinecone_projects.all.projects : project.name if project.force_encryption_with_cmek]
}

# Output projects with pod limits
output "projects_with_pod_limits" {
  description = "Projects with pod limits configured"
  value = [for project in data.pinecone_projects.all.projects : {
    name     = project.name
    max_pods = project.max_pods
  } if project.max_pods > 0]
}
