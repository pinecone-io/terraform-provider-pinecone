package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUserResourceCreateErrors verifies the delete-only guard: applying a
// pinecone_user block (rather than importing) must fail with a clear error and
// never attempt to provision a user. This is non-destructive.
func TestAccUserResourceCreateErrors(t *testing.T) {
	t.Parallel()
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_CLIENT_ID and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserResourceConfig("00000000-0000-0000-0000-000000000000"),
				ExpectError: regexp.MustCompile("Users cannot be created"),
			},
		},
	})
}

func testAccUserResourceConfig(id string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_user" "test" {
  id = %[3]q
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), id)
}
