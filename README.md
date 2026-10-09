# Terraform Provider for Pinecone

[![Go
Reference](https://pkg.go.dev/badge/github.com/pinecone-io/terraform-provider-pinecone.svg)](https://pkg.go.dev/github.com/pinecone-io/terraform-provider-pinecone)
[![Go Report
Card](https://goreportcard.com/badge/github.com/pinecone-io/terraform-provider-pinecone)](https://goreportcard.com/report/github.com/pinecone-io/terraform-provider-pinecone)
![Github Actions 
Workflow](https://github.com/pinecone-io/terraform-provider-pinecone/actions/workflows/test.yml/badge.svg)
![GitHub release (latest by
date)](https://img.shields.io/github/v/release/pinecone-io/terraform-provider-pinecone)

The Terraform Provider for Pinecone allows Terraform to manage Pinecone resources including indexes, collections, API keys, and projects.

Note: We take Terraform's security and our users' trust very seriously. If you
believe you have found a security issue in the Terraform Provider for Pinecone,
please responsibly disclose it by contacting us.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v1.4.6
- [Go](https://golang.org/doc/install) >= 1.24. This is necessary to build the
  provider plugin.

## Installing the provider

The provider is registered in the official [Terraform
registry](https://registry.terraform.io/providers/pinecone-io/pinecone/latest).
This enables the provider to be auto-installed when you run `terraform
init`. You can also download the latest binary for your target platform from
the
[releases](https://github.com/pinecone-io/terraform-provider-pinecone/releases)
tab.

## Building the provider

Follow these steps to build the Terraform Provider for Pinecone:

1.  Clone the repository using the following command:

    ```
    sh $ git clone https://github.com/pinecone-io/terraform-provider-pinecone
    ```

1.  Build the provider using the following command. The install directory depends
    on the `GOPATH` environment variable.

        ```
        sh $ go install .
        ```

## Usage

You can enable the provider in your Terraform configuration by adding the
following to your Terraform configuration file:

```terraform
terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}
```

### Authentication

The Terraform Provider for Pinecone supports two types of authentication depending on the operations you need to perform:

#### Regular Operations (Indexes, Collections)

For managing indexes and collections, you need a Pinecone API Key.

##### As a `PINECONE_API_KEY` environment variable

You can configure the Pinecone client using environment variables to avoid
setting sensitive values in the Terraform configuration file. To do so, set
`PINECONE_API_KEY` to your Pinecone API Key. Then the `provider` declaration
is simply:

```terraform
provider "pinecone" {}
```

##### As part of the `provider` declaration

If your API key was set as an [Input Variable](https://developer.hashicorp.com/terraform/language/values/variables),
you can use that value in the declaration. For example:

```terraform
provider "pinecone" {
  api_key = var.pinecone_api_key
}
```

Remember, your API Key should be a protected secret. See how to
[protect sensitive input variables](https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables)
when setting your API Key this way.

#### Admin Operations (API Key and Project Management)

For creating and managing API keys and projects, you need admin credentials (Client ID and Client Secret).

##### Using Environment Variables

Set the following environment variables:

- `PINECONE_CLIENT_ID`: Your Pinecone Client ID
- `PINECONE_CLIENT_SECRET`: Your Pinecone Client Secret

```terraform
provider "pinecone" {}
```

##### Using Provider Configuration

```terraform
provider "pinecone" {
  client_id     = var.pinecone_client_id
  client_secret = var.pinecone_client_secret
}
```

#### Example: Creating a Project

```terraform
# Create a basic project
resource "pinecone_project" "example" {
  name = "my-production-project"
}

# Create a project with CMEK encryption
resource "pinecone_project" "encrypted" {
  name                        = "secure-project"
  force_encryption_with_cmek  = true
}

# Create a project with custom pod limits
resource "pinecone_project" "custom_pods" {
  name     = "high-capacity-project"
  max_pods = 10
}
```

**Note**: Admin credentials are required for API key and project management operations. Regular API keys cannot be used to create or manage other API keys or projects.

### API Key Management

The Terraform Provider for Pinecone supports creating and managing Pinecone API keys. This is useful for automating the creation of API keys for different environments or applications.

#### Available Roles

The following roles can be assigned to API keys:

- `ProjectEditor`: Full access to project resources (default)
- `ProjectViewer`: Read-only access to project resources
- `ControlPlaneEditor`: Full access to control plane operations
- `ControlPlaneViewer`: Read-only access to control plane operations
- `DataPlaneEditor`: Full access to data plane operations
- `DataPlaneViewer`: Read-only access to data plane operations

### Project Management

The Terraform Provider for Pinecone supports creating and managing Pinecone projects. This is useful for organizing your Pinecone resources and managing project-level configurations.

#### Project Features

- **Project Creation**: Create new projects with custom names
- **CMEK Encryption**: Enable customer-managed encryption keys for enhanced security
- **Pod Limits**: Configure maximum number of pods per project

#### Project Configuration Options

- `name`: The name of the project (required)
- `force_encryption_with_cmek`: Enable CMEK encryption (optional, cannot be disabled once enabled)
- `max_pods`: Maximum number of pods allowed in the project (optional, default varies by plan)

**Note**: Project management requires admin credentials (Client ID and Client Secret). Regular API keys cannot be used to manage projects.

### Indexes

An index is described in one of two ways:

- **With `schema` and `deployment`**, the way Pinecone API version `2026-07` describes indexes. Use this for document
  indexes, which combine dense vector, sparse vector, and full-text-search fields.
- **With `dimension`, `metric`, and `spec`**, for vector indexes and indexes with integrated embedding.

A configuration uses one style or the other, and an existing index keeps the style it was created with. Upgrading
from version 4? See the [version 5 upgrade guide](docs/guides/version-5-upgrade.md).

#### Document indexes

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
    }
  }
}
```

Field names are at most 64 bytes and can't start with `$` or `_`. An index can have at most one dense vector field,
one sparse vector field, and 100 full-text-search fields. Metadata fields don't need to be declared: they're indexed
automatically when you upsert data. The schema can't be changed after the index is created; changing it replaces the
index. Document indexes run only on managed (serverless) deployments.

A schema made only of the reserved fields `_values` (dense) and `_sparse_values` (sparse) creates a vector index
instead, used with the vectors API:

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

#### Vector indexes

Serverless vector indexes take a `cloud` and `region`:

```terraform
resource "pinecone_index" "serverless" {
  name      = "my-serverless-index"
  dimension = 1536
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
}
```

BYOC (Bring Your Own Cloud) indexes are deployed into your own cloud environment, identified by the `environment`
Pinecone provides. They need dedicated read capacity:

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

#### Integrated embedding

An index with `embed` embeds text with a model hosted by Pinecone. The model and `field_map` are fixed when the
index is created; `read_parameters` and `write_parameters` can be updated in place.

```terraform
resource "pinecone_index" "integrated" {
  name = "my-integrated-index"
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
  embed = {
    model = "multilingual-e5-large"
    field_map = {
      text = "chunk_text"
    }
  }
}
```

#### Pod-based indexes

Pod-based indexes can't be created with Pinecone API version `2026-07`. Existing pod-based indexes can be imported,
scaled in place with `spec.pod.replicas` and `spec.pod.pod_type`, and deleted.

### Read Capacity

Serverless and BYOC indexes support configurable read capacity. With `schema`, set `read_capacity` at the top level;
with `spec`, set it inside the `serverless` or `byoc` block. Omitting `read_capacity` defaults to `on_demand`, which
BYOC indexes don't support, so BYOC indexes must set `dedicated`:

```terraform
# Dedicated read capacity on a document index
resource "pinecone_index" "dedicated_documents" {
  name = "my-documents"
  deployment = {
    managed = { cloud = "aws", region = "us-east-1" }
  }
  schema = {
    fields = {
      body = { string = { full_text_search = {} } }
    }
  }
  read_capacity = {
    dedicated = {
      node_type = "b1"
      replicas  = 1
      shards    = 1
    }
  }
}

# Dedicated read capacity on a vector index
resource "pinecone_index" "dedicated" {
  name      = "my-index"
  dimension = 1536
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
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

**Note**: To switch from `dedicated` back to `on_demand` after creation, explicitly set the `on_demand = {}` sub-block. Removing the `read_capacity` block entirely will not change the mode already recorded in state. Document indexes and BYOC indexes can't switch from `dedicated` back to `on_demand`.

### Metadata Schema (deprecated)

`spec.serverless.schema` and `spec.byoc.schema` are deprecated: metadata fields are indexed automatically when you
upsert data. New indexes accept them only together with `embed`. Existing indexes keep the value, and it can be
removed from the configuration without replacing the index.

## Documentation

Documentation can be found on the [Terraform
Registry](https://registry.terraform.io/providers/pinecone-io/pinecone/latest).

## Examples

See the
[examples](https://github.com/pinecone-io/terraform-provider-pinecone/tree/main/examples)
for example usage.

## Support

Please create an issue for any support requests.

## Contributing

Thank you to [skyscrapr](https://github.com/skyscrapr/) for developing this
Terraform Provider. The original repository can be found at
[skyscrapr/terraform-provider-pinecone](https://github.com/skyscrapr/terraform-provider-pinecone).
He continues to be the primary developer of this codebase.

We welcome all contributions. If you identify issues or improvements, please
create an issue or pull request.
