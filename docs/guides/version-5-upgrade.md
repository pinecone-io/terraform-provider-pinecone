---
page_title: "Upgrading to version 5.0.0"
subcategory: "Upgrade guides"
description: |-
  What changes in version 5.0.0 of the Pinecone provider, and how to update configurations written for earlier versions.
---

# Upgrading to version 5.0.0

Version 5.0.0 moves the provider to [go-pinecone v7](https://github.com/pinecone-io/go-pinecone) and Pinecone API
version `2026-07`, and adds support for creating indexes from an explicit schema.

Most configurations written for version 4 plan with no changes. The exceptions below are reported as errors at plan
time, before anything is changed. Pin the new version, run `terraform plan`, and work through any errors:

```terraform
terraform {
  required_providers {
    pinecone = {
      source  = "pinecone-io/pinecone"
      version = "~> 5.0"
    }
  }
}
```

## Breaking changes

### Pod-based indexes can't be created

API version `2026-07` doesn't support creating pod-based indexes. A `pinecone_index` with `spec.pod` that doesn't
exist yet fails at plan time with `Pod-based indexes can't be created`. Create a serverless or BYOC index instead.

Existing pod-based indexes keep working: they can be imported, refreshed, and deleted. `spec.pod.replicas` and
`spec.pod.pod_type` now update in place instead of replacing the index, and the apply waits until scaling finishes
(set `timeouts.update` to change how long). The pod size can only grow within the same pod family, for example from
`p1.x1` to `p1.x2`; other `pod_type` changes are rejected at plan time.

Creating an index from a collection (`spec.pod.source_collection`) isn't supported either. Pinecone restores data into
new indexes from backups instead, which the provider doesn't manage yet.

### Integrated embedding is fixed when an index is created

On an existing index, these changes are now rejected at plan time:

- adding `embed` to an index created without it
- removing `embed` from an index created with it
- changing `embed.model` or `embed.field_map`

`embed.read_parameters` and `embed.write_parameters` still update in place.

### The metadata schema is deprecated

`spec.serverless.schema` and `spec.byoc.schema` are deprecated: metadata fields are indexed automatically when you
upsert data, so they no longer need to be declared.

- New indexes accept `spec.serverless.schema` only together with `embed`. Anywhere else it fails at plan time with
  `Metadata schema isn't supported`.
- Existing indexes keep the value in state, and you can remove the attribute from your configuration without
  replacing the index.

### Recreating an index

The errors above don't plan a replacement, because replacing an index deletes all of its data. To recreate an index
with the new configuration on purpose, taint it first:

```shell
terraform taint pinecone_index.example
terraform apply
```

### BYOC indexes need dedicated read capacity

API version `2026-07` doesn't support on-demand read capacity on BYOC indexes, and leaving out `read_capacity`
selects on-demand. A new `spec.byoc` index needs `read_capacity.dedicated`, or the apply fails:

```terraform
resource "pinecone_index" "byoc" {
  name      = "my-byoc-index"
  dimension = 1536
  spec = {
    byoc = {
      environment = "my-byoc-env-id"
      read_capacity = {
        dedicated = {
          node_type = "b1"
          replicas  = 1
          shards    = 1
        }
      }
    }
  }
}
```

## New: indexes defined by schema and deployment

`pinecone_index` can describe an index with `schema` and `deployment`, the way API version `2026-07` does, instead
of `dimension`, `metric`, `vector_type`, and `spec`. A schema of named fields creates a document index, used with the
documents API, which can combine dense vector, sparse vector, and full-text-search fields:

```terraform
resource "pinecone_index" "articles" {
  name = "articles"
  deployment = {
    managed = { cloud = "aws", region = "us-east-1" }
  }
  schema = {
    fields = {
      embedding = { dense_vector = { dimension = 1536, metric = "dotproduct" } }
      terms     = { sparse_vector = {} }
      body      = { string = { full_text_search = { stemming = true } } }
      title     = { string = { full_text_search = { ngram = { min_gram = 2, max_gram = 4, prefix_only = true } } } }
    }
  }
}
```

A schema made only of the reserved fields `_values` (dense) and `_sparse_values` (sparse) creates a vector index,
used with the vectors API, the same kind of index `dimension` and `metric` create:

```terraform
resource "pinecone_index" "products" {
  name = "products"
  deployment = {
    managed = { cloud = "aws", region = "us-east-1" }
  }
  schema = {
    fields = {
      _values = { dense_vector = { dimension = 1024, metric = "dotproduct" } }
    }
  }
}
```

Things to know:

- A configuration uses one style or the other. `schema` and `deployment` can't be combined with `dimension`,
  `metric`, `vector_type`, `spec`, or `embed`.
- An existing index can't switch between the two styles. Keep describing existing indexes the way they were created.
- With `schema`, read capacity and encryption are set at the top level: `read_capacity` and `cmek_id`.
- Changing `schema`, `deployment`, or `cmek_id` replaces the index.
- Document indexes run only on managed deployments, and can't move from dedicated read capacity back to on-demand.
- `terraform import` reads document indexes with `schema` and vector indexes with `spec`.
- Indexes with integrated embedding (`embed`) are still created with `spec`.

## New data source attributes

`pinecone_index` and `pinecone_indexes` report each index's full `schema`, including the reserved vector fields and
the metadata fields added as data is upserted, as well as `deployment`, top-level `read_capacity`, `private_host`,
`cmek_id`, `source_backup_id`, and `source_collection`.

```terraform
data "pinecone_index" "products" {
  name = "products"
}

output "products_dimension" {
  value = data.pinecone_index.products.schema.fields["_values"].dense_vector.dimension
}
```

## Fixes

- Sparse indexes (`vector_type = "sparse"`) can be created, and `metric` can be left out: it defaults to
  `dotproduct`. Earlier versions always failed with `Dimension should not be specified when VectorType is 'sparse'`.
- Integrated indexes (`embed`) without `metric` use the model's metric. Earlier versions always sent `cosine`.
- `embed` without `field_map` fails at plan time. Earlier versions crashed the provider during apply.
- Values left out of `read_capacity.dedicated` keep their current setting. Earlier versions could send `0` replicas
  or shards, or an empty node type, when the index was updated for an unrelated change.
- `pinecone_indexes` reports `deletion_protection`. Earlier versions always reported it as null.
- Updating `embed.read_parameters` or `embed.write_parameters` no longer fails with
  `Provider produced inconsistent result after apply`.
