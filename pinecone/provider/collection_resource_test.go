// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccCollectionResource(t *testing.T) {
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
			// ImportState testing
			{
				ResourceName:      "pinecone_collection.test",
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIdentifierAttribute: "name",
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
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
