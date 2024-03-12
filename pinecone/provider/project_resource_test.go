// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pinecone_project.test", "id"),
					resource.TestCheckResourceAttr("pinecone_project.test", "name", rName),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pinecone_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update not supported for serverless specs
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProjectResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_project" "test" {
  name = %q
}
`, name)
}
