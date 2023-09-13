package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCollectionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_collections.test", "id"),
				),
			},
		},
	})
}

const testAccCollectionsDataSourceConfig = `
provider "pinecone" {
	environment = "us-west4-gcp"
}

data "pinecone_collections" "test" {
}
`
