package provider

import (
	"fmt"
)

// func TestAccCollectionDataSource(t *testing.T) {
// 	rName := acctest.RandomWithPrefix("tftest")

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccCollectionDataSourceConfig(rName),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("data.pinecone_collection.test", "id", rName),
// 					resource.TestCheckResourceAttr("data.pinecone_collection.test", "name", rName),
// 					resource.TestCheckResourceAttr("data.pinecone_collection.test", "status", "Ready"),
// 					resource.TestCheckResourceAttr("data.pinecone_collection.test", "dimension", "1536"),
// 					resource.TestCheckResourceAttrSet("data.pinecone_collection.test", "size"),
// 					// resource.TestCheckResourceAttrSet("data.pinecone_collection.test", "vector_count"),
// 				),
// 			},
// 		},
// 	})
// }

func testAccCollectionDataSourceConfig(name string) string {
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
