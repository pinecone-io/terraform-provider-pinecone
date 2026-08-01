# Import a role binding by its ID.
#
# The pinecone_role_bindings data source lists bindings filtered by principal or
# resource, which is how you find the ID of an existing binding.
#
# Bindings whose principal is an invite cannot be imported: this resource does
# not accept principal_type = "invite", because the server moves those bindings
# to the user principal once the invite is accepted. Import the binding after
# acceptance, when it belongs to the user.
terraform import pinecone_role_binding.example 03e8547d-d40f-45eb-b81f-0204e2f9bf0b
