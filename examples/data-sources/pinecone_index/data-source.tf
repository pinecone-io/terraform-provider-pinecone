terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  api_key = var.pinecone_api_key
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

data "pinecone_index" "test" {
  name = pinecone_index.test.name
}
