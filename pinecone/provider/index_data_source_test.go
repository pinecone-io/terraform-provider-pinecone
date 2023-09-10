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
					resource.TestCheckResourceAttr("data.pinecone_index.test", "id", "frank"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "name", "frank"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "dimension", "512"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "metric", "cosine"),
				),
			},
		},
	})
}

func testAccIndexDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	provider "pinecone" {
		environment = "gcp-starter"
	}
	
	resource "pinecone_index" "test" {
		name = %q
		dimension = 512
		metric = "cosine"
	}
	
	data "pinecone_index" "test" {
		name = "%s"

		depends_on = [pinecone_index.test]
	}
	`, name, name)
}
