// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccIndexResource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "512"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pods", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "replicas", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pod_type", "starter"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "source_collection"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pinecone_index.test",
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIdentifierAttribute: "name",
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				// ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				// TODO: update replicas test. Cannot do this currently in the free-tier.
				Config: testAccIndexResourceConfig(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "512"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pods", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "replicas", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pod_type", "starter"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "source_collection"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccIndexResourceConfig(name string, replicas int) string {
	return fmt.Sprintf(`
provider "pinecone" {
	environment = "gcp-starter"
}

resource "pinecone_index" "test" {
  name = %q
  dimension = 512
  replicas = %d
}
`, name, replicas)
}
