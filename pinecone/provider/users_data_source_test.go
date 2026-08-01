package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUsersDataSource is a non-destructive smoke test: the organization always
// has at least the caller as a member, so the list must read without error and
// expose the users attribute.
func TestAccUsersDataSource(t *testing.T) {
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
				Config: testAccUsersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_users.all", "users.#"),
					resource.TestCheckResourceAttrSet("data.pinecone_users.all", "id"),
				),
			},
		},
	})
}

func testAccUsersDataSourceConfig() string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

data "pinecone_users" "all" {}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"))
}
