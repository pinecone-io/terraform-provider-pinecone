terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_api_key" "example" {
  id = "your-api-key-id"
}

output "api_key_name" {
  value = data.pinecone_api_key.example.name
} 