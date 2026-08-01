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

# Users cannot be created through Terraform — they join the organization by
# accepting a pinecone_invite. This resource manages an EXISTING user so that
# destroying it removes the user from the organization.
#
# Bring an existing user under management by importing it:
#   terraform import pinecone_user.example <user-id>
resource "pinecone_user" "example" {
  id = "your-user-id"
}
