package knowledge

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/spf13/viper"
)

// GraphEmitter an emitter of full source graph
type GraphAPI struct {
	// GraphKB URL and auth token
	url       string
	authToken string

	client *http.Client
}

// NewGraphEmitter create an emitter of graph
func NewGraphAPI(url string, authToken string) *GraphAPI {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("graphkb.skip_verify")},
	}
	client := &http.Client{Transport: tr}

	return &GraphAPI{
		url:       url,
		authToken: authToken,
		client:    client,
	}
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gapi *GraphAPI) ReadCurrentGraph() (*Graph, error) {
	url := fmt.Sprintf("%s/api/graph/read?token=%s", gapi.url, gapi.authToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := gapi.client.Do(req)
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

	res, err := gapi.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Expected status code 200 and got %d", res.StatusCode)
	}
	return nil
}
