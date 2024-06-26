terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

resource "pinecone_index" "test" {
  name      = "tftestindex"
  metric    = "cosine"
  dimension = 1536
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
    }
  }
}

data "pinecone_index" "test" {
  name = pinecone_index.test.name
}
