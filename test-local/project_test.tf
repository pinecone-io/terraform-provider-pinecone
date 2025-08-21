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

# =============================================================================
# UPDATE PROJECT EXAMPLES - Test various update scenarios
# =============================================================================

# Test Case 1: Update project name
resource "pinecone_project" "update_name" {
  name = "terraform-updated-name-project"
}

# Test Case 2: Update max_pods (add to existing project)
resource "pinecone_project" "update_pods" {
  name     = "terraform-update-pods-project"
  max_pods = 15  # Increased from default 0
}

# Test Case 3: Update both name and max_pods
resource "pinecone_project" "update_multiple" {
  name     = "terraform-updated-multiple-project"
  max_pods = 8
}

# Test Case 4: Update project with CMEK (can only be enabled, not disabled)
resource "pinecone_project" "update_cmek" {
  name                        = "terraform-update-cmek-project"
  force_encryption_with_cmek  = true
  max_pods                    = 12
}

# Test Case 5: Update existing project with all attributes
resource "pinecone_project" "update_comprehensive" {
  name                        = "terraform-comprehensive-update-project"
  force_encryption_with_cmek  = true
  max_pods                    = 20
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

# =============================================================================
# OUTPUTS - Update Test Results
# =============================================================================

output "update_name_project_id" {
  description = "The ID of the updated name project"
  value       = pinecone_project.update_name.id
}

output "update_name_project_name" {
  description = "The updated name of the project"
  value       = pinecone_project.update_name.name
}

output "update_pods_project_id" {
  description = "The ID of the updated pods project"
  value       = pinecone_project.update_pods.id
}

output "update_pods_project_max_pods" {
  description = "The updated max_pods configuration"
  value       = pinecone_project.update_pods.max_pods
}

output "update_multiple_project_id" {
  description = "The ID of the multiple update project"
  value       = pinecone_project.update_multiple.id
}

output "update_multiple_project_name" {
  description = "The updated name of the multiple update project"
  value       = pinecone_project.update_multiple.name
}

output "update_multiple_project_max_pods" {
  description = "The updated max_pods of the multiple update project"
  value       = pinecone_project.update_multiple.max_pods
}

output "update_cmek_project_id" {
  description = "The ID of the CMEK update project"
  value       = pinecone_project.update_cmek.id
}

output "update_cmek_project_cmek_enabled" {
  description = "Whether CMEK encryption is enabled"
  value       = pinecone_project.update_cmek.force_encryption_with_cmek
}

output "update_comprehensive_project_id" {
  description = "The ID of the comprehensive update project"
  value       = pinecone_project.update_comprehensive.id
}

output "update_comprehensive_project_name" {
  description = "The name of the comprehensive update project"
  value       = pinecone_project.update_comprehensive.name
}

output "update_comprehensive_project_max_pods" {
  description = "The max_pods of the comprehensive update project"
  value       = pinecone_project.update_comprehensive.max_pods
}
