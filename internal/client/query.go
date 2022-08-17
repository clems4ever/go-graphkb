package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (gapi *GraphAPI) Query(ctx context.Context, q string, includeSources bool) (*QueryResponse, error) {
	b, err := json.Marshal(QueryRequestBody{
		Q:              q,
		IncludeSources: includeSources,
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to marshall request body")
	}

	req, err := gapi.client.newRequest(ctx, "POST", "/api/query", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	res, err := gapi.client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode == http.StatusTooManyRequests {
		return nil, ErrTooManyRequests
	} else if res.StatusCode != http.StatusOK {
		return nil, handleUnexpectedResponse(res)
	}

	result := &QueryResponse{}
	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
