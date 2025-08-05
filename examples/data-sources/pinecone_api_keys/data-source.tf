terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_api_keys" "example" {
  project_id = "your-project-id"
}

output "api_key_names" {
  value = [for api_key in data.pinecone_api_keys.example.api_keys : api_key.name]
}

output "api_key_count" {
  value = length(data.pinecone_api_keys.example.api_keys)
} 