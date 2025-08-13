# Terraform Provider for Pinecone

[![Go
Reference](https://pkg.go.dev/badge/github.com/pinecone-io/terraform-provider-pinecone.svg)](https://pkg.go.dev/github.com/pinecone-io/terraform-provider-pinecone)
[![Go Report
Card](https://goreportcard.com/badge/github.com/pinecone-io/terraform-provider-pinecone)](https://goreportcard.com/report/github.com/pinecone-io/terraform-provider-pinecone)
![Github Actions 
Workflow](https://github.com/pinecone-io/terraform-provider-pinecone/actions/workflows/test.yml/badge.svg)
![GitHub release (latest by
date)](https://img.shields.io/github/v/release/pinecone-io/terraform-provider-pinecone)

The Terraform Provider for Pinecone allows Terraform to manage Pinecone resources.

Note: We take Terraform's security and our users' trust very seriously. If you
believe you have found a security issue in the Terraform Provider for Pinecone,
please responsibly disclose it by contacting us.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v1.4.6
- [Go](https://golang.org/doc/install) >= 1.20. This is necessary to build the
  provider plugin.

## Installing the provider

The provider is registered in the official [Terraform 
registry](https://registry.terraform.io/providers/pinecone-io/pinecone/latest).
This enables the provider to be auto-installed when you run ```terraform
init```. You can also download the latest binary for your target platform from
the
[releases](https://github.com/pinecone-io/terraform-provider-pinecone/releases)
tab.

## Building the provider

Follow these steps to build the Terraform Provider for Pinecone: 

1. Clone the repository using the following command:

    ```
    sh $ git clone https://github.com/pinecone-io/terraform-provider-pinecone
    ```

1. Build the provider using the following command. The install directory depends
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

#### Admin Operations (API Key Management)

For creating and managing API keys, you need admin credentials (Client ID and Client Secret).

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

**Note**: Admin credentials are required for API key management operations. Regular API keys cannot be used to create or manage other API keys.

### API Key Management

The Terraform Provider for Pinecone supports creating and managing Pinecone API keys. This is useful for automating the creation of API keys for different environments or applications.

#### Creating API Keys

```terraform
# Create an API key with default roles (ProjectEditor)
resource "pinecone_api_key" "example" {
  name       = "example-api-key"
  project_id = "your-project-id"
}

# Create an API key with custom roles
resource "pinecone_api_key" "readonly" {
  name       = "readonly-api-key"
  project_id = "your-project-id"
  roles      = ["ProjectViewer", "DataPlaneViewer"]
}
```

#### Available Roles

The following roles can be assigned to API keys:

- `ProjectEditor`: Full access to project resources (default)
- `ProjectViewer`: Read-only access to project resources
- `ControlPlaneEditor`: Full access to control plane operations
- `ControlPlaneViewer`: Read-only access to control plane operations
- `DataPlaneEditor`: Full access to data plane operations
- `DataPlaneViewer`: Read-only access to data plane operations

#### Managing API Keys

You can update API key names and roles:

```terraform
# Update API key name and roles
resource "pinecone_api_key" "updatable" {
  name       = "updated-name"
  project_id = "your-project-id"
  roles      = ["ProjectViewer", "DataPlaneViewer"]
}
```

#### Importing Existing API Keys

You can import existing API keys using the format `project_id:api_key_id`:

```bash
terraform import pinecone_api_key.example your-project-id:your-api-key-id
```

#### Example: Complete API Key Management Workflow

```terraform
# Provider configuration for admin operations
provider "pinecone" {
  client_id     = var.pinecone_client_id
  client_secret = var.pinecone_client_secret
}

# Create API keys for different environments
resource "pinecone_api_key" "dev" {
  name       = "dev-environment-key"
  project_id = var.pinecone_project_id
  roles      = ["ProjectEditor", "DataPlaneEditor"]
}

resource "pinecone_api_key" "prod" {
  name       = "prod-environment-key"
  project_id = var.pinecone_project_id
  roles      = ["ProjectViewer", "DataPlaneViewer"]
}

# Output the generated API keys (sensitive)
output "dev_api_key" {
  description = "Development environment API key"
  value       = pinecone_api_key.dev.key
  sensitive   = true
}

output "prod_api_key" {
  description = "Production environment API key"
  value       = pinecone_api_key.prod.key
  sensitive   = true
}
```

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
