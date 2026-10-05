terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

# Serverless vector index
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

# Serverless vector index with dedicated read capacity
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

# Index with integrated embedding
resource "pinecone_index" "integrated" {
  name = "tftestindex-integrated"
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
  embed = {
    model = "multilingual-e5-large"
    field_map = {
      text = "chunk_text"
    }
  }
}

# BYOC (Bring Your Own Cloud) vector index
resource "pinecone_index" "byoc" {
  name      = "tftestindex-byoc"
  dimension = 1536
  spec = {
    byoc = {
      environment = "my-byoc-env-id"
    }
  }
}

# Document index with dense vector, sparse vector, and full-text search fields
resource "pinecone_index" "document" {
  name = "tftestindex-documents"
  deployment = {
    managed = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
  schema = {
    fields = {
      embedding = { dense_vector = { dimension = 1536, metric = "dotproduct" } }
      terms     = { sparse_vector = {} }
      body      = { string = { full_text_search = { stemming = true, stop_words = true } } }
      title     = { string = { full_text_search = { ngram = { min_gram = 2, max_gram = 4, prefix_only = true } } } }
    }
  }
  read_capacity = {
    on_demand = {}
  }
}

# Vector index defined by schema, using the reserved _values and _sparse_values fields
resource "pinecone_index" "schema_vectors" {
  name = "tftestindex-schema-vectors"
  deployment = {
    managed = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
  schema = {
    fields = {
      _values        = { dense_vector = { dimension = 1536, metric = "dotproduct" } }
      _sparse_values = { sparse_vector = {} }
    }
  }
}
