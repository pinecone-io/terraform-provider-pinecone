// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCollectionResource(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCollectionResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_collection.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_collection.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_collection.test", "source", rName),
					resource.TestCheckResourceAttrSet("pinecone_collection.test", "size"),
					resource.TestCheckResourceAttrSet("pinecone_collection.test", "status"),
				),
			},
			// Verify the data source reads back the same collection without creating new infrastructure.
			{
				Config: testAccCollectionResourceWithDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_collection.test", "id", rName),
					resource.TestCheckResourceAttr("data.pinecone_collection.test", "name", rName),
					resource.TestCheckResourceAttr("data.pinecone_collection.test", "dimension", "1536"),
					resource.TestCheckResourceAttrSet("data.pinecone_collection.test", "size"),
					resource.TestCheckResourceAttrSet("data.pinecone_collection.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "pinecone_collection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCollectionResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
	name = %q
	dimension = 1536
	spec = {
		pod = {
			environment = "us-west4-gcp"
			pod_type = "s1.x1"
		}
	}
}
  
resource "pinecone_collection" "test" {
	name = %q
	source = pinecone_index.test.name
}
`, name, name)
}

func testAccCollectionResourceWithDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
	name = %q
	dimension = 1536
	spec = {
		pod = {
			environment = "us-west4-gcp"
			pod_type = "s1.x1"
		}
	}
}
  
resource "pinecone_collection" "test" {
	name = %q
	source = pinecone_index.test.name
}

data "pinecone_collection" "test" {
	name = pinecone_collection.test.name
}
`, name, name)
}
