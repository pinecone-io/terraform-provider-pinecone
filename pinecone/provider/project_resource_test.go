package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig("test-project"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_project.test", "name", "test-project"),
					resource.TestCheckResourceAttrSet("pinecone_project.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_project.test", "organization_id"),
					resource.TestCheckResourceAttrSet("pinecone_project.test", "created_at"),
					resource.TestCheckResourceAttr("pinecone_project.test", "force_encryption_with_cmek", "false"),
					resource.TestCheckResourceAttr("pinecone_project.test", "max_pods", "20"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pinecone_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccProjectResourceConfig("updated-test-project"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_project.test", "name", "updated-test-project"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccProjectResourceWithOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with options
			{
				Config: testAccProjectResourceWithOptionsConfig("test-project-with-options", false, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_project.test", "name", "test-project-with-options"),
					resource.TestCheckResourceAttr("pinecone_project.test", "force_encryption_with_cmek", "false"),
					resource.TestCheckResourceAttr("pinecone_project.test", "max_pods", "5"),
				),
			},
			// Update and Read testing (only update name and max_pods, not CMEK)
			{
				Config: testAccProjectResourceWithOptionsConfig("updated-test-project-with-options", false, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_project.test", "name", "updated-test-project-with-options"),
					resource.TestCheckResourceAttr("pinecone_project.test", "force_encryption_with_cmek", "false"),
					resource.TestCheckResourceAttr("pinecone_project.test", "max_pods", "10"),
				),
			},
		},
	})
}

func testAccProjectResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "pinecone_project" "test" {
  name = %[1]q
}
`, name)
}

func testAccProjectResourceWithOptionsConfig(name string, forceEncryption bool, maxPods int) string {
	return fmt.Sprintf(`
resource "pinecone_project" "test" {
  name                        = %[1]q
  force_encryption_with_cmek  = %[2]t
  max_pods                    = %[3]d
}
`, name, forceEncryption, maxPods)
}

func TestAccProjectResourceWithCMEK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create project with CMEK enabled
			{
				Config: testAccProjectResourceWithCMEKConfig("test-project-with-cmek"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_project.test", "name", "test-project-with-cmek"),
					resource.TestCheckResourceAttr("pinecone_project.test", "force_encryption_with_cmek", "true"),
				),
			},
		},
	})
}

func testAccProjectResourceWithCMEKConfig(name string) string {
	return fmt.Sprintf(`
resource "pinecone_project" "test" {
  name                        = %[1]q
  force_encryption_with_cmek  = true
}
`, name)
}
