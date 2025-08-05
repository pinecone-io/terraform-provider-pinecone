terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_projects" "all" {}

output "project_names" {
  value = [for project in data.pinecone_projects.all.projects : project.name]
} 