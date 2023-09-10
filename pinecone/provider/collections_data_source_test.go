package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCollectionsDataSource(t *testing.T) {
	t.Skip("temporarily skipped...")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pinecone_collections.test", "id"),
					resource.TestCheckResourceAttrSet("data.pinecone_collections.test", "names.#"),
				),
			},
		},
	})
}

const testAccCollectionsDataSourceConfig = `
data "pinecone_collections" "test" {
}
`
