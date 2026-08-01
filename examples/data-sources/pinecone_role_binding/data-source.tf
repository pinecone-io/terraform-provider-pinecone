terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}

# Look up a single role binding by ID. The pinecone_role_bindings list cannot
# filter on a binding's own ID, so use this when you already have one.
data "pinecone_role_binding" "example" {
  id = "your-role-binding-id"
}

output "role" {
  value = data.pinecone_role_binding.example.role
}
