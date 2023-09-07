provider "pinecone" {
  // Any required provider configuration parameters, e.g., API key, endpoint, etc.
  api_key = "var.api_key"
}

resource "pinecone_index" "example_index" {
  name = "my_example_index"
  // Any other attributes or configurations specific to the index creation.
}
