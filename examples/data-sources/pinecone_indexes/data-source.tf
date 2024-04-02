terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "gcp-starter"
  # api_key = set via PINECONE_API_KEY env variable
}

resource "pinecone_index" "test" {
  name = "tftestindex"
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
    }
  }
}

data "pinecone_indexes" "test" {
}