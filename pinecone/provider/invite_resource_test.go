package provider

import (
	"fmt"
	"os"
	"regexp"
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

// TestAccInviteResource_scopeValidation exercises the plan-time ConfigValidator
// over the nested role_bindings list. It needs no credentials: validation rejects
// the config before the provider makes any API call.
func TestAccInviteResource_scopeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "organization binding rejects resource_id",
			config: `
resource "pinecone_invite" "test" {
  email = "teammate@example.com"

  role_bindings = [
    {
      resource_type = "organization"
      role          = "OrgMember"
      resource_id   = "00000000-0000-0000-0000-000000000001"
    }
  ]
}`,
			expectError: regexp.MustCompile("Unexpected resource_id"),
		},
		{
			name: "project binding requires resource_id",
			config: `
resource "pinecone_invite" "test" {
  email = "teammate@example.com"

  role_bindings = [
    {
      resource_type = "organization"
      role          = "OrgMember"
    },
    {
      resource_type = "project"
      role          = "ProjectViewer"
    }
  ]
}`,
			// The offending binding is the second one; the diagnostic must point
			// at that index rather than the list as a whole.
			expectError: regexp.MustCompile(`role_bindings\[1\]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccDummyAdminProviderConfig + tt.config,
						ExpectError: tt.expectError,
					},
				},
			})
		})
	}
}

func TestInviteEmailRegex(t *testing.T) {
	valid := []string{
		"teammate@example.com",
		"first.last@sub.example.co.uk",
		"user+tag@example.io",
	}
	invalid := []string{
		"",
		"not-an-email",
		"missing@domain", // no dot in the domain
		"@example.com",   // no local part
		"user@",          // no domain
		"user @example.com",
		"user@exa mple.com",
		"two@at@example.com",
	}
	for _, e := range valid {
		if !emailRegex.MatchString(e) {
			t.Errorf("expected %q to be accepted", e)
		}
	}
	for _, e := range invalid {
		if emailRegex.MatchString(e) {
			t.Errorf("expected %q to be rejected", e)
		}
	}
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
