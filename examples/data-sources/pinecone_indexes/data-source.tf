terraform {
  required_providers {
    pinecone = {
      source = "skyscrapr/pinecone"
    }
  }
}

provider "pinecone" {
  environment = "us-west4-gcp"
  api_key     = ""

}

data "pinecone_index" "example" {
  name = "frank"
}
