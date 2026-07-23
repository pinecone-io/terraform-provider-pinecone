package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceAccountResource(t *testing.T) {
	t.Parallel()
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_CLIENT_ID and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	// Captured across steps to assert secret rotation behavior.
	var secretAfterCreate string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccServiceAccountResourceConfig("test-service-account", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_service_account.test", "name", "test-service-account"),
					resource.TestCheckResourceAttrSet("pinecone_service_account.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_service_account.test", "client_id"),
					resource.TestCheckResourceAttrSet("pinecone_service_account.test", "client_secret"),
					resource.TestCheckResourceAttrSet("pinecone_service_account.test", "created_at"),
					resource.TestCheckResourceAttrWith("pinecone_service_account.test", "client_secret", func(v string) error {
						secretAfterCreate = v
						if v == "" {
							return fmt.Errorf("expected a non-empty client_secret")
						}
						return nil
					}),
				),
			},
			// Rename: the secret must not change.
			{
				Config: testAccServiceAccountResourceConfig("test-service-account-updated", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_service_account.test", "name", "test-service-account-updated"),
					resource.TestCheckResourceAttrWith("pinecone_service_account.test", "client_secret", func(v string) error {
						if v != secretAfterCreate {
							return fmt.Errorf("client_secret changed on a rename; expected it to be preserved")
						}
						return nil
					}),
				),
			},
			// Rotate: changing rotate_trigger must issue a new secret.
			{
				Config: testAccServiceAccountResourceConfig("test-service-account-updated", "rotate-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_service_account.test", "rotate_trigger", "rotate-1"),
					resource.TestCheckResourceAttrWith("pinecone_service_account.test", "client_secret", func(v string) error {
						if v == secretAfterCreate {
							return fmt.Errorf("client_secret was not rotated after rotate_trigger changed")
						}
						return nil
					}),
				),
			},
			// Import (the secret and trigger cannot be recovered from the API).
			{
				ResourceName:            "pinecone_service_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret", "rotate_trigger"},
			},
		},
	})
}

func testAccServiceAccountResourceConfig(name, rotateTrigger string) string {
	rotate := ""
	if rotateTrigger != "" {
		rotate = fmt.Sprintf("\n  rotate_trigger = %q", rotateTrigger)
	}
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_service_account" "test" {
  name = %[3]q%[4]s
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), name, rotate)
}
