package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	pinecone "github.com/nekomeowww/go-pinecone"
)

func NewIndexResource() resource.Resource {
	return &IndexResource{}
}

func (r *IndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// This resource does not support reads.
	resp.Diagnostics.AddError("Read not supported", "This resource does not support reads.")
}

// IndexResource defines the resource implementation.
type IndexResource struct {
	client *pinecone.Client
}

// IndexResourceModel describes the resource data model.
type IndexResourceModel struct {
	Name      types.String `tfsdk:"name"`
}
func (r *IndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IndexResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set up the index configuration
	config := pinecone.IndexConfig{
		Name:      string(data.Name),
		Dimension: int(data.Dimension),
		Metric:    string(data.Metric),
	}

	// Create the index
	index, err := r.client.CreateIndex(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create index, got error: %s", err))
		return
	}

	// Save the index ID to the Terraform state
	resp.State.Set(ctx, "id", index.ID)
}

}

func (r *IndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource does not support updates.
	resp.Diagnostics.AddError("Update not supported", "This resource does not support updates.")
}

func (r *IndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// This resource does not support deletes.
	resp.Diagnostics.AddError("Delete not supported", "This resource does not support deletes.")
}

if resp.Diagnostics.HasError() {
	return
}
        return
    }

    // Set up the index configuration
    config := pinecone.IndexConfig{
        Name:      string(data.Name),
        Dimension: int(data.Dimension),
        Metric:    string(data.Metric),
    }

    // Create the index
    index, err := r.client.CreateIndex(ctx, config)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create index, got error: %s", err))
        return
    }

    // Save the index ID to the Terraform state
    resp.State.Set(ctx, "id", index.ID)
}

func (r *IndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // This resource does not support updates.
    resp.Diagnostics.AddError("Update not supported", "This resource does not support updates.")
}

func (r *IndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // This resource does not support deletes.
    resp.Diagnostics.AddError("Delete not supported", "This resource does not support deletes.")
}