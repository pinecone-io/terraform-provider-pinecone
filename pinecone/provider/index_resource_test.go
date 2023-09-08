// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig("frank2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "name", "frank2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "512"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
				),
			},
			// ImportState testing
			// {
			// 	ResourceName:      "pinecone_example.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	// This is not normally necessary, but is here because this
			// 	// example code does not have an actual upstream service.
			// 	// Once the Read method is able to refresh information from
			// 	// the upstream service, this can be removed.
			// 	ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			// },
			// Update and Read testing
			// {
			// 	Config: testAccIndexResourceConfig("frank2"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("pinecone_example.test", "configurable_attribute", "two"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccIndexResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "pinecone_index" "test" {
  name = %[1]q
  dimension = 512
  metric = "cosine"
}
`, name)
}
