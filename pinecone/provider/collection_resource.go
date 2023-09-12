package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/skyscrapr/pinecone-sdk-go/pinecone"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CollectionResource{}
var _ resource.ResourceWithImportState = &CollectionResource{}

func NewCollectionResource() resource.Resource {
	return &CollectionResource{}
}

// CollectionResource defines the resource implementation.
type CollectionResource struct {
	client *pinecone.Client
}

// CollectionResourceModel describes the resource data model.
type CollectionResourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// Add other attributes as needed
}

func (r *CollectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

func (r *CollectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Collection resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Collection identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the collection to be created.",
				Required:            true,
			},
			// Add other attributes as needed
		},
	}
}

func (r *CollectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pinecone.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *CollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CollectionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use the Pinecone SDK to create the collection
	err := r.client.CreateCollection(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create collection", err.Error())
		return
	}

	// Set the ID for the resource
	resp.State.SetAttribute(ctx, "id", types.String{Value: data.Name.ValueString()})
}

func (r *CollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CollectionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use the Pinecone SDK to describe the collection
	collection, err := r.client.DescribeCollection(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe collection", err.Error())
		return
	}

	readCollectionData(collection, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implement update logic if needed
}

func (r *CollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CollectionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCollection(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete collection", err.Error())
		return
	}
}

func (r *CollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readCollectionData(collection *pinecone.Collection, model *CollectionResourceModel) {
	model.Id = types.StringValue(collection.Name)
	model.Name = types.StringValue(collection.Name)
	// Add other attributes as needed
}
