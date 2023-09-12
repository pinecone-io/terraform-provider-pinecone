package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/tftest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPineconeCollectionResource(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: tftest.ProviderFactories(map[string]tfsdk.ProviderFactory{
			"pinecone": func() (tfsdk.Provider, error) {
				return New("test"), nil
			},
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccPineconeCollectionResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_collection.test", "name", "test-collection"),
				),
			},
			{
				ResourceName:      "pinecone_collection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPineconeCollectionResourceConfig() string {
	return `
resource "pinecone_collection" "test" {
  name = "test-collection"
}
`
}
