package knowledge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"io/ioutil"
	"net/http"
)

// GraphEmitter an emitter of full source graph
type GraphAPI struct {
	// GraphKB URL and auth token
	url       string
	authToken string
}

// NewGraphEmitter create an emitter of graph
func NewGraphAPI(url string, authToken string) *GraphAPI {
	return &GraphAPI{
		url:       url,
		authToken: authToken,
	}
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gapi *GraphAPI) ReadCurrentGraph() (*Graph, error) {
	url := fmt.Sprintf("%s/api/graph/read?token=%s", gapi.url, gapi.authToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Expected status code was 200 but got %d", res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	graph := NewGraph()
	err = json.Unmarshal(b, graph)
	if err != nil {
		return nil, err
	}
	return graph, nil
}

func (gapi *GraphAPI) UpdateGraph(sg schema.SchemaGraph, updates GraphUpdatesBulk) error {
	requestBody := GraphUpdateRequestBody{}
	requestBody.Updates = updates
	requestBody.Schema = sg

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	url := fmt.Sprintf("%s/api/graph/update?token=%s", gapi.url, gapi.authToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Expected status code 200 and got %d", res.StatusCode)
	}
	return nil
}
