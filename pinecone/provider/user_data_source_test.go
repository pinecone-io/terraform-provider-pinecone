package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUserDataSource_exactlyOneOf exercises the plan-time ExactlyOneOf
// validator wiring id and email. It needs no credentials: validation rejects the
// config before the provider makes any API call.
func TestAccUserDataSource_exactlyOneOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "neither id nor email",
			config: `
data "pinecone_user" "test" {}`,
			expectError: regexp.MustCompile("Missing Attribute Configuration"),
		},
		{
			name: "both id and email",
			config: `
data "pinecone_user" "test" {
  id    = "00000000-0000-0000-0000-000000000000"
  email = "teammate@example.com"
}`,
			expectError: regexp.MustCompile("Invalid Attribute Combination"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccDummyAdminProviderConfig + tt.config,
						ExpectError: tt.expectError,
					},
				},
			})
		})
	}
}
