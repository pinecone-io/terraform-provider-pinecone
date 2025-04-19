// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

// CollectionModel describes the collection data model.
type CollectionModel struct {
	Name        types.String `tfsdk:"name"`
	Size        types.Int64  `tfsdk:"size"`
	Status      types.String `tfsdk:"status"`
	Dimension   types.Int32  `tfsdk:"dimension"`
	VectorCount types.Int32  `tfsdk:"vector_count"`
	Environment types.String `tfsdk:"environment"`
}

func NewCollectionModel(collection *pinecone.Collection) *CollectionModel {
	if collection != nil {
		newCollection := &CollectionModel{
			Name:        types.StringValue(collection.Name),
			Status:      types.StringValue(string(collection.Status)),
			Environment: types.StringValue(collection.Environment),
			Size:        types.Int64Value(collection.Size),
			Dimension:   types.Int32Value(collection.Dimension),
		}
		return newCollection
	}
	return nil
}

// CollectionResourceModel describes the resource data model.
type CollectionResourceModel struct {
	Name        types.String   `tfsdk:"name"`
	Size        types.Int64    `tfsdk:"size"`
	Status      types.String   `tfsdk:"status"`
	Dimension   types.Int32    `tfsdk:"dimension"`
	VectorCount types.Int32    `tfsdk:"vector_count"`
	Environment types.String   `tfsdk:"environment"`
	Id          types.String   `tfsdk:"id"`
	Source      types.String   `tfsdk:"source"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (model *CollectionResourceModel) Read(collection *pinecone.Collection) {
	model.Id = types.StringValue(collection.Name)
	model.Name = types.StringValue(collection.Name)
	model.Source = types.StringValue(model.Source.ValueString())
	model.Status = types.StringValue(string(collection.Status))
	model.Environment = types.StringValue(collection.Environment)
	model.Size = types.Int64Value(collection.Size)
	model.Dimension = types.Int32Value(collection.Dimension)
	model.VectorCount = types.Int32Value(collection.VectorCount)
}

// CollectionDataSourceModel describes the data source data model.
type CollectionDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Size        types.Int64  `tfsdk:"size"`
	Status      types.String `tfsdk:"status"`
	Dimension   types.Int32  `tfsdk:"dimension"`
	VectorCount types.Int32  `tfsdk:"vector_count"`
	Environment types.String `tfsdk:"environment"`
	Id          types.String `tfsdk:"id"`
}

func (model *CollectionDataSourceModel) Read(collection *pinecone.Collection) {
	model.Id = types.StringValue(collection.Name)
	model.Name = types.StringValue(collection.Name)
	model.Status = types.StringValue(string(collection.Status))
	model.Environment = types.StringValue(collection.Environment)
	model.Size = types.Int64Value(collection.Size)
	model.Dimension = types.Int32Value(collection.Dimension)
	model.VectorCount = types.Int32Value(collection.VectorCount)
}

// CollectionsDataSourceModel describes the data source data model.
type CollectionsDataSourceModel struct {
	Collections []CollectionModel `tfsdk:"collections"`
	Id          types.String      `tfsdk:"id"`
}
