terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "us-west4-gcp"
  api_key     = "api-key"
}

resource "pinecone_collection" "example_collection" {
  name = "my_example_collection"
  # Add any other required or optional attributes for the collection here
}

output "example_collection_id" {
  value = pinecone_collection.example_collection.id
}