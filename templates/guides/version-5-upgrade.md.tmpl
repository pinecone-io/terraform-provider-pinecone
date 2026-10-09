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

A change that would replace an existing pod-based index fails at plan time with the same error, because the
replacement couldn't be created. These changes are to `name`, `dimension`, `metric`, `spec.pod.environment`,
`spec.pod.shards`, or `spec.pod.source_collection`. Tainting a pod-based index, or passing it to `-replace`, fails
the same way.

Creating an index from a collection (`spec.pod.source_collection`) isn't supported either. Pinecone restores data into
new indexes from backups instead, which the provider doesn't manage yet. `pinecone_collection` can still create
collections, but only from an existing pod-based index.

### Integrated embedding is fixed when an index is created

On an existing index, these changes are now rejected at plan time:

- adding `embed` to an index created without it
- removing `embed` from an index created with it
- changing `embed.model` or `embed.field_map`

`embed.read_parameters` and `embed.write_parameters` still update in place.

API version `2026-07` calls indexes created this way legacy integrated indexes, read and written through the legacy
Records API. The API can also declare integrated embedding on the `string` fields of a `schema`, but the provider
doesn't support that yet. Importing a document index whose fields use it leaves the embedding settings out of state.

### The metadata schema is deprecated

`spec.serverless.schema` and `spec.byoc.schema` are deprecated: metadata fields are indexed automatically when you
upsert data, so they no longer need to be declared.

- New indexes accept `spec.serverless.schema` only together with `embed`. Anywhere else it fails at plan time with
  `Metadata schema isn't supported`.
- Existing indexes keep the value in state, and you can remove the attribute from your configuration without
  replacing the index.
- On an index without `embed`, a change that replaces the index fails at plan time with `Metadata schema isn't
  supported` until you remove the attribute.

### Recreating an index

The errors above don't plan a replacement, because replacing an index deletes all of its data. To recreate an index
with the new configuration on purpose, taint it first:

```shell
terraform taint pinecone_index.example
terraform apply
```

Terraform plans the replacement of a tainted index as a new index, so the errors above still apply. A configuration
that can't be created fails at plan time, before the existing index is deleted.

### BYOC indexes need dedicated read capacity

API version `2026-07` doesn't support on-demand read capacity on BYOC indexes, and leaving out `read_capacity`
selects on-demand. A new BYOC index, with `spec.byoc` or `deployment.byoc`, needs `read_capacity.dedicated`, or it
fails at plan time with `BYOC indexes need dedicated read capacity`. Switching an existing BYOC index from dedicated to
`on_demand` fails at plan time too.

An existing BYOC index without `read_capacity` in its configuration keeps working. A change that replaces it, such as to
`name`, `dimension`, `metric`, or `spec.byoc.environment`, or tainting it, fails at plan time until you add
`read_capacity.dedicated`.

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

### Read-only attributes can't be set

`dimension`, `size`, and `vector_count` on `pinecone_collection` are read-only: they're reported by Pinecone, and
setting them never changed the collection. Configurations that set them now fail validation. Remove them from your
configuration.

The same applies to `embed`, `spec`, and `status` on the `pinecone_index` data source.

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
- `terraform import` reads document indexes with `schema` and vector indexes with `spec`, including vector indexes
  created with a `schema` of reserved fields. To import a vector index, describe it with `dimension`, `metric`, and
  `spec`. Otherwise every plan fails with `An index can't change how it's described`.
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

Some existing data source attributes changed:

- `spec.serverless.schema` and `spec.byoc.schema` list only the fields declared with the deprecated metadata
  schema, and are null for indexes created without one. Read the top-level `schema` for the full field list.
- `spec.pod.metadata_config.indexed` is always null, because API version `2026-07` no longer reports it.
- `status.state` can now be `Failed` or `Disabled`. `ScalingDownPodSize` and `Upgrading` are no longer reported.

## Other changes

- Creating an index, or scaling a pod-based index, now fails as soon as the index reaches `InitializationFailed`,
  `Failed`, or `Disabled`. Earlier versions kept waiting on a failed index until the timeout, and treated a disabled
  index as ready. Because setting replicas to 0 disables an index, creating one with
  `read_capacity.dedicated.replicas = 0` fails.
- Creating an index with `read_capacity`, or changing `read_capacity`, now waits until the read capacity is `Ready`
  with the configured replicas and shards running. If it reports `Error`, the apply fails with the API's error message.
  Earlier versions returned as soon as the change was accepted, so a failed scale went unreported. Set
  `timeouts.create` or `timeouts.update` to change how long to wait (10 minutes by default). If a change to an existing
  index times out, it carries on, and the next apply waits for it again. If creating an index times out, Terraform
  marks the index tainted and replaces it on the next apply, so allow enough time in `timeouts.create` for dedicated
  read capacity to be provisioned.
- `embed` requires `model`, and `spec.pod.replicas` must be at least 1.

## Fixes

- Sparse indexes (`vector_type = "sparse"`) can be created, and `metric` can be left out: it defaults to
  `dotproduct`. Earlier versions always failed with `Dimension should not be specified when VectorType is 'sparse'`.
- Integrated indexes (`embed`) without `metric` use the model's metric. Earlier versions always sent `cosine`.
- `embed` without `field_map` fails at plan time. Earlier versions crashed the provider during apply.
- An index without `spec` or `schema`, with an empty `spec`, or with more than one type in `spec` fails at plan
  time. Earlier versions failed during apply.
- Values left out of `read_capacity.dedicated` keep their current setting. Earlier versions could send `0` replicas
  or shards, or an empty node type, when the index was updated for an unrelated change.
- Changing `vector_type` replaces the index, like `dimension` and `metric`. Earlier versions left the index unchanged
  and failed with `Provider produced inconsistent result after apply`.
- Tag values can't be empty: `tags = { team = "" }` fails at plan time. Earlier versions documented `""` as the way
  to remove a tag, but the apply then failed with `Provider produced inconsistent result after apply`. To remove a tag,
  remove its key from `tags`. To remove every tag, set `tags = {}`.
- `pinecone_indexes` reports `deletion_protection`. Earlier versions always reported it as null.
- Updating `embed.read_parameters` or `embed.write_parameters` no longer fails with
  `Provider produced inconsistent result after apply`.
