# pinecone_api_key

The `pinecone_api_key` resource lets you create and manage API keys in Pinecone. Learn more about API keys in the [docs](https://docs.pinecone.io/guides/authentication/api-keys).

## Example Usage

```hcl
terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}

resource "pinecone_api_key" "example" {
  name       = "example-api-key"
  project_id = "your-project-id"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the API key to be created.
* `project_id` - (Required) The project ID where the API key will be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The API key identifier.
* `key` - The generated API key value. This is sensitive and will not be displayed in logs.

## Import

API keys can be imported using the format `project_id:api_key_id`, for example:

```bash
terraform import pinecone_api_key.example your-project-id:your-api-key-id
```

## Notes

* This resource requires admin client credentials (`client_id` and `client_secret`) to be configured in the provider.
* API keys cannot be updated after creation. Any changes to the `name` or `project_id` will result in the creation of a new API key.
* The API key value is only returned during creation and is not available in subsequent reads for security reasons. 