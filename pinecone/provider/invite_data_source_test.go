package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccInviteDataSource creates a pending invite and reads it back by ID.
// CheckDestroy confirms the invite is revoked from the API afterwards.
//
// The accepted (processed) invite path — the gap this data source exists to
// close, since the invites list omits processed invites — cannot be automated
// because acceptance requires a human clicking through the invite email. It is
// verified manually.
func TestAccInviteDataSource(t *testing.T) {
	t.Parallel()
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_CLIENT_ID and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	// A unique address avoids colliding with an existing pending invite or member.
	email := fmt.Sprintf("tf-acc-invite-ds-%d@example.com", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInviteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccInviteDataSourceConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.pinecone_invite.test", "id", "pinecone_invite.test", "id"),
					resource.TestCheckResourceAttrPair("data.pinecone_invite.test", "created_at", "pinecone_invite.test", "created_at"),
					resource.TestCheckResourceAttr("data.pinecone_invite.test", "email", email),
					resource.TestCheckResourceAttr("data.pinecone_invite.test", "status", "pending"),
					resource.TestCheckResourceAttrSet("data.pinecone_invite.test", "expires_at"),
					// Not yet accepted, so there is no processed timestamp.
					resource.TestCheckNoResourceAttr("data.pinecone_invite.test", "processed_at"),
				),
			},
		},
	})
}

// TestAccInviteRoleBindingsReadable checks the claim the invite docs make: the
// roles granted at invite time are not returned by the invite endpoint, but are
// readable from the role bindings list under principal_type = "invite".
//
// This is the only automated check of that claim. If it fails, the guidance on
// pinecone_invite.role_bindings is wrong and needs revising, not the test.
func TestAccInviteRoleBindingsReadable(t *testing.T) {
	t.Parallel()
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_CLIENT_ID and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	email := fmt.Sprintf("tf-acc-invite-rb-%d@example.com", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInviteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccInviteRoleBindingsConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The invite was created with exactly one organization-scoped
					// binding, which the role bindings list should surface.
					resource.TestCheckResourceAttr("data.pinecone_role_bindings.for_invite", "role_bindings.#", "1"),
					resource.TestCheckResourceAttr("data.pinecone_role_bindings.for_invite", "role_bindings.0.principal_type", "invite"),
					resource.TestCheckResourceAttr("data.pinecone_role_bindings.for_invite", "role_bindings.0.resource_type", "organization"),
					resource.TestCheckResourceAttr("data.pinecone_role_bindings.for_invite", "role_bindings.0.role", "OrgMember"),
					resource.TestCheckResourceAttrPair("data.pinecone_role_bindings.for_invite", "role_bindings.0.principal_id", "pinecone_invite.test", "id"),
				),
			},
		},
	})
}

func testAccInviteDataSourceConfig(email string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_invite" "test" {
  email = %[3]q

  role_bindings = [
    {
      resource_type = "organization"
      role          = "OrgMember"
    }
  ]
}

data "pinecone_invite" "test" {
  id = pinecone_invite.test.id
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), email)
}

func testAccInviteRoleBindingsConfig(email string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_invite" "test" {
  email = %[3]q

  role_bindings = [
    {
      resource_type = "organization"
      role          = "OrgMember"
    }
  ]
}

data "pinecone_role_bindings" "for_invite" {
  principal_type = "invite"
  principal_id   = pinecone_invite.test.id
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), email)
}
