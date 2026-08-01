package provider

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleBindingResource(t *testing.T) {
	t.Parallel()
	projectId := os.Getenv("PINECONE_PROJECT_ID")
	clientId := os.Getenv("PINECONE_CLIENT_ID")
	clientSecret := os.Getenv("PINECONE_CLIENT_SECRET")

	if projectId == "" || clientId == "" || clientSecret == "" {
		t.Skip("PINECONE_PROJECT_ID, PINECONE_CLIENT_ID, and PINECONE_CLIENT_SECRET environment variables are required for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleBindingResourceConfig(projectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "principal_type", "api_key"),
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "resource_type", "project"),
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "resource_id", projectId),
					resource.TestCheckResourceAttr("pinecone_role_binding.test", "role", "ProjectViewer"),
					resource.TestCheckResourceAttrSet("pinecone_role_binding.test", "id"),
					resource.TestCheckResourceAttrSet("pinecone_role_binding.test", "created_at"),
					resource.TestCheckResourceAttrPair("pinecone_role_binding.test", "principal_id", "pinecone_api_key.test", "id"),
				),
			},
			{
				ResourceName:      "pinecone_role_binding.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccRoleBindingResource_rejectsInvitePrincipal asserts that the resource
// refuses principal_type = "invite" at plan time, and that the error explains the
// redirect rather than just omitting the value from a list. Needs no credentials.
//
// The exclusion exists because the server re-points an invite's bindings to the
// user principal on acceptance, keeping the same binding id — so a managed invite
// binding would refresh into a forced replacement that deletes the user's role.
func TestAccRoleBindingResource_rejectsInvitePrincipal(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDummyAdminProviderConfig + `
resource "pinecone_role_binding" "test" {
  principal_id   = "00000000-0000-0000-0000-000000000000"
  principal_type = "invite"
  resource_type  = "organization"
  role           = "OrgMember"
}`,
				// The message must name the supported alternative, not just reject.
				ExpectError: regexp.MustCompile(`pinecone_invite\.role_bindings`),
			},
		},
	})
}

// TestRoleBindingPrincipalTypeValidator_allowsTrackableTypes guards the other
// half of the contract: narrowing the resource must not narrow the data source
// filter, which is how an invite's bindings are read.
func TestRoleBindingPrincipalTypeValidator_allowsTrackableTypes(t *testing.T) {
	for _, want := range []string{"user", "service_account", "api_key"} {
		if !slices.Contains(roleBindingPrincipalTypes, want) {
			t.Errorf("principal_type %q must remain valid on pinecone_role_binding", want)
		}
	}
	if slices.Contains(roleBindingPrincipalTypes, "invite") {
		t.Error("principal_type \"invite\" must not be valid on pinecone_role_binding")
	}

	d := &RoleBindingsDataSource{}
	resp := &datasource.SchemaResponse{}
	d.Schema(t.Context(), datasource.SchemaRequest{}, resp)
	attr, ok := resp.Schema.Attributes["principal_type"].(dsschema.StringAttribute)
	if !ok {
		t.Fatal("principal_type is not a schema.StringAttribute on the role bindings data source")
	}
	if len(attr.Validators) == 0 {
		t.Fatal("expected the data source principal_type filter to keep its validator")
	}
	// The filter must still accept "invite"; reading invite bindings is the
	// documented replacement for managing them.
	req := validator.StringRequest{
		Path:        path.Root("principal_type"),
		ConfigValue: types.StringValue("invite"),
	}
	vResp := &validator.StringResponse{}
	attr.Validators[0].ValidateString(t.Context(), req, vResp)
	if vResp.Diagnostics.HasError() {
		t.Errorf("data source filter rejected principal_type \"invite\": %v", vResp.Diagnostics.Errors())
	}
}

// TestAccRoleBindingResource_scopeValidation exercises the plan-time
// ConfigValidator. It needs no credentials: validation rejects the config before
// the provider makes any API call, so the dummy credentials below are never used.
func TestAccRoleBindingResource_scopeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "organization scope rejects resource_id",
			config: `
resource "pinecone_role_binding" "test" {
  principal_id   = "00000000-0000-0000-0000-000000000000"
  principal_type = "user"
  resource_type  = "organization"
  resource_id    = "00000000-0000-0000-0000-000000000001"
  role           = "OrgMember"
}`,
			expectError: regexp.MustCompile("Unexpected resource_id"),
		},
		{
			name: "project scope requires resource_id",
			config: `
resource "pinecone_role_binding" "test" {
  principal_id   = "00000000-0000-0000-0000-000000000000"
  principal_type = "user"
  resource_type  = "project"
  role           = "ProjectViewer"
}`,
			expectError: regexp.MustCompile("Missing resource_id"),
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

func testAccRoleBindingResourceConfig(projectId string) string {
	return fmt.Sprintf(`
provider "pinecone" {
  client_id     = "%s"
  client_secret = "%s"
}

resource "pinecone_api_key" "test" {
  name       = "test-role-binding-key"
  project_id = %[3]q
}

resource "pinecone_role_binding" "test" {
  principal_id   = pinecone_api_key.test.id
  principal_type = "api_key"
  resource_type  = "project"
  resource_id    = %[3]q
  role           = "ProjectViewer"
}
`, os.Getenv("PINECONE_CLIENT_ID"), os.Getenv("PINECONE_CLIENT_SECRET"), projectId)
}
