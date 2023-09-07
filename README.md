# Terraform Pinecone Provider

The Terraform Pinecone Provider allows Terraform to manage Pinecone resources.

- Contributing guide
- Quarterly development roadmap
- FAQ
- Tutorials

Please note: We take Terraform's security and our users' trust very seriously. If you believe you have found a security issue in the Terraform Pinecone Provider, please responsibly disclose it by contacting us.

## Development

To enable local development you can use the follow `~/.terraformrc` file:
```
provider_installation {

  dev_overrides {
    "registry.terraform.io/skyscrapr/pinecone" = "/Users/[USERNAME]/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```