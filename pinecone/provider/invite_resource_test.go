package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInviteResource(t *testing.T) {
	t.Parallel()
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_CLIENT_ID and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	// A unique address avoids colliding with an existing pending invite or member.
	email := fmt.Sprintf("tf-acc-invite-%d@example.com", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInviteResourceConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_invite.test", "email", email),
					resource.TestCheckResourceAttr("pinecone_invite.test", "status", "pending"),
					resource.TestCheckResourceAttr("pinecone_invite.test", "role_bindings.#", "1"),
					resource.TestCheckResourceAttr("pinecone_invite.test", "role_bindings.0.resource_type", "organization"),
					resource.TestCheckResourceAttr("pinecone_invite.test", "role_bindings.0.role", "OrgMember"),
					resource.TestCheckResourceAttrSet("pinecone_invite.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_invite.test", "created_at"),
					resource.TestCheckResourceAttrSet("pinecone_invite.test", "expires_at"),
				),
			},
		},
	})
}

func testAccInviteResourceConfig(email string) string {
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
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), email)
}
