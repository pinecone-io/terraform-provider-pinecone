// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig("test-project"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_project.test", "name", "test-project"),
					resource.TestCheckResourceAttrSet("pinecone_project.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_project.test", "organization_id"),
				),
			},
		},
	})
}

func testAccProjectResourceConfig(name string) string {
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
  name = "` + name + `"
}
`
}
