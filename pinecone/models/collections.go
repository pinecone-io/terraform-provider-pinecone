// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package models

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

// CollectionModel describes the collection data model.
type CollectionModel struct {
	Name        types.String `tfsdk:"name"`
	Size        types.Int64  `tfsdk:"size"`
	Status      types.String `tfsdk:"status"`
	Dimension   types.Int64  `tfsdk:"dimension"`
	VectorCount types.Int64  `tfsdk:"vector_count"`
	Environment types.String `tfsdk:"environment"`
}

func NewCollectionModel(collection *pinecone.Collection) *CollectionModel {
	if collection != nil {
		newCollection := &CollectionModel{
			Name:        types.StringValue(collection.Name),
			Size:        types.Int64Value(*collection.Size),
			Status:      types.StringValue(string(collection.Status)),
			Dimension:   types.Int64Value(int64(*collection.Dimension)),
			VectorCount: types.Int64Value(int64(*collection.VectorCount)),
			Environment: types.StringValue(collection.Environment),
		}
		return newCollection
	}
	return nil
}

// CollectionResourceModel describes the resource data model.
type CollectionResourceModel struct {
	Name      types.String `tfsdk:"name"`
	Size      types.Int64  `tfsdk:"size"`
	Status    types.String `tfsdk:"status"`
	Dimension types.Int64  `tfsdk:"dimension"`
	// VectorCount types.Int64    `tfsdk:"vector_count"`
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
	if collection.Size != nil {
		model.Size = types.Int64Value(*collection.Size)
	}
	if collection.Dimension != nil {
		model.Dimension = types.Int64Value(int64(*collection.Dimension))
	}
	// var vectorCount *int64
	// if collection.VectorCount != nil {
	// 	tempVectorCount := int64(*collection.VectorCount)
	// 	vectorCount = &tempVectorCount
	// }
	// model.VectorCount = types.Int64PointerValue(vectorCount)
}

// CollectionDataSourceModel describes the data source data model.
type CollectionDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Size        types.Int64  `tfsdk:"size"`
	Status      types.String `tfsdk:"status"`
	Dimension   types.Int64  `tfsdk:"dimension"`
	VectorCount types.Int64  `tfsdk:"vector_count"`
	Environment types.String `tfsdk:"environment"`
	Id          types.String `tfsdk:"id"`
}

func (model *CollectionDataSourceModel) Read(collection *pinecone.Collection) {
	model.Id = types.StringValue(collection.Name)
	model.Name = types.StringValue(collection.Name)
	model.Size = types.Int64Value(*collection.Size)
	model.Status = types.StringValue(string(collection.Status))
	model.Dimension = types.Int64Value(int64(*collection.Dimension))
	model.VectorCount = types.Int64Value(int64(*collection.VectorCount))
	model.Environment = types.StringValue(collection.Environment)
}

// CollectionsDataSourceModel describes the data source data model.
type CollectionsDataSourceModel struct {
	Collections []CollectionModel `tfsdk:"collections"`
	Id          types.String      `tfsdk:"id"`
}
