terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_project" "test" {
  id = "your-project-id"
}

output "project_name" {
  value = data.pinecone_project.test.name
} 