# Terraform Pinecone Provider

[![Go Reference](https://pkg.go.dev/badge/github.com/skyscrapr/terraform-provider-pinecone.svg)](https://pkg.go.dev/github.com/skyscrapr/terraform-provider-pinecone)
[![Go Report Card](https://goreportcard.com/badge/github.com/skyscrapr/terraform-provider-pinecone)](https://goreportcard.com/report/github.com/skyscrapr/terraform-provider-pinecone)
[![codecov](https://codecov.io/gh/skyscrapr/terraform-provider-pinecone/graph/badge.svg?token=qobuIzQPuM)](https://codecov.io/gh/skyscrapr/terraform-provider-pinecone)
![Github Actions Workflow](https://github.com/skyscrapr/terraform-provider-pinecone/actions/workflows/test.yml/badge.svg)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/skyscrapr/terraform-provider-pinecone)
![License](https://img.shields.io/dub/l/vibe-d.svg)

The Terraform Pinecone Provider allows Terraform to manage Pinecone resources.

Note: We take Terraform's security and our users' trust very seriously. If you believe you have found a security issue in the Terraform Pinecone Provider, please responsibly disclose it by contacting us.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v1.4.6
- [Go](https://golang.org/doc/install) >= 1.20. This is necessary tto build the provider plugin.

## Installing the provider

The provider is registered in the official [terraform registry](https://registry.terraform.io/providers/skyscrapr/pinecone/latest). This enables the provider to be auto-installed when you run ```terraform init```. You can also download the latest binary for your target platform from the [releases](https://github.com/skyscrapr/terraform-provider-pinecone/releases) tab.

## Building the provider

Follow these steps to build the Terraform Pinecone provider:

1. Clone the repository using the following command:

    ```sh
    $ git clone https://github.com/skyscrapr/terraform-provider-pinecone
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
      source = "skyscrapr/pinecone"
    }
  }
}
```

You can configure the Pinecone client using environment variables to avoid setting sensitive values in the Terraform configuration file. To do so, follow these steps:

+ Set `PINECONE_API_KEY` to your Pinecone API Key.
+ Set `PINECONE_ENVIRONMENT` to your Pinecone environment. 

## Documentation

Documentation can be found on the [Terraform Registry](https://registry.terraform.io/providers/skyscrapr/pinecone/latest). 

## Examples

See the [examples](https://github.com/pinecone-io/pinecone-terraform-provider/examples) for example usage.

## Support

Raise an issue for any support-related requirements.
