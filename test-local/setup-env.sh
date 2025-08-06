#!/bin/bash

# Set your Pinecone admin credentials
export PINECONE_CLIENT_ID="pinecone-client-id"
export PINECONE_CLIENT_SECRET="pinecone-client-secret"

echo "Environment variables set:"
echo "PINECONE_CLIENT_ID: $PINECONE_CLIENT_ID"
echo "PINECONE_CLIENT_SECRET: $PINECONE_CLIENT_SECRET"
echo "Note: No API key needed for admin operations" 
