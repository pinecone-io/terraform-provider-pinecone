// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectDataSource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccProjectDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_project.test", "id"),
					resource.TestCheckResourceAttr("data.pinecone_project.test", "name", rName),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	provider "pinecone" {
	}
	
	resource "pinecone_project" "test" {
		name = %q
	}
	
	data "pinecone_project" "test" {
		name = pinecone_project.test.name
	}
	`, name)
}
