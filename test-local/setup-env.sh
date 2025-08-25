#!/bin/bash

# Clean up previous setup
echo "Cleaning up previous setup..."
rm -f .terraform
rm -rf .terraform .terraform.lock.hcl terraform.tfstate*
echo "Cleanup complete!"

# Set your Pinecone admin credentials, changes in this file will not be committed to the repo
export PINECONE_CLIENT_ID="pinecone-client-id"
export PINECONE_CLIENT_SECRET="pinecone-client-secret"

# Configure Terraform to use local provider
export TF_CLI_CONFIG_FILE=.terraformrc

echo "Environment variables set:"
echo "PINECONE_CLIENT_ID: $PINECONE_CLIENT_ID"
echo "PINECONE_CLIENT_SECRET: $PINECONE_CLIENT_SECRET"
echo "TF_CLI_CONFIG_FILE: $TF_CLI_CONFIG_FILE"
echo "Note: No API key needed for admin operations"
