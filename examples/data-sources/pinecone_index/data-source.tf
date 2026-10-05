terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_index" "products" {
  name = "products"
}

output "products_vectors" {
  value = {
    dimension = data.pinecone_index.products.schema.fields["_values"].dense_vector.dimension
    metric    = data.pinecone_index.products.schema.fields["_values"].dense_vector.metric
    region    = data.pinecone_index.products.deployment.managed.region
  }
}

data "pinecone_index" "articles" {
  name = "articles"
}

output "articles_full_text_search_fields" {
  value = {
    for name, field in data.pinecone_index.articles.schema.fields :
    name => field.string.full_text_search
    if try(field.string.full_text_search, null) != null
  }
}
