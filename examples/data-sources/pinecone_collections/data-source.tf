terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  # api_key = set via PINECONE_API_KEY env variable
}

data "pinecone_collections" "test" {
}

