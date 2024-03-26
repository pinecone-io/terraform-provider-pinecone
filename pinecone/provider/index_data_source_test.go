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
					resource.TestCheckResourceAttr("data.pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "metric", "cosine"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.cloud", "aws"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "spec.serverless.region", "us-west-2"),
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
	}
	
	data "pinecone_index" "test" {
		name = pinecone_index.test.name
	}
	`, name)
}
