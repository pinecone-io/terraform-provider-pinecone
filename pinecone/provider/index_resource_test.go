// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const providerName = "pinecone_index"
const resourceName = "test"
const resourceAddress = providerName + "." + resourceName

func TestAccIndexResource_serverless_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_serverless(rName, "enabled", map[string]string{"test": "testval", "remove": "testremove", "update": "testupdate"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
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
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("pinecone_index.test", "spec.serverless.region", "us-west-2"),
				),
			},
			// Convert to integrated inference
			// ImportState testing
			{
				ResourceName:      "pinecone_index.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccIndexResource_pod_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIndexResourceConfig_pod(rName, "enabled", map[string]string{"test": "testval", "remove": "testremove", "update": "testupdate"}, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
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
					testAccCheckIndexExists(),
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

func TestAccIndexResource_pod_invalidEmbedConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "pinecone_index" "test" {
  name = "test"
  dimension = 1024
  metric = "cosine"
  spec = {
	pod = {
		environment = "us-west4-gcp"
		pod_type = "s1.x1"
	}
  }
  embed = {
    model = "multilingual-e5-large"
	field_map = {
		text = "chunk_text"
	}
  }
}`,
			ExpectError: regexp.MustCompile("Pod-based indexes cannot have an embed configuration."),
		}},
	})
}

func TestAccIndexResource_pod_invalidDimension(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "pinecone_index" "test" {
  name = "test"
  metric = "cosine"
  spec = {
	pod = {
		environment = "us-west4-gcp"
		pod_type = "s1.x1"
	}
  }
}`,
			ExpectError: regexp.MustCompile("Pod-based indexes must have a dimension."),
		}},
	})
}

func TestAccIndexResource_pod_invalidVectorType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "pinecone_index" "test" {
  name = "test"
  dimension = 1024
  metric = "cosine"
  spec = {
	pod = {
		environment = "us-west4-gcp"
		pod_type = "s1.x1"
	}
  }
  vector_type = "sparse"
}`,
			ExpectError: regexp.MustCompile("Pod-based indexes cannot have a sparse vector_type."),
		}},
	})
}

func testAccCheckIndexExists() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		indexResource, found := state.RootModule().Resources[resourceAddress]
		if !found {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		c, _ := NewTestClient()
		fetchedIndex, err := c.DescribeIndex(context.Background(), indexResource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error describing index: %w", err)
		}
		if fetchedIndex == nil {
			return fmt.Errorf("Index not found for ID: %s", indexResource.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIndexDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, found := state.RootModule().Resources[resourceAddress]
		if !found {
			// If the terraform resource is not found we can assume it's destroyed
			return nil
		}

		c, _ := NewTestClient()
		_, err := c.DescribeIndex(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Index still exists in backend: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccIndexResourceConfig_serverless(name string, deletionProtection string, tags map[string]string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "%s" {
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
`, resourceName, name, deletionProtection, convertMapToString(tags))
}

func testAccIndexResourceConfig_pod(name string, deletionProtection string, tags map[string]string, replicas string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "%s" {
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
`, resourceName, name, replicas, deletionProtection, convertMapToString(tags))
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
