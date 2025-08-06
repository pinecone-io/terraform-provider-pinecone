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
