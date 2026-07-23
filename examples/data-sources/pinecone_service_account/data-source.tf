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

# Look up a single service account by ID. The client secret is never returned.
data "pinecone_service_account" "example" {
  id = "your-service-account-id"
}

output "service_account_name" {
  value = data.pinecone_service_account.example.name
}
