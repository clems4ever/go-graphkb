package graphkb

import "github.com/clems4ever/go-graphkb/internal/client"

// GraphAPI is the representation of the graphkb API exposed to data sources.
type GraphAPI = client.GraphAPI

// GraphAPIOptions are the options provided to GraphAPI
type GraphAPIOptions = client.GraphAPIOptions

// NewGraphAPI creates a new graph API
var NewGraphAPI = client.NewGraphAPI
