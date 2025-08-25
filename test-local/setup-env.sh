#!/bin/bash

# Clean up previous setup
echo "Cleaning up previous setup..."
rm -f .terraform
rm -rf .terraform .terraform.lock.hcl
echo "Cleanup complete!"

# Set your Pinecone admin credentials, changes in this file will not be committed to the repo
export PINECONE_CLIENT_ID="pinecone-client-id"
export PINECONE_CLIENT_SECRET="pinecone-client-secret"

echo "Environment variables set:"
echo "PINECONE_CLIENT_ID: $PINECONE_CLIENT_ID"
echo "PINECONE_CLIENT_SECRET: $PINECONE_CLIENT_SECRET"
echo "Note: No API key needed for admin operations"
