package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApiKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApiKeyResourceConfig("test-api-key", "test-project-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_api_key.test", "name", "test-api-key"),
					resource.TestCheckResourceAttr("pinecone_api_key.test", "project_id", "test-project-id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.test", "key"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pinecone_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Note: ImportStateVerify is set to true, but the key attribute
				// will not be available during import for security reasons
				ImportStateVerifyIgnore: []string{"key"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccApiKeyResourceConfig(name, projectId string) string {
	return fmt.Sprintf(`
resource "pinecone_api_key" "test" {
  name       = %[1]q
  project_id = %[2]q
}
`, name, projectId)
}

func TestAccApiKeyResource_requiresAdminClient(t *testing.T) {
	// Test that the resource requires admin client credentials
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApiKeyResourceConfig("test-api-key", "test-project-id"),
				ExpectError: regexp.MustCompile("Admin client not configured"),
			},
		},
	})
} 