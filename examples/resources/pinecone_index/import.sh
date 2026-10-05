# Import an index by name. Document indexes are imported with schema and deployment; vector
# indexes, including those with integrated embedding, are imported with spec.
terraform import pinecone_index.example my-index
