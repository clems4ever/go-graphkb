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

	if res.StatusCode == 401 {
		return nil, fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode != 200 {
		return nil, fmt.Errorf("Expected status code 200 and got %d", res.StatusCode)
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

// UpdateGraph send a graph update to the API
func (gc *GraphClient) UpdateGraph(sg schema.SchemaGraph, updates knowledge.GraphUpdatesBulk) error {
	requestBody := GraphUpdateRequestBody{}
	requestBody.Updates = &updates
	requestBody.Schema = sg

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	url := fmt.Sprintf("%s/api/graph/update?token=%s", gc.url, gc.authToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := gc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		return fmt.Errorf("Unauthorized access. Check your auth token")
	} else if res.StatusCode != 200 {
		return fmt.Errorf("Expected status code 200 and got %d", res.StatusCode)
	}
	return nil
}
