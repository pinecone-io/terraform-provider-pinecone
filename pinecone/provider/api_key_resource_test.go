package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApiKeyResource(t *testing.T) {
	projectId := os.Getenv("PINECONE_PROJECT_ID")
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if projectId == "" || clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_PROJECT_ID, PINECONE_CLIENT_ID, and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfig(projectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test default roles
					resource.TestCheckResourceAttr("pinecone_api_key.default", "name", "test-api-key-default"),
					resource.TestCheckResourceAttr("pinecone_api_key.default", "project_id", projectId),
					resource.TestCheckResourceAttrSet("pinecone_api_key.default", "id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.default", "key"),
					resource.TestCheckResourceAttr("pinecone_api_key.default", "roles.#", "1"),
				),
			},
		},
	})
}

func TestAccApiKeyResourceWithCustomRoles(t *testing.T) {
	projectId := os.Getenv("PINECONE_PROJECT_ID")
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if projectId == "" || clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_PROJECT_ID, PINECONE_CLIENT_ID, and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceWithCustomRolesConfig(projectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test custom roles
					resource.TestCheckResourceAttr("pinecone_api_key.custom", "name", "test-api-key-custom"),
					resource.TestCheckResourceAttr("pinecone_api_key.custom", "project_id", projectId),
					resource.TestCheckResourceAttrSet("pinecone_api_key.custom", "id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.custom", "key"),
					resource.TestCheckResourceAttr("pinecone_api_key.custom", "roles.#", "2"),
				),
			},
		},
	})
}

func TestAccApiKeyResourceUpdate(t *testing.T) {
	projectId := os.Getenv("PINECONE_PROJECT_ID")
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if projectId == "" || clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_PROJECT_ID, PINECONE_CLIENT_ID, and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceUpdateConfig(projectId, "test-api-key-update", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_api_key.update_test", "name", "test-api-key-update"),
					resource.TestCheckResourceAttr("pinecone_api_key.update_test", "project_id", projectId),
					resource.TestCheckResourceAttrSet("pinecone_api_key.update_test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.update_test", "key"),
					resource.TestCheckResourceAttr("pinecone_api_key.update_test", "roles.#", "1"),
				),
			},
			{
				Config: testAccApiKeyResourceUpdateConfig(projectId, "test-api-key-updated", `roles = ["ProjectViewer", "DataPlaneViewer"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_api_key.update_test", "name", "test-api-key-updated"),
					resource.TestCheckResourceAttr("pinecone_api_key.update_test", "project_id", projectId),
					resource.TestCheckResourceAttrSet("pinecone_api_key.update_test", "id"),
					resource.TestCheckResourceAttr("pinecone_api_key.update_test", "roles.#", "2"),
				),
			},
		},
	})
}

func testAccApiKeyResourceConfig(projectId string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

# Test API key with default roles (ProjectEditor)
resource "pinecone_api_key" "default" {
  name       = "test-api-key-default"
  project_id = %[3]q
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), projectId)
}

func testAccApiKeyResourceWithCustomRolesConfig(projectId string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

# Test API key with custom roles
resource "pinecone_api_key" "custom" {
  name       = "test-api-key-custom"
  project_id = %[3]q
  roles      = ["ProjectViewer", "DataPlaneViewer"]
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), projectId)
}

func testAccApiKeyResourceUpdateConfig(projectId, name, roles string) string {
	rolesConfig := ""
	if roles != "" {
		rolesConfig = fmt.Sprintf("\n  %s", roles)
	}

	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

# Test API key with update functionality
resource "pinecone_api_key" "update_test" {
  name       = %[3]q
  project_id = %[4]q%[5]s
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), name, projectId, rolesConfig)
}
