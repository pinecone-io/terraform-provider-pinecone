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

**Note**: You can modify `main.tf` to test any resource type. Refer to the `../examples/` folder for examples of how to create different resources:
- `../examples/resources/pinecone_index/` - For index creation
- `../examples/resources/pinecone_collection/` - For collection creation
- `../examples/data-sources/` - For data source usage

### Step 3: Initialize with Published Provider

```bash
terraform init
```

This downloads the published provider and creates the necessary directory structure.

### Step 4: Replace with Local Binary

```bash
cp ../terraform-provider-pinecone .terraform/providers/registry.terraform.io/pinecone-io/pinecone/1.0.0/darwin_arm64/terraform-provider-pinecone
```

This replaces the published provider binary with your local development version.

### Step 5: Test Your Changes

```bash
terraform plan
```

Now you can test your local provider changes!

## Configuration

The provider will automatically pick up credentials from environment variables:

- `PINECONE_CLIENT_ID` - Your Pinecone client ID
- `PINECONE_CLIENT_SECRET` - Your Pinecone client secret

The `main.tf` file uses an empty provider block since credentials are loaded from environment variables:

```hcl
provider "pinecone" {}
```

## Notes

- The local provider will have access to new features not yet in the published version
- If you get checksum errors, remove `.terraform.lock.hcl` and re-run `terraform init`
- This setup allows you to test new resources like `pinecone_api_key` before they're published
- Credentials are stored in `setup-env.sh` which is gitignored for security
- You can test any resource type by updating `main.tf` - see the `../examples/` folder for reference
- The warning about missing `.terraformrc` file is expected and harmless
- The `setup-env.sh` script automatically cleans up previous setup

## Troubleshooting

If you encounter issues:

1. **Checksum errors**: Remove `.terraform.lock.hcl` and re-run `terraform init`
2. **Provider not found**: Ensure the binary path is correct for your OS/architecture
3. **Authentication errors**: Verify your environment variables are set correctly
4. **Environment variables not loaded**: Make sure to run `source setup-env.sh` before testing
5. **Previous setup conflicts**: Run `source setup-env.sh` again to clean up
