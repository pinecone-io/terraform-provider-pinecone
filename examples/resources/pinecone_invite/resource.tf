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

# role_bindings grants the invitee's initial roles. When the invite is accepted,
# the server moves those bindings to the new user principal — the same bindings,
# reassigned — so the roles carry over without being reissued. From that point on
# manage them with pinecone_role_binding using principal_type = "user".
#
# Invite a user as an organization member.
resource "pinecone_invite" "member" {
  email = "teammate@example.com"

  role_bindings = [
    {
      resource_type = "organization"
      role          = "OrgMember"
    }
  ]
}

# Invite a user with both an organization membership and a project-scoped role.
resource "pinecone_invite" "project_editor" {
  email = "contractor@example.com"

  role_bindings = [
    {
      resource_type = "organization"
      role          = "OrgMember"
    },
    {
      resource_type = "project"
      role          = "ProjectEditor"
      resource_id   = "your-project-id"
    }
  ]
}
