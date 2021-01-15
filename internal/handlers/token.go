package handlers

import (
	"fmt"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/sources"
)

// IsTokenValid is the token valid
func IsTokenValid(registry sources.Registry, r *http.Request) (bool, string, error) {
	token, ok := r.URL.Query()["token"]

	if !ok || len(token) != 1 {
		return false, "", fmt.Errorf("Unable to detect token query parameter")
	}

	sourceToToken, err := registry.ListSources(r.Context())

	if err != nil {
		return false, "", err
	}

	for sn, t := range sourceToToken {
		if t == token[0] {
			return true, sn, nil
		}
	}

	return false, "", nil
}
