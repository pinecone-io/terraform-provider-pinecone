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

# Serverless Indexes
resource "pinecone_index" "serverless_index" {
  name      = "tf_serverless_index"
  dimension = 1536
  metric    = "cosine"
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
    }
  }
}
# Pod Indexes
resource "pinecone_index" "pod_index" {
  name      = "tf_pod_index"
  dimension = 1536
  metric    = "cosine"
  spec = {
    pod = {
      environment = "us-west1-gcp"
    }
  }
}
