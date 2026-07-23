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
