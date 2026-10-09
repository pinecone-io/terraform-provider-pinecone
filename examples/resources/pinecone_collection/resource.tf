terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

# Collections are created from pod-based indexes, which can no longer be created, so the source
# is an existing pod-based index.
variable "pod_index_name" {
  type = string
}

resource "pinecone_collection" "test" {
  name   = "tftestcollection"
  source = var.pod_index_name
}
