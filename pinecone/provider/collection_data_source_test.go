package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCollectionDataSource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pinecone_collection.test", "id", rName),
					resource.TestCheckResourceAttr("data.pinecone_collection.test", "name", rName),
					resource.TestCheckResourceAttr("data.pinecone_collection.test", "status", "Ready"),
					resource.TestCheckResourceAttr("data.pinecone_collection.test", "dimension", "1536"),
					resource.TestCheckResourceAttrSet("data.pinecone_collection.test", "size"),
					resource.TestCheckResourceAttrSet("data.pinecone_collection.test", "vector_count"),
				),
			},
		},
	})
}

func testAccCollectionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "pinecone" {
	environment = "us-west4-gcp"
}

resource "pinecone_index" "test" {
	name = %q
	spec = {
		pod = {
			environment = "us-west4-gcp"
			pod_type = "s1.x1"
		}
	}
}
  
resource "pinecone_collection" "test" {
	name = %q
	source = pinecone_index.test.name
}

data "pinecone_collection" "test" {
	name = pinecone_collection.test.name
}
`, name, name)
}
