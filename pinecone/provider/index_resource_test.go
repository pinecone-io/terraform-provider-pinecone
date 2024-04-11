// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

func TestAccIndexResource_serverless(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_serverless(rName),
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
				Config: testAccIndexResourceConfig_pod_basic(rName, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pod_type", "s1.x1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.replicas", "2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pods", "2"),
					// resource.TestCheckNoResourceAttr("pinecone_index.test", "metadata_config"),
					// resource.TestCheckNoResourceAttr("pinecone_index.test", "source_collection"),
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
			// Update not supported
			// {
			// 	Config: testAccIndexResourceConfig_pod_basic(rName, "2", "2"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
			// 		resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
			// 		resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
			// 		resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
			// 		resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pod_type", "s1.x1"),
			// 		resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.replicas", "2"),
			// 		resource.TestCheckResourceAttr("pinecone_index.test", "spec.pod.pods", "2"),
			// 		// resource.TestCheckNoResourceAttr("pinecone_index.test", "metadata_config"),
			// 		// resource.TestCheckNoResourceAttr("pinecone_index.test", "source_collection"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
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
				Config: testAccIndexResourceConfig_dimension(rName, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1"),
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
		// CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_serverless(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					testAccDeleteIndex(resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccDeleteIndex(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		indexResource, found := state.RootModule().Resources[resourceName]
		if !found {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := NewTestClient()

		err := c.DeleteIndex(context.Background(), indexResource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error deleting index: %w", err)
		}

		ctx := context.TODO()
		deleteTimeout := defaultIndexDeleteTimeout
		err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
			index, err := c.DescribeIndex(ctx, indexResource.Primary.ID)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return nil
				}
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(fmt.Errorf("index not deleted. State: %s", index.Status.State))
		})
		return err
	}
}

func testAccIndexResourceConfig_serverless(name string) string {
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
`, name)
}

func testAccIndexResourceConfig_pod_basic(name string, replicas string) string {
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
}
`, name, replicas)
}

func testAccIndexResourceConfig_dimension(name string, dimension string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name = %q
  dimension = %q
  spec = {
	serverless = {
		cloud = "aws"
		region = "us-west-2"
	}
  }
}
`, name, dimension)
}
