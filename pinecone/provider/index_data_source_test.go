// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexDataSource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccIndexDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_index.test", "id", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "name", rName),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "dimension", "512"),
					resource.TestCheckResourceAttr("data.pinecone_index.test", "metric", "cosine"),
				),
			},
		},
	})
}

func testAccIndexDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	provider "pinecone" {
		environment = "us-west4-gcp"
	}
	
	resource "pinecone_index" "test" {
		name = %q
		dimension = 512
		metric = "cosine"
		pod_type = "s1.x1"
	}
	
	data "pinecone_index" "test" {
		name = pinecone_index.test.name
	}
	`, name)
}
