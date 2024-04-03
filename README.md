# Terraform Pinecone Provider

[![Go Reference](https://pkg.go.dev/badge/github.com/pinecone-io/terraform-provider-pinecone.svg)](https://pkg.go.dev/github.com/pinecone-io/terraform-provider-pinecone)
[![Go Report Card](https://goreportcard.com/badge/github.com/pinecone-io/terraform-provider-pinecone)](https://goreportcard.com/report/github.com/pinecone-io/terraform-provider-pinecone)
![Github Actions Workflow](https://github.com/pinecone-io/terraform-provider-pinecone/actions/workflows/test.yml/badge.svg)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/pinecone-io/terraform-provider-pinecone)

The Terraform Pinecone Provider allows Terraform to manage Pinecone resources.

Note: We take Terraform's security and our users' trust very seriously. If you believe you have found a security issue in the Terraform Pinecone Provider, please responsibly disclose it by contacting us.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v1.4.6
- [Go](https://golang.org/doc/install) >= 1.20. This is necessary tto build the provider plugin.

## Installing the provider

The provider is registered in the official [Terraform registry](https://registry.terraform.io/providers/skyscrapr/pinecone/latest). This enables the provider to be auto-installed when you run ```terraform init```. You can also download the latest binary for your target platform from the [releases](https://github.com/pinecone-io/terraform-provider-pinecone/releases) tab.

## Building the provider

Follow these steps to build the Terraform Pinecone provider. You can also download the latest binary for your target platform from the [releases](https://github.com/pinecone-io/terraform-provider-pinecone/releases) tab.

1. Clone the repository using the following command:

    ```sh
    $ git clone https://github.com/pinecone-io/terraform-provider-pinecone
    ```

1. Build the provider using the following command. The install directory depends on the GOPATH environment variable.

    ```sh
    $ go install .
    ```

## Usage

You can enable the provider in your Terraform configuration by add the folowing to the configuration file:

```terraform
terraform {
  required_providers {
    openai = {
      source = "pinecone-io/pinecone"
    }
  }
}
```

You can configure the Pinecone client using environment variables to avoid setting sensitive values in the Terraform configuration file. To do so, follow these steps:

+ Set `PINECONE_API_KEY` to your Pinecone API Key.
+ Set `PINECONE_ENVIRONMENT` to your Pinecone environment. 

## Documentation

Documentation can be found on the [Terraform Registry](https://registry.terraform.io/providers/pinecone-io/pinecone/latest). 

## Examples

<<<<<<< HEAD
See the [examples](https://github.com/pinecone-io/pinecone-terraform-provider/examples) for example usage.

## Support

Raise an issue for any support-related requirements.

## Contributing

Thank you to [skyscrapr](https://github.com/skyscrapr/) for developing this Terraform Provider. The original repository can be
found at [skyscrapr/terraform-provider-pinecone](https://github.com/skyscrapr/terraform-provider-pinecone). He continues
to be the primary developer of this codebase.

We welcome all contributions. If you identify issues or improvements, please file an issue or raise a pull request.
