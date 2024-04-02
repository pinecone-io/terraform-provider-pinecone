terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "us-west4-gcp"
  # api_key = set via PINECONE_API_KEY env variable
}

resource "pinecone_index" "test" {
  name = "tftestindex"
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
