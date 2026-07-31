package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRoleBindingDataSource creates a project-scoped binding and reads it back
// by ID, which is the lookup the pinecone_role_bindings list cannot express.
// CheckDestroy confirms the binding is gone from the API afterwards.
//
// Only project scope is covered. An organization-scoped binding would also
// exercise the resource_id divergence documented on both schemas, but creating
// and deleting org-level grants risks tripping the API's last-owner and
// last-membership conflicts, so it is verified manually instead.
func TestAccRoleBindingDataSource(t *testing.T) {
	t.Parallel()
	projectId := os.Getenv("PINECONE_PROJECT_ID")
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if projectId == "" || clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_PROJECT_ID, PINECONE_CLIENT_ID, and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	// The acceptance workflow runs a five-version Terraform matrix concurrently
	// against the same organization, so a fixed key name would collide across
	// jobs.
	keyName := fmt.Sprintf("test-rb-ds-key-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoleBindingDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleBindingDataSourceConfig(projectId, keyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The data source must report the same binding the resource created.
					resource.TestCheckResourceAttrPair("data.pinecone_role_binding.test", "id", "pinecone_role_binding.test", "id"),
					resource.TestCheckResourceAttrPair("data.pinecone_role_binding.test", "principal_id", "pinecone_role_binding.test", "principal_id"),
					resource.TestCheckResourceAttrPair("data.pinecone_role_binding.test", "created_at", "pinecone_role_binding.test", "created_at"),
					resource.TestCheckResourceAttr("data.pinecone_role_binding.test", "principal_type", "api_key"),
					resource.TestCheckResourceAttr("data.pinecone_role_binding.test", "resource_type", "project"),
					resource.TestCheckResourceAttr("data.pinecone_role_binding.test", "resource_id", projectId),
					resource.TestCheckResourceAttr("data.pinecone_role_binding.test", "role", "ProjectViewer"),
				),
			},
		},
	})
}

func testAccRoleBindingDataSourceConfig(projectId, keyName string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_api_key" "test" {
  name       = %[4]q
  project_id = %[3]q
}

resource "pinecone_role_binding" "test" {
  principal_id   = pinecone_api_key.test.id
  principal_type = "api_key"
  resource_type  = "project"
  resource_id    = %[3]q
  role           = "ProjectViewer"
}

data "pinecone_role_binding" "test" {
  id = pinecone_role_binding.test.id
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), projectId, keyName)
}
