terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_indexes" "all" {}

output "protected_indexes" {
  value = [
    for index in data.pinecone_indexes.all.indexes : index.name
    if index.deletion_protection == "enabled"
  ]
}

output "dedicated_indexes" {
  value = [
    for index in data.pinecone_indexes.all.indexes : index.name
    if try(index.read_capacity.dedicated, null) != null
  ]
}
