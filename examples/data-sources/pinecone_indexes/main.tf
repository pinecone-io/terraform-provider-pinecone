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

data "pinecone_indexes" "example" {
}
