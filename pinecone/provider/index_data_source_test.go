// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexDataSource_serverless(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccIndexDataSourceConfig_serverlessIntegrated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "dimension", "1024"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.region", "us-west-2"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "deletion_protection", "disabled"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "tags.test", "test1"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "tags.test2", "test2"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "embed.model", "multilingual-e5-large"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "embed.field_map.%", "1"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "embed.field_map.text", "chunk_text"),
				),
			},
		},
	})
}

func TestAccIndexDataSource_serverless_plain(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexDataSourceConfig_serverless(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "dimension", "1024"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.region", "us-west-2"),
					// Non-integrated index: embed and schema children must be absent.
					resource.TestCheckNoResourceAttr("data.pinecone_index.test", "embed.model"),
					resource.TestCheckNoResourceAttr("data.pinecone_index.test", "spec.serverless.schema.fields.%"),
				),
			},
		},
	})
}

func TestAccIndexDataSource_serverless_withSchema(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexDataSourceConfig_serverlessWithSchema(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "dimension", "1024"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.region", "us-west-2"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.schema.fields.%", "2"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.schema.fields.genre.filterable", "true"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.schema.fields.year.filterable", "true"),
				),
			},
		},
	})
}

func TestAccIndexDataSource_serverless_withReadCapacity(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexDataSourceConfig_serverlessWithReadCapacity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "dimension", "1024"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.region", "us-west-2"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.read_capacity.dedicated.node_type", "b1"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.read_capacity.dedicated.replicas", "1"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.read_capacity.dedicated.shards", "1"),
				),
			},
		},
	})
}

func testAccIndexDataSourceConfig_serverlessIntegrated(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name = %q
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
    }
  }
  tags = {
    test  = "test1"
    test2 = "test2"
  }
  embed = {
    model = "multilingual-e5-large"
    field_map = {
      text = "chunk_text"
    }
  }
}

data "pinecone_index" "test" {
  name = pinecone_index.test.name
}
`, name)
}

func testAccIndexDataSourceConfig_serverless(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name      = %q
  dimension = 1024
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
    }
  }
  deletion_protection = "disabled"
}

data "pinecone_index" "test" {
  name = pinecone_index.test.name
}
`, name)
}

func testAccIndexDataSourceConfig_serverlessWithSchema(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name      = %q
  dimension = 1024
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
      schema = {
        fields = {
          "genre" = { filterable = true }
          "year"  = { filterable = true }
        }
      }
    }
  }
  deletion_protection = "disabled"
}

data "pinecone_index" "test" {
  name = pinecone_index.test.name
}
`, name)
}

func testAccIndexDataSourceConfig_serverlessWithReadCapacity(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name      = %q
  dimension = 1024
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
      read_capacity = {
        dedicated = {
          node_type = "b1"
          replicas  = 1
          shards    = 1
        }
      }
    }
  }
  deletion_protection = "disabled"
}

data "pinecone_index" "test" {
  name = pinecone_index.test.name
}
`, name)
}
