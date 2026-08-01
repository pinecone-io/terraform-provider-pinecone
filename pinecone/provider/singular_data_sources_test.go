package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TestSingularAdminDataSourceSchemas asserts that every admin entity with a
// Describe operation in the SDK is reachable by ID through a singular data
// source, and that each one takes a required id. This is the parity contract
// raised in review: pinecone_invite exists because the invites list cannot
// return accepted (processed) invites, and pinecone_role_binding exists because
// the role bindings list cannot filter on a binding's own ID.
func TestSingularAdminDataSourceSchemas(t *testing.T) {
	dataSources := map[string]datasource.DataSource{
		"pinecone_role_binding":    &RoleBindingDataSource{},
		"pinecone_service_account": &ServiceAccountDataSource{},
		"pinecone_invite":          &InviteDataSource{},
	}

	for name, d := range dataSources {
		t.Run(name, func(t *testing.T) {
			resp := &datasource.SchemaResponse{}
			d.Schema(t.Context(), datasource.SchemaRequest{}, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("schema produced errors: %v", resp.Diagnostics.Errors())
			}

			attr, ok := resp.Schema.Attributes["id"].(schema.StringAttribute)
			if !ok {
				t.Fatalf("id is not a schema.StringAttribute")
			}
			if !attr.Required {
				t.Error("id must be required for a by-id lookup")
			}
		})
	}

	// pinecone_user is the exception: it looks up by id or email, so neither is
	// required on its own. ExactlyOneOf enforces the choice during plan.
	t.Run("pinecone_user", func(t *testing.T) {
		d := &UserDataSource{}
		resp := &datasource.SchemaResponse{}
		d.Schema(t.Context(), datasource.SchemaRequest{}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("schema produced errors: %v", resp.Diagnostics.Errors())
		}
		if len(d.ConfigValidators(t.Context())) == 0 {
			t.Error("expected an ExactlyOneOf config validator wiring id and email")
		}
	})
}
