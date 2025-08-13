# Test Local Environment

This directory contains a local testing environment for the Terraform Pinecone provider. It allows you to test the provider functionality locally before deploying to production.

## Purpose

The `test-local` directory is designed for:
- Local development and testing of the Pinecone Terraform provider
- Validating provider functionality with real Pinecone resources
- Rapid iteration during provider development
- Testing different resource configurations safely (see `examples/` folder for other resource types)

## Files

- `main.tf` - Terraform configuration for testing API key creation
- `setup-env.sh` - Script to set up environment variables for Pinecone admin credentials
- `terraform.tfstate` - Terraform state file (ignored by git)

## Setup Instructions

### 1. Configure Pinecone Admin Credentials

Before running the tests, you need to set up your Pinecone admin credentials:

1. Edit `setup-env.sh` and replace the placeholder values:
   ```bash
   export PINECONE_CLIENT_ID="your-actual-client-id"
   export PINECONE_CLIENT_SECRET="your-actual-client-secret"
   ```
**Note**: If you plan to test index creation or other resources (not just API key management), you may also need to set the `PINECONE_API_KEY` environment variable:
   ```bash
   export PINECONE_API_KEY="your-pinecone-api-key"
   ```

2. Source the environment variables:
   ```bash
   source setup-env.sh
   ```

### 2. Update Project ID

Edit `main.tf` and replace the placeholder project ID:
```hcl
project_id = "your-actual-project-id"
```

### 3. Initialize Terraform

```bash
terraform init
```

### 4. Test the Provider

```bash
# Plan the changes
terraform plan

# Apply the changes (creates test API key)
terraform apply

# Destroy the test resources
terraform destroy
```

## Cleanup

Always clean up test resources when you're done:

```bash
terraform destroy
```

This will remove the test API key and any other resources created during testing.
