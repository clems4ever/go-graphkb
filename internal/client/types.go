package client

import (
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
)

// GraphUpdateRequestBody a request body for the graph update API
type GraphUpdateRequestBody struct {
	Updates *knowledge.GraphUpdatesBulk `json:"updates"`
	Schema  schema.SchemaGraph          `json:"schema"`
}

// PutGraphSchemaRequestBody a request body for the schema update
type PutGraphSchemaRequestBody struct {
	Schema schema.SchemaGraph `json:"schema"`
}

// PutGraphAssetRequestBody a request body for the asset upsert
type PutGraphAssetRequestBody struct {
	Asset knowledge.Asset `json:"asset"`
}

// PutGraphRelationRequestBody a request body for the relation upsert
type PutGraphRelationRequestBody struct {
	Relation knowledge.Relation `json:"relation"`
}

// DeleteGraphAssetRequestBody a request body for the asset removal
type DeleteGraphAssetRequestBody struct {
	Asset knowledge.Asset `json:"asset"`
}

// DeleteGraphRelationRequestBody a request body for the relation removal
type DeleteGraphRelationRequestBody struct {
	Relation knowledge.Relation `json:"relation"`
}
