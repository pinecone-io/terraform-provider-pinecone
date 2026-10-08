// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func testAccDocumentIndexConfig(name, tags string) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name = %q
  deployment = {
    managed = { cloud = "aws", region = "us-west-2" }
  }
  schema = {
    fields = {
      embedding = { dense_vector = { dimension = 1536, metric = "dotproduct" } }
      terms     = { sparse_vector = {} }
      body      = { string = { full_text_search = { stemming = true } } }
      title     = { string = { full_text_search = { ngram = { min_gram = 2, max_gram = 4, prefix_only = true } } } }
    }
  }
  tags = { %s }
}
`, name, tags)
}

func testAccSpecIndexConfig(name string, dimension int) string {
	return fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name      = %q
  dimension = %d
  spec = {
    serverless = { cloud = "aws", region = "us-west-2" }
  }
}
`, name, dimension)
}

func TestAccIndexResource_documentIndex(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentIndexConfig(rName, `team = "search"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.%", "4"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.embedding.dense_vector.dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.body.string.full_text_search.stemming", "true"),
					resource.TestCheckResourceAttrSet("pinecone_index.test", "schema.fields.body.string.full_text_search.language"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.title.string.full_text_search.ngram.max_gram", "4"),
					resource.TestCheckResourceAttr("pinecone_index.test", "deployment.managed.region", "us-west-2"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "dimension"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "metric"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "spec.serverless.cloud"),
				),
			},
			// The API's defaults and extra fields don't show up as changes.
			{
				Config:   testAccDocumentIndexConfig(rName, `team = "search"`),
				PlanOnly: true,
			},
			{
				Config: testAccDocumentIndexConfig(rName, `team = "search", env = "test"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("pinecone_index.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("pinecone_index.test", "tags.env", "test"),
			},
			{
				ResourceName:      "pinecone_index.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccSpecIndexConfig(rName, 1536),
				ExpectError: regexp.MustCompile("An index can't change how it's described"),
			},
		},
	})
}

// TestAccIndexResource_documentIndexExplicitValues sets the values the API defaults, such as false
// full-text-search options, explicitly. If the API leaves a false value or a description out of its
// response, state wouldn't match the configuration and the schema would plan a replacement.
func TestAccIndexResource_documentIndexExplicitValues(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")
	config := fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name = %q
  deployment = {
    managed = { cloud = "aws", region = "us-west-2" }
  }
  schema = {
    fields = {
      embedding = { dense_vector = { dimension = 8, metric = "cosine", description = "Article embedding" } }
      terms     = { sparse_vector = { description = "Article keywords" } }
      body = {
        string = {
          description      = "Article body"
          full_text_search = { language = "en", stemming = false, stop_words = false }
        }
      }
      title = { string = { full_text_search = { ngram = { min_gram = 2, max_gram = 3, prefix_only = false } } } }
    }
  }
}
`, rName)
	fts := "schema.fields.body.string.full_text_search."

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", fts+"language", "en"),
					resource.TestCheckResourceAttr("pinecone_index.test", fts+"stemming", "false"),
					resource.TestCheckResourceAttr("pinecone_index.test", fts+"stop_words", "false"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.title.string.full_text_search.ngram.prefix_only", "false"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.body.string.description", "Article body"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.embedding.dense_vector.description", "Article embedding"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.terms.sparse_vector.description", "Article keywords"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				ResourceName:      "pinecone_index.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIndexResource_reservedVectorIndex(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")
	config := fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name = %q
  deployment = {
    managed = { cloud = "aws", region = "us-west-2" }
  }
  schema = {
    fields = {
      _values = { dense_vector = { dimension = 1024, metric = "dotproduct" } }
    }
  }
}
`, rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields.%", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields._values.dense_vector.metric", "dotproduct"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "dimension"),
				),
			},
			// The API also reports _sparse_values, which wasn't declared.
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				Config:      testAccSpecIndexConfig(rName, 1024),
				ExpectError: regexp.MustCompile("An index can't change how it's described"),
			},
		},
	})
}

// TestAccIndexResource_schemaModeValidation covers schema configurations that fail at plan time,
// so the provider needs no real credentials.
func TestAccIndexResource_schemaModeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "schema with dimension",
			config: `
resource "pinecone_index" "test" {
  name       = "test"
  dimension  = 1024
  deployment = { managed = { cloud = "aws", region = "us-west-2" } }
  schema     = { fields = { body = { string = { full_text_search = {} } } } }
}`,
			expectError: regexp.MustCompile("Conflicting index configuration"),
		},
		{
			name: "reserved fields mixed with named fields",
			config: `
resource "pinecone_index" "test" {
  name       = "test"
  deployment = { managed = { cloud = "aws", region = "us-west-2" } }
  schema = {
    fields = {
      _values = { dense_vector = { dimension = 1024, metric = "cosine" } }
      body    = { string = { full_text_search = {} } }
    }
  }
}`,
			expectError: regexp.MustCompile("Reserved fields mixed with named fields"),
		},
		{
			name: "document index on BYOC",
			config: `
resource "pinecone_index" "test" {
  name       = "test"
  deployment = { byoc = { environment = "aws-us-east-1-b921" } }
  schema     = { fields = { body = { string = { full_text_search = {} } } } }
}`,
			expectError: regexp.MustCompile("Document indexes require a managed deployment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccDummyApiKeyProviderConfig + tt.config,
						ExpectError: tt.expectError,
					},
				},
			})
		})
	}
}
