terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  api_key = var.pinecone_api_key
}

data "pinecone_collections" "test" {
}

