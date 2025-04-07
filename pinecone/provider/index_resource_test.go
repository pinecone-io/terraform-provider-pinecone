// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func TestAccIndexResource_serverless_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_serverless(rName, "enabled", map[string]string{"test": "testval", "remove": "testremove", "update": "testupdate"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.serverless.region", "us-west-2"),
				),
			},
			// Disable deletion_protection, update tags
			{
				Config: testAccIndexResourceConfig_serverless(rName, "disabled", map[string]string{"test": "testval", "update": "testupdatenew"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.serverless.region", "us-west-2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pinecone_index.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update not supported for serverless specs
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccIndexResource_pod_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_pod(rName, "enabled", map[string]string{"test": "testval", "remove": "testremove", "update": "testupdate"}, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "deletion_protection", "enabled"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pod_type", "s1.x1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.replicas", "2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pods", "2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.%", "3"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.test", "testval"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.remove", "testremove"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.update", "testupdate"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.test", "testval"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "metadata_config"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "source_collection"),
				),
			},
			// Disable deletion_protection, update tags
			{
				Config: testAccIndexResourceConfig_pod(rName, "disabled", map[string]string{"test": "testval", "update": "testupdatenew"}, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "deletion_protection", "disabled"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pod_type", "s1.x1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.replicas", "2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pods", "2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.test", "testval"),
					resource.TestCheckResourceAttr("pinecone_index.test", "tags.update", "testupdatenew"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "metadata_config"),
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
		},
	})
}

func TestAccIndexResource_dimension(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_serverless(rName, "disabled", nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccIndexResource_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")
	resourceName := "pinecone_index.test"

	var index pinecone.Index

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(rName),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_serverless(rName, "disabled", nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
				),
			},
		},
	})
}

func testAccCheckIndexExists(resourceName string, index *pinecone.Index) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		indexResource, found := state.RootModule().Resources[resourceName]
		if !found {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := NewTestClient()

		fetchedIndex, err := c.DescribeIndex(context.Background(), indexResource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error describing index: %w", err)
		}
		if fetchedIndex == nil {
			return fmt.Errorf("Index not found for ID: %s", indexResource.Primary.ID)
		}

		*index = *fetchedIndex

		return nil
	}
}

func testAccCheckIndexDestroy(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, found := state.RootModule().Resources[resourceName]
		if found {
			return fmt.Errorf("Resource still found in state: %s", resourceName)
		}
		return nil
	}
}

func testAccIndexResourceConfig_serverless(name string, deletionProtection string, tags map[string]string) string {
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
  deletion_protection = %q
%s
}
`, name, deletionProtection, convertMapToString(tags))
}

func testAccIndexResourceConfig_pod(name string, deletionProtection string, tags map[string]string, replicas string) string {
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
			replicas = %q
		}
	}
	deletion_protection = %q
%s
}
`, name, replicas, deletionProtection, convertMapToString(tags))
}

func convertMapToString(in map[string]string) string {
	var mapStr string
	if len(in) > 0 {
		mapStr = "  tags = {\n"
		for k, v := range in {
			mapStr += fmt.Sprintf("    %q = %q\n", k, v)
		}
		mapStr += "  }\n"
	}
	return mapStr
}
