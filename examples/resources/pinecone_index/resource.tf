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
  dimension = 10
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
    }
  }
}
