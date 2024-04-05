terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "us-west4-gcp"
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
