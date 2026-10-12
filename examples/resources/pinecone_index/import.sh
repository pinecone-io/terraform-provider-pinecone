# Import an index by name. Document indexes are recorded in the schema style, and every other index in
# the spec style. If the configuration describes an imported vector index in the schema style, the
# next plan moves it there with an in-place update that only changes Terraform state.
terraform import pinecone_index.example my-index
