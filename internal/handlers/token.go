package handlers

import (
	"fmt"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/clems4ever/go-graphkb/internal/utils"
)

// IsTokenValid is the token valid
func IsTokenValid(registry sources.Registry, r *http.Request) (bool, string, error) {
	token := r.Header.Get(utils.XAuthTokenHeader)

	if token == "" {
		return false, "", fmt.Errorf("No auth token provided")
	}

	sourceToToken, err := registry.ListSources(r.Context())

	if err != nil {
		return false, "", fmt.Errorf("Unable to list the sources: %v", err)
	}

	for sn, t := range sourceToToken {
		if t == token {
			return true, sn, nil
		}
	}

	return false, "", nil
}
