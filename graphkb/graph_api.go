package graphkb

import (
	"github.com/clems4ever/go-graphkb/internal/client"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
)

// GraphAPI is the representation of the graphkb API exposed to data sources.
type GraphAPI = client.GraphAPI

// GraphAPIOptions are the options provided to GraphAPI
type GraphAPIOptions = client.GraphAPIOptions

// NewGraphAPI creates a new graph API
var NewGraphAPI = client.NewGraphAPI

// QueryResponse is the response from the GraphAPI query
type QueryResponse = client.QueryResponse

// QueryRequestBody is the request body for a GraphAPI query
type QueryRequestBody = client.QueryRequestBody

// Column is a column from the QueryResponse
type Column = client.Column

// Item is map with the members of a row item in the QueryResponse
type Item = client.Item

type AssetWithID = knowledge.AssetWithID

type RelationWithID = knowledge.RelationWithID

type Property = knowledge.Property

// PutGraphSchemaRequestBody a request body for the schema update
type PutGraphSchemaRequestBody = client.PutGraphSchemaRequestBody

// PutGraphAssetRequestBody a request body for the asset upsert
type PutGraphAssetRequestBody = client.PutGraphAssetRequestBody

// PutGraphRelationRequestBody a request body for the relation upsert
type PutGraphRelationRequestBody = client.PutGraphRelationRequestBody

// DeleteGraphAssetRequestBody a request body for the asset removal
type DeleteGraphAssetRequestBody = client.DeleteGraphAssetRequestBody

// DeleteGraphRelationRequestBody a request body for the relation removal
type DeleteGraphRelationRequestBody = client.DeleteGraphRelationRequestBody
