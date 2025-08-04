// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAPIKeyResourceConfig("test-project", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("pinecone_api_key.test", "name", "test-api-key"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.test", "project_id"),
					resource.TestCheckResourceAttrSet("pinecone_api_key.test", "value"),
				),
			},
		},
	})
}

func testAccAPIKeyResourceConfig(projectName, apiKeyName string) string {
	return `
terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

resource "pinecone_project" "test" {
  name = "` + projectName + `"
}

resource "pinecone_api_key" "test" {
  name       = "` + apiKeyName + `"
  project_id = pinecone_project.test.id
}
`
}
