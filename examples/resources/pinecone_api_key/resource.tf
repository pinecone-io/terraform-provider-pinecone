terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

resource "pinecone_project" "example" {
  name = "terraform-example-project"
}

resource "pinecone_api_key" "example" {
  name       = "terraform-example-api-key"
  project_id = pinecone_project.example.id
} 