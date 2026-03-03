terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

# Basic serverless index
resource "pinecone_index" "serverless" {
  name      = "tftestindex"
  dimension = 1536
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
}

# Serverless index with dedicated read capacity
resource "pinecone_index" "serverless_dedicated" {
  name      = "tftestindex-dedicated"
  dimension = 1536
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
      read_capacity = {
        dedicated = {
          node_type = "b1"
          replicas  = 1
          shards    = 1
        }
      }
    }
  }
}

# Serverless index with metadata schema (selective field indexing)
resource "pinecone_index" "serverless_schema" {
  name      = "tftestindex-schema"
  dimension = 1536
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
      schema = {
        fields = {
          "category" = { filterable = true }
          "language" = { filterable = true }
        }
      }
    }
  }
}

# BYOC (Bring Your Own Cloud) index
resource "pinecone_index" "byoc" {
  name      = "tftestindex-byoc"
  dimension = 1536
  spec = {
    byoc = {
      environment = "my-byoc-env-id"
    }
  }
}

# Pod-based index
resource "pinecone_index" "pod" {
  name      = "tftestindex-pod"
  dimension = 1536
  spec = {
    pod = {
      environment = "us-west4-gcp"
      pod_type    = "s1.x1"
    }
  }
}
