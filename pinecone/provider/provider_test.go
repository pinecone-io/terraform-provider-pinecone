// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/pinecone-io/go-pinecone/v6/pinecone"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"pinecone": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccDummyAdminProviderConfig configures the provider with placeholder admin
// credentials. Config validation runs before any API call, so tests that only
// assert on validation errors can use this and run without real credentials.
const testAccDummyAdminProviderConfig = `
provider "pinecone" {
  client_id     = "dummy-client-id"
  client_secret = "dummy-client-secret"
}
`

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func testAccPreCheckAdmin(t *testing.T) {
	// Check for admin client credentials
	if v := os.Getenv("PINECONE_CLIENT_ID"); v == "" {
		t.Fatal("PINECONE_CLIENT_ID environment variable must be set for admin acceptance tests")
	}
	if v := os.Getenv("PINECONE_CLIENT_SECRET"); v == "" {
		t.Fatal("PINECONE_CLIENT_SECRET environment variable must be set for admin acceptance tests")
	}
}

// testAccAdminClient builds an admin client from the environment. Destroy
// verification needs its own client because the provider's is torn down with
// the test's provider server before CheckDestroy runs.
func testAccAdminClient(t *testing.T) (*pinecone.AdminClient, error) {
	t.Helper()
	sourceTag := "terraform"
	return pinecone.NewAdminClient(pinecone.NewAdminClientParams{
		ClientId:     os.Getenv("PINECONE_CLIENT_ID"),
		ClientSecret: os.Getenv("PINECONE_CLIENT_SECRET"),
		SourceTag:    &sourceTag,
	})
}

// testAccCheckAdminResourceDestroyed returns a CheckDestroy that asserts every
// resource of resourceType named in state is gone from the API. describe is the
// SDK call for that type; a 404 means the destroy landed, a success means it did
// not, and any other error is reported rather than swallowed so a transient
// failure is never mistaken for successful cleanup.
func testAccCheckAdminResourceDestroyed(t *testing.T, resourceType string, describe func(*pinecone.AdminClient, context.Context, string) error) func(*terraform.State) error {
	return func(s *terraform.State) error {
		t.Helper()

		client, err := testAccAdminClient(t)
		if err != nil {
			return fmt.Errorf("failed to build an admin client to verify %s cleanup: %w", resourceType, err)
		}

		for name, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			id := rs.Primary.Attributes["id"]
			if id == "" {
				return fmt.Errorf("%s has no id in state; cannot verify it was destroyed", name)
			}

			err := describe(client, context.Background(), id)
			switch {
			case err == nil:
				return fmt.Errorf("%s (%s) still exists after destroy", name, id)
			case isNotFoundErr(err):
				continue
			default:
				return fmt.Errorf("could not verify %s (%s) was destroyed: %w", name, id, err)
			}
		}

		return nil
	}
}

func testAccCheckRoleBindingDestroy(t *testing.T) func(*terraform.State) error {
	return testAccCheckAdminResourceDestroyed(t, "pinecone_role_binding",
		func(c *pinecone.AdminClient, ctx context.Context, id string) error {
			_, err := c.RoleBinding.Describe(ctx, id)
			return err
		})
}

func testAccCheckInviteDestroy(t *testing.T) func(*terraform.State) error {
	return testAccCheckAdminResourceDestroyed(t, "pinecone_invite",
		func(c *pinecone.AdminClient, ctx context.Context, id string) error {
			_, err := c.Invite.Describe(ctx, id)
			return err
		})
}
