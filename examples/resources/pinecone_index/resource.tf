terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "us-west4-gcp"
  # api_key = set via PINECONE_API_KEY env variable
}

resource "pinecone_index" "test" {
  name = "tftestindex"
  # dimension = 512
  # metric    = "cosine"
  # pod_type  = "s1.x1"
}