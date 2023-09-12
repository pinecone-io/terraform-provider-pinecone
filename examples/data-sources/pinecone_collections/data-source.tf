terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "us-west4-gcp"
  api_key     = "1928cd1e-4987-48ba-b0a0-48cdc5f0acdc"
}

data "pinecone_collections" "example" {
}

