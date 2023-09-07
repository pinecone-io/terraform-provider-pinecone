terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
  # environment = "gcp-starter"
}

data "pinecone_indexes" "example" {}
