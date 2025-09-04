// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectDataSource(t *testing.T) {
	// Test with invalid UUID format first
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "pinecone_project" "test" {
						id = "invalid-uuid-format"
					}
				`,
				ExpectError: regexp.MustCompile("invalid UUID"),
			},
		},
	})
}

func TestAccProjectDataSourceWithRealProject(t *testing.T) {
	// Test with a real project ID if credentials are available
	projectID := os.Getenv("PINECONE_PROJECT_ID")
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if projectID == "" {
		t.Skip("PINECONE_PROJECT_ID environment variable is required for this test")
	}
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

					data "pinecone_project" "test" {
						id = "%s"
					}
				`, clientId, clientSecret, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_project.test", "id", projectID),
					resource.TestCheckResourceAttrSet("data.pinecone_project.test", "name"),
					resource.TestCheckResourceAttrSet("data.pinecone_project.test", "organization_id"),
					resource.TestCheckResourceAttrSet("data.pinecone_project.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.pinecone_project.test", "force_encryption_with_cmek"),
					resource.TestCheckResourceAttrSet("data.pinecone_project.test", "max_pods"),
				),
			},
		},
	})
}
