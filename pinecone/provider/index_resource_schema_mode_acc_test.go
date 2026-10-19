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
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
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

func testAccSchemaVectorIndexConfig(name string, dimension int) string {
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
      _values = { dense_vector = { dimension = %d, metric = "cosine" } }
    }
  }
}
`, name, dimension)
}

// TestAccIndexResource_specToSchema moves an index created with spec to schema and deployment. The
// move only changes state, and a move back to spec is rejected.
func TestAccIndexResource_specToSchema(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccSpecIndexConfig(rName, 1024),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1024"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "cosine"),
				),
			},
			{
				Config:      testAccSchemaVectorIndexConfig(rName, 768),
				ExpectError: regexp.MustCompile("The configuration describes a different index"),
			},
			{
				Config: testAccSchemaVectorIndexConfig(rName, 1024),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("pinecone_index.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields._values.dense_vector.dimension", "1024"),
					resource.TestCheckResourceAttr("pinecone_index.test", "deployment.managed.region", "us-west-2"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "dimension"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "spec.serverless.region"),
				),
			},
			{
				Config:   testAccSchemaVectorIndexConfig(rName, 1024),
				PlanOnly: true,
			},
			{
				Config:      testAccSpecIndexConfig(rName, 1024),
				ExpectError: regexp.MustCompile("An index can't move back to spec"),
			},
		},
	})
}

// TestAccIndexResource_importSpecIndexIntoSchema imports an index created with spec into a schema
// configuration. A removed block drops it from state without deleting it, then an import block
// brings it back under a new address, as a user would after losing state.
func TestAccIndexResource_importSpecIndexIntoSchema(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")
	original := fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "original" {
  name      = %q
  dimension = 1024
  spec = {
    serverless = { cloud = "aws", region = "us-west-2" }
  }
}
`, rName)
	imported := testAccSchemaVectorIndexConfig(rName, 1024) + fmt.Sprintf(`
removed {
  from = pinecone_index.original
  lifecycle {
    destroy = false
  }
}

import {
  to = pinecone_index.test
  id = %q
}
`, rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		// removed blocks need Terraform 1.7, and import blocks 1.5.
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_7_0)},
		Steps: []resource.TestStep{
			{
				Config: original,
				Check:  resource.TestCheckResourceAttr("pinecone_index.original", "dimension", "1024"),
			},
			{
				Config: imported,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("pinecone_index.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", "schema.fields._values.dense_vector.dimension", "1024"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "spec.serverless.region"),
				),
			},
			{
				Config:   imported,
				PlanOnly: true,
			},
		},
	})
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
				ExpectError: regexp.MustCompile("An index can't move back to spec"),
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

// TestAccIndexResource_documentIndexIntegratedEmbedding creates a document index whose string fields
// use dense and sparse integrated embedding. The API defaults the model, dimension, metric, and
// full-text search, and merges the configured parameters with the model's defaults, none of which
// show up as changes. After an import, which can't record what was configured, the configured
// parameters are taken into state without replacing the index.
func TestAccIndexResource_documentIndexIntegratedEmbedding(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tftest")
	config := fmt.Sprintf(`
provider "pinecone" {
}

resource "pinecone_index" "test" {
  name = %q
  deployment = {
    managed = { cloud = "aws", region = "us-east-1" }
  }
  schema = {
    fields = {
      body = {
        string = {
          embed = { read_parameters = { truncate = "END" } }
        }
      }
      summary = {
        string = {
          sparse_embed = {}
          store_text   = false
        }
      }
    }
  }
}
`, rName)
	body, summary := "schema.fields.body.string.", "schema.fields.summary.string."

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(),
					resource.TestCheckResourceAttr("pinecone_index.test", body+"embed.model", "llama-text-embed-v2"),
					resource.TestCheckResourceAttrSet("pinecone_index.test", body+"embed.dimension"),
					resource.TestCheckResourceAttrSet("pinecone_index.test", body+"embed.metric"),
					resource.TestCheckResourceAttr("pinecone_index.test", body+"embed.read_parameters.%", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", body+"embed.effective_read_parameters.truncate", "END"),
					resource.TestCheckResourceAttr("pinecone_index.test", body+"store_text", "true"),
					resource.TestCheckResourceAttrSet("pinecone_index.test", body+"full_text_search.language"),
					resource.TestCheckResourceAttr("pinecone_index.test", summary+"sparse_embed.model", "pinecone-sparse-english-v0"),
					resource.TestCheckResourceAttr("pinecone_index.test", summary+"store_text", "false"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", summary+"full_text_search.language"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				ResourceName:            "pinecone_index.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{body + "embed.read_parameters.%", body + "embed.read_parameters.truncate"},
			},
			{
				ResourceName:       "pinecone_index.test",
				ImportState:        true,
				ImportStatePersist: true,
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("pinecone_index.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("pinecone_index.test", body+"embed.read_parameters.truncate", "END"),
			},
			{
				Config:   config,
				PlanOnly: true,
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
				ExpectError: regexp.MustCompile("An index can't move back to spec"),
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
		{
			name: "store_text false with full-text search",
			config: `
resource "pinecone_index" "test" {
  name       = "test"
  deployment = { managed = { cloud = "aws", region = "us-west-2" } }
  schema     = { fields = { body = { string = { embed = {}, full_text_search = {}, store_text = false } } } }
}`,
			expectError: regexp.MustCompile("store_text = false can't be combined with full_text_search"),
		},
		{
			name: "dense vector field and embed",
			config: `
resource "pinecone_index" "test" {
  name       = "test"
  deployment = { managed = { cloud = "aws", region = "us-west-2" } }
  schema = {
    fields = {
      embedding = { dense_vector = { dimension = 1024, metric = "cosine" } }
      body      = { string = { embed = {} } }
    }
  }
}`,
			expectError: regexp.MustCompile("Too many dense vector fields"),
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
