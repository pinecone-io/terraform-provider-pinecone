// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const publishedProviderVersions = ">= 3.0.0, < 5.0.0"

// TestAccIndexResource_upgradeFromPublished creates indexes with the latest published provider,
// then checks the current provider refreshes and plans them with no changes.
func TestAccIndexResource_upgradeFromPublished(t *testing.T) {
	// The published provider lives at pinecone-io/pinecone; the in-process provider must use the
	// same address to take over its state. t.Setenv rules out t.Parallel.
	t.Setenv(resource.EnvTfAccProviderNamespace, "pinecone-io")

	tests := []struct {
		name   string
		config func(name string) string
	}{
		{"dense", func(name string) string {
			return testAccIndexUpgradeConfig(name, "dimension = 1024", "")
		}},
		{"sparse", func(name string) string {
			return testAccIndexUpgradeConfig(name, `metric = "dotproduct"
  vector_type = "sparse"`, "")
		}},
		{"metadata schema", func(name string) string {
			return testAccIndexUpgradeConfig(name, "dimension = 1024", `
      schema = {
        fields = {
          genre = { filterable = true }
        }
      }`)
		}},
		{"dedicated read capacity", func(name string) string {
			return testAccIndexUpgradeConfig(name, "dimension = 1024", `
      read_capacity = {
        dedicated = {
          node_type = "b1"
          replicas  = 1
          shards    = 1
        }
      }`)
		}},
		{"integrated", func(name string) string {
			return testAccIndexUpgradeConfig(name, `embed = {
    model = "multilingual-e5-large"
    field_map = {
      text = "chunk_text"
    }
    read_parameters = {
      input_type = "query"
    }
  }`, "")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config(acctest.RandomWithPrefix("tftest"))
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				CheckDestroy: testAccCheckIndexDestroy(),
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"pinecone": {Source: "pinecone-io/pinecone", VersionConstraint: publishedProviderVersions},
						},
						Config: config,
					},
					{
						ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
						Config:                   config,
						PlanOnly:                 true,
					},
				},
			})
		})
	}
}

func testAccIndexUpgradeConfig(name, attributes, serverless string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name = %q
  %s
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"%s
    }
  }
  tags = {}
}
`, name, attributes, serverless)
}
