# Import a pending invite by its ID.
#
# Note that role_bindings cannot be imported: the invite endpoint does not return
# the roles it granted, so the attribute stays empty in state and the next plan
# proposes a replacement until you set it to match the original invite. Use the
# pinecone_role_bindings data source with principal_type = "invite" to see what
# the invite granted, and copy those values into your configuration.
terraform import pinecone_invite.example e7b2e42d-b5b7-4e67-8a07-40384c129f58
