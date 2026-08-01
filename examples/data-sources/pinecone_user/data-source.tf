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

# Look up a user by email (exactly one of id or email must be set).
data "pinecone_user" "by_email" {
  email = "teammate@example.com"
}

# Or look up a user by ID.
data "pinecone_user" "by_id" {
  id = "your-user-id"
}

output "user_id" {
  value = data.pinecone_user.by_email.id
}
