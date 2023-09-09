terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
	environment = "gcp-starter"
}

resource "pinecone_index" "test" {
  name = "frank"
  dimension = 512
  metric = "cosine"
}