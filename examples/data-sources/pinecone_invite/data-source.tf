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

# Look up a single invite by ID. Unlike the pinecone_invites list, which returns
# only pending and expired invites, this also returns an accepted ("processed")
# invite.
data "pinecone_invite" "example" {
  id = "your-invite-id"
}

output "invite_status" {
  value = data.pinecone_invite.example.status
}

# The granted roles are not returned by the invite endpoint. Read them from the
# role bindings list instead.
data "pinecone_role_bindings" "for_invite" {
  principal_type = "invite"
  principal_id   = data.pinecone_invite.example.id
}
