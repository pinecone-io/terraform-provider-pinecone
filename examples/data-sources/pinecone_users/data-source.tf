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

# List all users in your organization.
data "pinecone_users" "all" {}

# Optionally filter by email.
data "pinecone_users" "filtered" {
  email = "teammate@example.com"
}

output "user_emails" {
  value = [for u in data.pinecone_users.all.users : u.email]
}
