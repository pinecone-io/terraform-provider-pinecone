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

# List all service accounts in your organization. Client secrets are never returned.
data "pinecone_service_accounts" "all" {}

output "service_account_names" {
  value = [for sa in data.pinecone_service_accounts.all.service_accounts : sa.name]
}
