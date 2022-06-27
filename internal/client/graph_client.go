package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/utils"
	"github.com/sirupsen/logrus"
)

// ErrTooManyRequests error representing too many requests to the API
var ErrTooManyRequests = fmt.Errorf("Too Many Requests")

// GraphClient is a client of the GraphKB API
type GraphClient struct {
	url           string
	authToken     string
	basicAuthUser string
	basicAuthPass string

	client *http.Client
}

// NewGraphClient create a client of the GraphKB API
func NewGraphClient(URL, authToken, basicAuthUser, basicAuthPass string, skipVerify bool) *GraphClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}

	return &GraphClient{
		url:           URL,
		authToken:     authToken,
		basicAuthUser: basicAuthUser,
		basicAuthPass: basicAuthPass,
		client:        client,
	}
}

func (gc *GraphClient) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", gc.url, path), body)
	if err != nil {
		return nil, err
	}
	if gc.authToken != "" {
		req.Header.Add(utils.XAuthTokenHeader, gc.authToken)
	}
	if gc.basicAuthUser != "" {
		req.SetBasicAuth(gc.basicAuthUser, gc.basicAuthPass)
	}
	return req, nil
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gc *GraphClient) ReadCurrentGraph() (*knowledge.Graph, error) {
	req, err := gc.newRequest("GET", "/api/graph/read", nil)
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
		return nil, handleUnexpectedResponse(res)
	}

	graphDecoder := knowledge.NewGraphDecoder(res.Body)
	graph := knowledge.NewGraph()
	err = graphDecoder.Decode(graph)
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

	req, err := gc.newRequest("PUT", "/api/graph/schema", bytes.NewBuffer(b))
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
		return handleUnexpectedResponse(res)
	}
	return nil
}

// InsertAssets send asset insert operations to the API
func (gc *GraphClient) InsertAssets(assets []knowledge.Asset) error {
	requestBody := PutGraphAssetRequestBody{}
	requestBody.Assets = assets

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	req, err := gc.newRequest("PUT", "/api/graph/assets", bytes.NewBuffer(b))
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
		return handleUnexpectedResponse(res)
	}
	return nil
}

// DeleteAssets send asset removal operations to the API
func (gc *GraphClient) DeleteAssets(assets []knowledge.Asset) error {
	requestBody := DeleteGraphAssetRequestBody{}
	requestBody.Assets = assets

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	req, err := gc.newRequest("DELETE", "/api/graph/assets", bytes.NewBuffer(b))
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
		return handleUnexpectedResponse(res)
	}
	return nil
}

// InsertRelations send relation insert operations to the API
func (gc *GraphClient) InsertRelations(relations []knowledge.Relation) error {
	requestBody := PutGraphRelationRequestBody{}
	requestBody.Relations = relations

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	req, err := gc.newRequest("PUT", "/api/graph/relations", bytes.NewBuffer(b))
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
		return handleUnexpectedResponse(res)
	}
	return nil
}

// DeleteRelations send relation removal operations to the API
func (gc *GraphClient) DeleteRelations(relations []knowledge.Relation) error {
	requestBody := DeleteGraphRelationRequestBody{}
	requestBody.Relations = relations

	b, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Unable to marshall request body")
	}

	req, err := gc.newRequest("DELETE", "/api/graph/relations", bytes.NewBuffer(b))
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
		return handleUnexpectedResponse(res)
	}
	return nil
}

func handleUnexpectedResponse(res *http.Response) error {
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("Unable to read error payload: %v", err)
	}
	bodyString := string(bodyBytes)
	return fmt.Errorf("Unexpected HTTP status %d with content %s: %s", res.StatusCode, res.Status, bodyString)
}
