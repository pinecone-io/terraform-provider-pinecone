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
			"dimension": {
				Type:     types.IntType,
				Required: true,
			},
			"metric": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (r IndexResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Extract the attributes from the request
	name := req.Plan.Get(ctx, tftypes.NewAttributePath().WithAttributeName("name")).(types.String)

	// Initialize the Pinecone client
	// Assuming you have some way to configure the client, like an API key or other authentication method
	client := pinecone.NewClient("<YOUR_API_KEY_OR_CONFIG>")

	// Prepare the payload for the API request
	payload := map[string]interface{}{
		"name":      name.Value,
		"dimension": req.Plan.Get(ctx, tftypes.NewAttributePath().WithAttributeName("dimension")).(types.Int).Value,
		"metric":    req.Plan.Get(ctx, tftypes.NewAttributePath().WithAttributeName("metric")).(types.String).Value,
	}

	// Call the Pinecone API to create the index
	response, err := client.Post("/indexes", payload)
	if err != nil {
		// Handle the error, maybe set a diagnostic in the response
		resp.Diagnostics.AddError("API Call Failed", err.Error())
		return
	}
	defer response.Body.Close()

	// Decode the API response
	var data map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		// Handle the error, maybe set a diagnostic in the response
		resp.Diagnostics.AddError("Failed to decode API response", err.Error())
		return
	}

	// Extract the ID from the API response and set it for the created resource
	if id, ok := data["id"].(string); ok {
		resp.State.Set(ctx, tftypes.NewAttributePath().WithAttributeName("id"), types.String{Value: id})
	} else {
		// Handle the case where the ID is not found or is not a string
		resp.Diagnostics.AddError("Invalid ID in API response", "The API response did not contain a valid ID.")
	}
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
