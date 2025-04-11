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
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccIndexDataSourceConfig_serverless(rName),
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

func testAccIndexDataSourceConfig_serverless(name string) string {
	return fmt.Sprintf(`
	provider "pinecone" {
	}
	
	resource "pinecone_index" "test" {
		name = %q
		spec = {
		    serverless = {
		        cloud = "aws"
			    region = "us-west-2"
		    }
		}
		tags = {
			test = "test1"
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
