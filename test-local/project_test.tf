# Create a test project
resource "pinecone_project" "test" {
  name = "terraform-test-project"
}

# Create a project with CMEK encryption
resource "pinecone_project" "encrypted" {
  name                        = "terraform-encrypted-project"
  force_encryption_with_cmek  = true
}

# Create a project with custom max pods
resource "pinecone_project" "custom_pods" {
  name     = "terraform-custom-pods-project"
  max_pods = 5
}

output "test_project_id" {
  description = "The ID of the test project"
  value       = pinecone_project.test.id
}

output "test_project_name" {
  description = "The name of the test project"
  value       = pinecone_project.test.name
}

output "encrypted_project_id" {
  description = "The ID of the encrypted project"
  value       = pinecone_project.encrypted.id
}

output "custom_pods_project_id" {
  description = "The ID of the custom pods project"
  value       = pinecone_project.custom_pods.id
}
