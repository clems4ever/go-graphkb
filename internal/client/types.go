package client

import (
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
)

// PutGraphSchemaRequestBody a request body for the schema update
type PutGraphSchemaRequestBody struct {
	Schema schema.SchemaGraph `json:"schema"`
}

// PutGraphAssetRequestBody a request body for the asset upsert
type PutGraphAssetRequestBody struct {
	Assets []knowledge.Asset `json:"assets"`
}

// PutGraphRelationRequestBody a request body for the relation upsert
type PutGraphRelationRequestBody struct {
	Relations []knowledge.Relation `json:"relations"`
}

// DeleteGraphAssetRequestBody a request body for the asset removal
type DeleteGraphAssetRequestBody struct {
	Assets []knowledge.Asset `json:"assets"`
}

// DeleteGraphRelationRequestBody a request body for the relation removal
type DeleteGraphRelationRequestBody struct {
	Relations []knowledge.Relation `json:"relations"`
}

type QueryRequestBody struct {
	Q              string `json:"q"`
	IncludeSources bool   `json:"include_sources"`
}

type QueryResponse struct {
	Columns []Column
	Items   [][]Item
}

// Column a column as returned by the graphdb api.
type Column struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type Item map[string]any

func (ra Item) Asset() knowledge.AssetWithID {
	return knowledge.AssetWithID{
		ID: ra["_id"].(string),
		Asset: knowledge.Asset{
			Type: schema.AssetType(ra["type"].(string)),
			Key:  ra["key"].(string),
		},
	}
}

func (ra Item) Relation() knowledge.RelationWithID {
	return knowledge.RelationWithID{
		ID:   ra["_id"].(string),
		From: ra["from_id"].(string),
		To:   ra["to_id"].(string),
		Type: schema.RelationKeyType(ra["type"].(string)),
	}
}

func (ra Item) Sources() []string {
	s, _ := ra["sources"].([]string)
	return s
}
