terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "gcp-starter"
  # api_key = set via PINECONE_API_KEY env variable
}

resource "pinecone_index" "test" {
  name      = "tftestindex"
  dimension = 512
  metric    = "cosine"
}

resource "pinecone_collection" "test" {
  name   = "tftestindex"
  source = pinecone_index.test.name
}