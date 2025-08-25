# Testing Local Provider Development

This directory contains test configurations for the Pinecone Terraform provider during development.

## Prerequisites

- Go installed and configured
- Terraform CLI installed
- Pinecone account with admin credentials (client_id and client_secret)

## Building the Local Provider

First, build the provider binary from the root directory:

```bash
cd ../
go build -o terraform-provider-pinecone
```

## Testing with Local Provider

### Step 1: Configure Credentials and Setup

Edit `setup-env.sh` and add your Pinecone admin credentials:

```bash
export PINECONE_CLIENT_ID="your-actual-client-id"
export PINECONE_CLIENT_SECRET="your-actual-client-secret"
```

Then run the setup script (this will clean up previous setup and set environment variables):

```bash
cd test-local
source setup-env.sh
```

### Step 2: Update Project ID

Edit `main.tf` and update the project ID in the resource:

```hcl
resource "pinecone_api_key" "test" {
  name       = "terraform-test-api-key"
  project_id = "your-actual-project-id"  # Replace with your actual project ID
}
```

**Note**: You can modify `main.tf` to test any resource type, including unreleased features. Refer to the `../examples/` folder for examples of how to create different resources:
- `../examples/resources/pinecone_index/` - For index creation
- `../examples/resources/pinecone_collection/` - For collection creation
- `../examples/resources/pinecone_project/` - For project creation (requires admin credentials)
- `../examples/data-sources/` - For data source usage

### Step 3: Initialize Terraform (Optional)

You can run `terraform init` to initialize the workspace, but this will use the registry provider which doesn't support all features:

```bash
terraform init
```

### Step 4: Test with Local Provider

To use your locally built provider with all features, simply run:

```bash
source setup-env.sh && terraform plan
```

This command:
1. Sources your credentials from `setup-env.sh`
2. Automatically configures Terraform to use your local provider
3. Runs `terraform plan` using your local provider

## Configuration

### Dev Overrides Setup

The `.terraformrc` file configures Terraform to use your locally built provider:

```hcl
provider_installation {
  dev_overrides {
    "pinecone-io/pinecone" = "../"
  }
  direct {
    exclude = ["pinecone-io/pinecone"]
  }
}
```

### Provider Configuration

The provider will automatically pick up credentials from environment variables:

- `PINECONE_CLIENT_ID` - Your Pinecone client ID
- `PINECONE_CLIENT_SECRET` - Your Pinecone client secret

The `main.tf` file uses an empty provider block since credentials are loaded from environment variables:

```hcl
provider "pinecone" {}
```

## Notes

- **Unreleased features**: The local provider includes features not yet in the published version
- **Dev_overrides bypass the registry**: The local provider is used instead of downloading from the registry
- **terraform init works**: You can run `terraform init` without dev_overrides to initialize the workspace
- **Environment variables**: Credentials are loaded from `setup-env.sh` which is gitignored for security
- **Warning messages**: The "Provider development overrides are in effect" warning is expected and normal
- **Simplified workflow**: `setup-env.sh` automatically handles cleanup and provider configuration

## Troubleshooting

If you encounter issues:

1. **Provider not found**: Ensure the binary exists at `../terraform-provider-pinecone`
2. **Authentication errors**: Verify your credentials in `setup-env.sh` are correct
3. **Environment variables not loaded**: Make sure to run `source setup-env.sh` before testing
4. **Previous setup conflicts**: Run `source setup-env.sh` again to clean up
5. **"Resource not supported"**: You're using the registry provider instead of the local one. Make sure to run `source setup-env.sh` first.

## Common Commands

**Test your configuration (with unreleased features):**
```bash
source setup-env.sh && terraform plan
```

**Apply your changes:**
```bash
source setup-env.sh && terraform apply
```

**Clean up:**
```bash
source setup-env.sh && terraform destroy
```
