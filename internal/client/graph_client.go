package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
)

// ErrTooManyRequests error representing too many requests to the API
var ErrTooManyRequests = fmt.Errorf("Too Many Requests")

// GraphClient is a client of the GraphKB API
type GraphClient struct {
	url       string
	authToken string

	client *http.Client
}

// NewGraphClient create a client of the GraphKB API
func NewGraphClient(URL, authToken string, skipVerify bool) *GraphClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}

	return &GraphClient{
		url:       URL,
		authToken: authToken,
		client:    client,
	}
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gc *GraphClient) ReadCurrentGraph() (*knowledge.Graph, error) {
	url := fmt.Sprintf("%s/api/graph/read?token=%s", gc.url, gc.authToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := gc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Expected status code 200 and got %d: %s", res.StatusCode, res.Status)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	graph := knowledge.NewGraph()
	err = json.Unmarshal(b, graph)
	if err != nil {
		return nil, err
	}
	return graph, nil
}

// UpdateSchema send a graph schema update to the API
func (gc *GraphClient) UpdateSchema(sg schema.SchemaGraph) error {
	requestBody := PutGraphSchemaRequestBody{}
	requestBody.Schema = sg

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	url := fmt.Sprintf("%s/api/graph/schema?token=%s", gc.url, gc.authToken)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := gc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status code 200 and got %d: %s", res.StatusCode, res.Status)
	}
	return nil
}

// InsertAsset send an asset insert operation to the API
func (gc *GraphClient) InsertAsset(asset knowledge.Asset) error {
	requestBody := PutGraphAssetRequestBody{}
	requestBody.Asset = asset

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	url := fmt.Sprintf("%s/api/graph/asset?token=%s", gc.url, gc.authToken)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := gc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status code 200 and got %d: %s", res.StatusCode, res.Status)
	}
	return nil
}

// DeleteAsset send an asset removal operation to the API
func (gc *GraphClient) DeleteAsset(asset knowledge.Asset) error {
	requestBody := DeleteGraphAssetRequestBody{}
	requestBody.Asset = asset

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	url := fmt.Sprintf("%s/api/graph/asset?token=%s", gc.url, gc.authToken)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := gc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status code 200 and got %d: %s", res.StatusCode, res.Status)
	}
	return nil
}

// InsertRelation send a relation insert operation to the API
func (gc *GraphClient) InsertRelation(relation knowledge.Relation) error {
	requestBody := PutGraphRelationRequestBody{}
	requestBody.Relation = relation

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	url := fmt.Sprintf("%s/api/graph/relation?token=%s", gc.url, gc.authToken)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := gc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status code 200 and got %d: %s", res.StatusCode, res.Status)
	}
	return nil
}

// DeleteRelation send a relation upsert operation to the API
func (gc *GraphClient) DeleteRelation(relation knowledge.Relation) error {
	requestBody := DeleteGraphRelationRequestBody{}
	requestBody.Relation = relation

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	url := fmt.Sprintf("%s/api/graph/relation?token=%s", gc.url, gc.authToken)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := gc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status code 200 and got %d: %s", res.StatusCode, res.Status)
	}
	return nil
}
