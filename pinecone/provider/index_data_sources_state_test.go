// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/pinecone-io/go-pinecone/v7/pinecone"
	"github.com/pinecone-io/terraform-provider-pinecone/pinecone/models"
)

func detailedTestIndex() *pinecone.Index {
	dimension := int32(1024)
	metric := pinecone.IndexMetricCosine
	stemming, filterable := true, true
	nodeType := "b1"
	replicas, shards := int32(1), int32(1)
	privateHost, cmekId := "my-index.private.example.com", "cmek-123"
	return &pinecone.Index{
		Name:        "my-index",
		Host:        "https://my-index.example.com",
		PrivateHost: &privateHost,
		CmekId:      &cmekId,
		Metric:      metric,
		VectorType:  "dense",
		Dimension:   &dimension,
		Schema: &pinecone.IndexSchema{Fields: map[string]pinecone.IndexSchemaField{
			"_values":        {DenseVector: &pinecone.DenseVectorField{Dimension: dimension, Metric: metric}},
			"_sparse_values": {SparseVector: &pinecone.SparseVectorField{}},
			"chunk_text": {SemanticText: &pinecone.SemanticTextField{
				Model: "multilingual-e5-large", Dimension: &dimension, Metric: &metric,
				ReadParameters: &map[string]interface{}{"input_type": "query"},
			}},
			"body":  {String: &pinecone.StringField{FullTextSearch: &pinecone.FullTextSearchConfig{Stemming: &stemming, Ngram: &pinecone.NgramConfig{MinGram: 2, MaxGram: 4}}}},
			"tags":  {StringList: &pinecone.StringListField{Filterable: &filterable}},
			"draft": {Boolean: &pinecone.BooleanField{Filterable: &filterable}},
			"score": {Float: &pinecone.FloatField{Filterable: &filterable}},
			"year":  {Integer: &pinecone.IntegerField{Filterable: &filterable}},
			"genre": {LegacyMetadata: &pinecone.LegacyMetadataField{Filterable: true}},
		}},
		Deployment: &pinecone.IndexDeployment{Managed: &pinecone.ManagedDeployment{Cloud: pinecone.CloudAWS, Region: "us-east-1"}},
		ReadCapacity: &pinecone.ReadCapacity{Dedicated: &pinecone.ReadCapacityDedicated{
			NodeType: &nodeType,
			Scaling:  &pinecone.ReadCapacityScaling{Manual: &pinecone.ReadCapacityManualScaling{Replicas: &replicas, Shards: &shards}},
			Status:   pinecone.ReadCapacityStatus{State: "Ready", CurrentReplicas: &replicas, CurrentShards: &shards},
		}},
		Spec:               &pinecone.IndexSpec{Serverless: &pinecone.ServerlessSpec{Cloud: pinecone.CloudAWS, Region: "us-east-1"}},
		Embed:              &pinecone.IndexEmbed{Model: "multilingual-e5-large", Dimension: &dimension, Metric: &metric, FieldMap: &map[string]interface{}{"text": "chunk_text"}},
		Status:             &pinecone.IndexStatus{Ready: true, State: pinecone.IndexStatusStateReady},
		DeletionProtection: pinecone.DeletionProtectionDisabled,
	}
}

func dataSourceState(t *testing.T, d datasource.DataSource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	var resp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("building schema: %v", resp.Diagnostics)
	}
	return tfsdk.State{Schema: resp.Schema, Raw: tftypes.NewValue(resp.Schema.Type().TerraformType(ctx), nil)}
}

func checkStringAttribute(t *testing.T, state tfsdk.State, p path.Path, want string) {
	t.Helper()
	var got types.String
	if d := state.GetAttribute(context.Background(), p, &got); d.HasError() {
		t.Fatalf("reading %s: %v", p, d)
	}
	if got.ValueString() != want {
		t.Errorf("%s = %q, want %q", p, got.ValueString(), want)
	}
}

func TestIndexDataSource_stateMatchesSchema(t *testing.T) {
	ctx := context.Background()
	var model models.IndexDatasourceModel
	if d := model.Read(ctx, detailedTestIndex()); d.HasError() {
		t.Fatalf("Read returned errors: %v", d)
	}

	state := dataSourceState(t, NewIndexDataSource())
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatalf("model doesn't fit the data source schema: %v", d)
	}

	fields := path.Root("schema").AtName("fields")
	checkStringAttribute(t, state, fields.AtMapKey("_values").AtName("dense_vector").AtName("metric"), "cosine")
	checkStringAttribute(t, state, fields.AtMapKey("chunk_text").AtName("semantic_text").AtName("model"), "multilingual-e5-large")
	checkStringAttribute(t, state, fields.AtMapKey("chunk_text").AtName("semantic_text").AtName("read_parameters").AtMapKey("input_type"), "query")
	checkStringAttribute(t, state, path.Root("deployment").AtName("managed").AtName("region"), "us-east-1")
	checkStringAttribute(t, state, path.Root("read_capacity").AtName("dedicated").AtName("state"), "Ready")
	checkStringAttribute(t, state, path.Root("private_host"), "my-index.private.example.com")
	checkStringAttribute(t, state, path.Root("cmek_id"), "cmek-123")

	var minGram types.Int64
	if d := state.GetAttribute(ctx, fields.AtMapKey("body").AtName("string").AtName("full_text_search").AtName("ngram").AtName("min_gram"), &minGram); d.HasError() || minGram.ValueInt64() != 2 {
		t.Errorf("body ngram min_gram = %v (%v), want 2", minGram, d)
	}
	var sourceBackupId types.String
	if d := state.GetAttribute(ctx, path.Root("source_backup_id"), &sourceBackupId); d.HasError() || !sourceBackupId.IsNull() {
		t.Errorf("source_backup_id = %v (%v), want null", sourceBackupId, d)
	}
}

func TestIndexesDataSource_stateMatchesSchema(t *testing.T) {
	ctx := context.Background()
	var index models.IndexModel
	if d := index.Read(ctx, detailedTestIndex()); d.HasError() {
		t.Fatalf("Read returned errors: %v", d)
	}

	state := dataSourceState(t, NewIndexesDataSource())
	data := models.IndexesDataSourceModel{Indexes: []models.IndexModel{index}, Id: types.StringValue("test")}
	if d := state.Set(ctx, &data); d.HasError() {
		t.Fatalf("model doesn't fit the data source schema: %v", d)
	}

	first := path.Root("indexes").AtListIndex(0)
	var genreFilterable types.Bool
	if d := state.GetAttribute(ctx, first.AtName("schema").AtName("fields").AtMapKey("genre").AtName("legacy_metadata").AtName("filterable"), &genreFilterable); d.HasError() || !genreFilterable.ValueBool() {
		t.Errorf("genre filterable = %v (%v), want true", genreFilterable, d)
	}
	checkStringAttribute(t, state, first.AtName("deployment").AtName("managed").AtName("cloud"), "aws")
	checkStringAttribute(t, state, first.AtName("deletion_protection"), "disabled")

	var fields types.Map
	if d := state.GetAttribute(ctx, first.AtName("schema").AtName("fields"), &fields); d.HasError() {
		t.Fatalf("reading fields: %v", d)
	}
	if got := len(fields.Elements()); got != 9 {
		t.Errorf("got %d schema fields, want 9", got)
	}
}
