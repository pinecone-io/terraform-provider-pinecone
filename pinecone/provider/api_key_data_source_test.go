// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApiKeyDataSource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tftest")
	rShortName := fmt.Sprintf("%s%d", "tf", sdkacctest.RandIntRange(0, 9999))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccApiKeyDataSourceConfig(rName, rShortName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_project_api_key.test", "id"),
					// resource.TestCheckResourceAttr("data.pinecone_project_api_key.test", "name", rShortName),
				),
			},
		},
	})
}

func testAccApiKeyDataSourceConfig(name string, rShortName string) string {
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
	
	data "pinecone_project_api_key" "test" {
		name = pinecone_project_api_key.test.name
		project_id = pinecone_project.test.id
	}
	`, name, rShortName)
}
