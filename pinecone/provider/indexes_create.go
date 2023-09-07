package pinecone

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	pinecone "github.com/nekomeowww/go-pinecone"
)

type IndexResourceType struct{}

func (r IndexResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, error) {
	return IndexResource{}, nil
}

type IndexResource struct{}

func (r IndexResource) Schema(ctx context.Context) (tfsdk.Schema, diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			// Add other attributes as needed
		},
	}, nil
}

func (r IndexResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Extract the attributes from the request
	name := req.Plan.Get(ctx, tftypes.NewAttributePath().WithAttributeName("name")).(types.String)

	// Call the Pinecone API to create the index
	// You'll need to implement the actual API call here
	payload := map[string]interface{}{
		"name":      name,
		"dimension": dimension,
		"metric":    metric,
	}
	resp, err := client.Post("/indexes", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	id, err := createIndex(client, name, dimension, metric)
	if err != nil {
		return "", err
	}

	// Set the ID for the created resource
	resp.State.Set(ctx, tftypes.NewAttributePath().WithAttributeName("id"), types.String{Value: id})
}

func createIndex(client *pinecone.Client, name string, dimension int, metric string) (string, error) {
	payload := map[string]interface{}{
		"name":      name,
		"dimension": dimension,
		"metric":    metric,
	}
	resp, err := client.Post("/indexes", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	return data["id"].(string), nil
}

// Implement other CRUD operations as needed
