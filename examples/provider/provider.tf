terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
  api_key     = var.api_key
  environment = "us-west4-gcp"
}

resource "pinecone_index" "example" {
  name      = "frank"
  dimension = 512
  metric    = "cosine"
}


