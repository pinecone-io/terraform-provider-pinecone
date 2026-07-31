package provider

import (
	"fmt"
	"os"
	"regexp"
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

// TestAccRoleBindingResource_scopeValidation exercises the plan-time
// ConfigValidator. It needs no credentials: validation rejects the config before
// the provider makes any API call, so the dummy credentials below are never used.
func TestAccRoleBindingResource_scopeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "organization scope rejects resource_id",
			config: `
resource "pinecone_role_binding" "test" {
  principal_id   = "00000000-0000-0000-0000-000000000000"
  principal_type = "user"
  resource_type  = "organization"
  resource_id    = "00000000-0000-0000-0000-000000000001"
  role           = "OrgMember"
}`,
			expectError: regexp.MustCompile("Unexpected resource_id"),
		},
		{
			name: "project scope requires resource_id",
			config: `
resource "pinecone_role_binding" "test" {
  principal_id   = "00000000-0000-0000-0000-000000000000"
  principal_type = "user"
  resource_type  = "project"
  role           = "ProjectViewer"
}`,
			expectError: regexp.MustCompile("Missing resource_id"),
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
