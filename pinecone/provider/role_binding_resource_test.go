package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleBindingResource(t *testing.T) {
	t.Parallel()
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
				Config: testAccRoleBindingResourceConfig(projectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "principal_type", "api_key"),
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "resource_type", "project"),
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "resource_id", projectId),
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "role", "ProjectViewer"),
					resource.TestCheckResourceAttrSet("pinecone_role_binding.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_role_binding.test", "created_at"),
					resource.TestCheckResourceAttrPair("pinecone_role_binding.test", "principal_id", "pinecone_api_key.test", "id"),
				),
			},
			{
				ResourceName:      "pinecone_role_binding.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoleBindingResourceConfig(projectId string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_api_key" "test" {
  name       = "test-role-binding-key"
  project_id = %[3]q
}

resource "pinecone_role_binding" "test" {
  principal_id   = pinecone_api_key.test.id
  principal_type = "api_key"
  resource_type  = "project"
  resource_id    = %[3]q
  role           = "ProjectViewer"
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), projectId)
}
