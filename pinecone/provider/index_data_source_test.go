// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccIndexDataSourceConfig("frank"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_index.test", "id"),
				),
			},
		},
	})
}

func testAccIndexDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	# resource "pinecone_index" "test" {
	#	name = "%s"
	# }
	
	data "pinecone_index" "test" {
		name = "%s"
	}
	`, name, name)
}
