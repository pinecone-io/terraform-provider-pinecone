package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectsDataSource(t *testing.T) {
	// Test listing all projects
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if clientId == "" {
		t.Skip("PINECONE_CLIENT_ID environment variable is required for this test")
	}
	if clientSecret == "" {
		t.Skip("PINECONE_CLIENT_SECRET environment variable is required for this test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "pinecone" {
						client_id     = "%s"
						client_secret = "%s"
					}

					data "pinecone_projects" "test" {}
				`, clientId, clientSecret),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_projects.test", "id"),
					resource.TestCheckResourceAttrSet("data.pinecone_projects.test", "projects.#"),            // Should have projects
					resource.TestCheckResourceAttrSet("data.pinecone_projects.test", "projects.0.created_at"), // Check first project has created_at
				),
			},
		},
	})
}
