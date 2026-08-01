# Import a service account by its ID.
#
# Note that client_secret cannot be imported: the API returns it only once, at
# creation or rotation, so it stays empty in state for an imported service
# account. Change rotate_trigger to issue and store a new secret.
terraform import pinecone_service_account.example 8c1d4e7a-3f92-4b18-9d5c-6a2e0b7f1c34
