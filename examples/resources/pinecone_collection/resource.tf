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
    pod = {
      environment = "us-west4-gcp"
      pod_type    = "s1.x1"
    }
  }
}

resource "pinecone_collection" "test" {
  name   = "tftestcollection"
  source = pinecone_index.test.name
}
