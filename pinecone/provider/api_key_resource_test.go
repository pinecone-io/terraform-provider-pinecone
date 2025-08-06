package provider

import (
	"fmt"
	"os"
	"regexp"
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
				Config: testAccApiKeyResourceConfig("test-api-key", projectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_api_key.test", "name", "test-api-key"),
					resource.TestCheckResourceAttr("pinecone_api_key.test", "project_id", projectId),
					resource.TestCheckResourceAttrSet("pinecone_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.test", "key"),
				),
			},
		},
	})
}

func testAccApiKeyResourceConfig(name, projectId string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_api_key" "test" {
  name       = %[3]q
  project_id = %[4]q
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), name, projectId)
}

func TestAccApiKeyResource_requiresAdminClient(t *testing.T) {
	projectId := os.Getenv("PINECONE_PROJECT_ID")
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")
	
	if projectId == "" || clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_PROJECT_ID, PINECONE_CLIENT_ID, and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	// Temporarily unset admin credentials to test the error case when no credentials are provided.
	// This test verifies that the API key resource properly requires admin credentials.
	originalClientId := os.Getenv("PINECONE_CLIENT_ID")
	originalClientSecret := os.Getenv("PINECONE_CLIENT_SECRET")
	os.Unsetenv("PINECONE_CLIENT_ID")
	os.Unsetenv("PINECONE_CLIENT_SECRET")
	defer func() {
		if originalClientId != "" {
			os.Setenv("PINECONE_CLIENT_ID", originalClientId)
		}
		if originalClientSecret != "" {
			os.Setenv("PINECONE_CLIENT_SECRET", originalClientSecret)
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApiKeyResourceConfigWithoutAdmin("test-api-key", projectId),
				ExpectError: regexp.MustCompile("No credentials provided"),
			},
		},
	})
}

func testAccApiKeyResourceConfigWithoutAdmin(name, projectId string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_api_key" "test" {
  name       = %[1]q
  project_id = %[2]q
}
`, name, projectId)
} 