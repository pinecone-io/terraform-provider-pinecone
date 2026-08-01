# Users cannot be created through Terraform — they join the organization by
# accepting an invite — so importing is the only way to bring one under
# management. Once imported, destroying the resource removes the user from the
# organization.
#
# Find the user's ID with the pinecone_user data source, which accepts an email.
terraform import pinecone_user.example 495a1437-79de-4121-9326-c0d3aa6090f8
