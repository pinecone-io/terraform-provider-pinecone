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

# Document index with integrated embedding: Pinecone embeds the text of body with a dense model
# and the text of summary with a sparse model. body is also stored and indexed for full-text
# search. summary's text is embedded and then discarded.
resource "pinecone_index" "embedded" {
  name = "tftestindex-embedded"
  deployment = {
    managed = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
  schema = {
    fields = {
      body = {
        string = {
          embed            = { model = "llama-text-embed-v2", dimension = 1024 }
          full_text_search = { stemming = true }
        }
      }
      summary = {
        string = {
          sparse_embed = {}
          store_text   = false
        }
      }
    }
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
# This style is planned for deprecation. Legacy integrated indexes still need it.

# Legacy integrated index, read and written through the legacy Records API. For new indexes, use
# integrated embedding on a document index instead.
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
