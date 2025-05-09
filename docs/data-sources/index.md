---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pinecone_index Data Source - terraform-provider-pinecone"
subcategory: ""
description: |-
  Index data source
---

# pinecone_index (Data Source)

Index data source

## Example Usage

```terraform
terraform {
  required_providers {
    pinecone = {
      source = "pinecone-io/pinecone"
    }
  }
}

provider "pinecone" {}

resource "pinecone_index" "test" {
  name      = "tftestindex"
  metric    = "cosine"
  dimension = 1536
  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-west-2"
    }
  }
}

data "pinecone_index" "test" {
  name = pinecone_index.test.name
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Index name

### Optional

- `embed` (Attributes) Specify the integrated inference embedding configuration for the index. Once set, the model cannot be changed. However, you can later update the embedding configuration—including field map, read parameters, and write parameters.

Refer to the [model guide](https://docs.pinecone.io/guides/inference/understanding-inference#embedding-models) for available models and details. (see [below for nested schema](#nestedatt--embed))
- `spec` (Attributes) Spec (see [below for nested schema](#nestedatt--spec))
- `status` (Attributes) Configuration for the behavior of Pinecone's internal metadata index. By default, all metadata is indexed; when metadata_config is present, only specified metadata fields are indexed. To specify metadata fields to index, provide an array of the following form: [example_metadata_field] (see [below for nested schema](#nestedatt--status))

### Read-Only

- `deletion_protection` (String) Index deletion protection can be one of 'enabled' or 'disabled'.
- `dimension` (Number) Index dimension
- `host` (String) The URL address where the index is hosted.
- `id` (String) Index identifier
- `metric` (String) Index metric can be one of 'cosine', 'dotproduct', or 'euclidean'.
- `tags` (Map of String) Custom user tags added to an index. Keys must be 80 characters or less. Values must be 120 characters or less. Keys must be alphanumeric, '', or '-'. Values must be alphanumeric, ';', '@', '', '-', '.', '+', or ' '. To unset a key, set the value to be an empty string.
- `vector_type` (String) Index vector type, for example 'dense' or 'sprase'.

<a id="nestedatt--embed"></a>
### Nested Schema for `embed`

Read-Only:

- `dimension` (Number) The dimension of the embedding model, specifying the size of the output vector.
- `field_map` (Map of String) Identifies the name of the text field from your document model that will be embedded.
- `metric` (String) The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'. If the 'vector_type' is 'sparse', the metric must be 'dotproduct'. If the vector_type is dense, the metric defaults to 'cosine'.
- `model` (String) the name of the embedding model to use for the index.
- `read_parameters` (Map of String) The read parameters for the embedding model.
- `vector_type` (String) The index vector type associated with the model. If 'dense', the vector dimension must be specified. If 'sparse', the vector dimension will be nil.
- `write_parameters` (Map of String) The write parameters for the embedding model.


<a id="nestedatt--spec"></a>
### Nested Schema for `spec`

Optional:

- `pod` (Attributes) Configuration needed to deploy a pod-based index. (see [below for nested schema](#nestedatt--spec--pod))
- `serverless` (Attributes) Configuration needed to deploy a serverless index. (see [below for nested schema](#nestedatt--spec--serverless))

<a id="nestedatt--spec--pod"></a>
### Nested Schema for `spec.pod`

Optional:

- `metadata_config` (Attributes) Configuration for the behavior of Pinecone's internal metadata index. By default, all metadata is indexed; when metadata_config is present, only specified metadata fields are indexed. These configurations are only valid for use with pod-based indexes. (see [below for nested schema](#nestedatt--spec--pod--metadata_config))

Read-Only:

- `environment` (String) The environment where the index is hosted.
- `pod_type` (String) The type of pod to use. One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8.
- `pods` (Number) The number of pods to be used in the index. This should be equal to shards x replicas.'
- `replicas` (Number) The number of replicas. Replicas duplicate your index. They provide higher availability and throughput. Replicas can be scaled up or down as your needs change.
- `shards` (Number) The number of shards. Shards split your data across multiple pods so you can fit more data into an index.
- `source_collection` (String) The name of the collection to create an index from.

<a id="nestedatt--spec--pod--metadata_config"></a>
### Nested Schema for `spec.pod.metadata_config`

Read-Only:

- `indexed` (List of String) The indexed fields.



<a id="nestedatt--spec--serverless"></a>
### Nested Schema for `spec.serverless`

Read-Only:

- `cloud` (String) Ready.
- `region` (String) Initializing InitializationFailed ScalingUp ScalingDown ScalingUpPodSize ScalingDownPodSize Upgrading Terminating Ready



<a id="nestedatt--status"></a>
### Nested Schema for `status`

Read-Only:

- `ready` (Boolean) Ready.
- `state` (String) Initializing InitializationFailed ScalingUp ScalingDown ScalingUpPodSize ScalingDownPodSize Upgrading Terminating Ready
