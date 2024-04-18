// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApiKeyResource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tftest")
	rShortName := fmt.Sprintf("%s%d", "tf", sdkacctest.RandIntRange(0, 9999))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApiKeyResourceConfig(rName, rShortName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pinecone_project_api_key.test", "id"),
					resource.TestCheckResourceAttr("pinecone_project_api_key.test", "name", rShortName),
					resource.TestCheckResourceAttrSet("pinecone_project_api_key.test", "secret"),
				),
			},
			// ImportState testing
			// {
			// 	ResourceName:      "pinecone_project.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
			// Update not supported for serverless specs
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccApiKeyResourceConfig(name string, shortname string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_project" "test" {
  name = %q
}

resource "pinecone_project_api_key" "test" {
  name = %q
  project_id = pinecone_project.test.id
}

`, name, shortname)
}
