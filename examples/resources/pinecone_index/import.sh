# Import an index by name. Document indexes are recorded with schema and deployment, and other indexes
# with spec. If the configuration describes an imported vector index with schema and deployment, the
# next plan moves it there in place, without changing the index.
terraform import pinecone_index.example my-index
