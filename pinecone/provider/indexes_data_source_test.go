// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexesDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccIndexesDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_indexes.test", "id"),
				),
			},
		},
	})
}

func testAccIndexesDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	provider "pinecone" {
	}
	
	resource "pinecone_index" "test" {
		name = %q
		dimension = 1536
		spec = {
		    serverless = {
		        cloud = "aws"
			    region = "us-west-2"
		    }
		}
	}
	
	data "pinecone_indexes" "test" {
	}
	`, name)
}
