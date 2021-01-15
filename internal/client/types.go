package client

import (
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
)

// GraphUpdateRequestBody a request body for the graph update api
type GraphUpdateRequestBody struct {
	Updates *knowledge.GraphUpdatesBulk `json:"updates"`
	Schema  schema.SchemaGraph          `json:"schema"`
}
