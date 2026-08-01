package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TestRoleBindingsDataSourceSchema_PairedFilterValidators asserts that the
// paired filter parameters carry validators, wiring principal_id -> principal_type
// and resource_id -> resource_type per the SDK's list contract. These attributes
// had no validators before, so a non-empty validator set confirms the wiring is
// present and guards against its accidental removal.
func TestRoleBindingsDataSourceSchema_PairedFilterValidators(t *testing.T) {
	d := &RoleBindingsDataSource{}
	resp := &datasource.SchemaResponse{}
	d.Schema(t.Context(), datasource.SchemaRequest{}, resp)

	for _, name := range []string{"principal_id", "resource_id"} {
		attr, ok := resp.Schema.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("attribute %q is not a schema.StringAttribute", name)
		}
		if len(attr.Validators) == 0 {
			t.Errorf("attribute %q has no validators; expected an AlsoRequires validator", name)
		}
	}
}
