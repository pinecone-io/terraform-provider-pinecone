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

# List the organization's pending and expired invites.
data "pinecone_invites" "all" {}

output "pending_invite_emails" {
  value = [for i in data.pinecone_invites.all.invites : i.email if i.status == "pending"]
}
