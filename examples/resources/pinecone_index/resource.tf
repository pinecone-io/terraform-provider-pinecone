terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

# ── Indexes described with schema and deployment ─────────────────────────────

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

# Vector index, described by the reserved fields _values (dense) and _sparse_values (sparse)
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
      _values = { dense_vector = { dimension = 1536, metric = "cosine" } }
    }
  }
}

# BYOC (Bring Your Own Cloud) vector index. BYOC indexes need dedicated read capacity.
resource "pinecone_index" "schema_byoc" {
  name = "tftestindex-schema-byoc"
  deployment = {
    byoc = {
      environment = "my-byoc-env-id"
    }
  }
  schema = {
    fields = {
      _values = { dense_vector = { dimension = 1536, metric = "cosine" } }
    }
  }
  read_capacity = {
    dedicated = {
      node_type = "b1"
      replicas  = 1
      shards    = 1
    }
  }
}

# ── Indexes described with dimension, metric, and spec ───────────────────────
# This style is planned for deprecation. Indexes with integrated embedding still need it.

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

# Serverless vector index with dedicated read capacity
resource "pinecone_index" "serverless" {
  name      = "tftestindex"
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
