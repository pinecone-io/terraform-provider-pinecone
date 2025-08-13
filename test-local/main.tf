terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

# Configure the provider with admin credentials
provider "pinecone" {}

# Test create API key resource
resource "pinecone_api_key" "test" {
  name       = "terraform-test-api-key"
  project_id = "7fc0d584-1b12-4c70-872b-281d65426961"  # Update this with your actual project ID
}

# Output the results
output "api_key_id" {
  value = pinecone_api_key.test.id
}

output "api_key_name" {
  value = pinecone_api_key.test.name
}

output "api_key_value" {
  value     = pinecone_api_key.test.key
  sensitive = true
}