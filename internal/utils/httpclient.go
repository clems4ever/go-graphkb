package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// HTTPClient is an http client perofrming caching
type HTTPClient struct {
	*http.Client

	CacheDir string
}

// NewHTTPClient create an http client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		Client:   http.DefaultClient,
		CacheDir: viper.GetString("http_cache_dir"),
	}
}

// BytesBuffer represent bytes buffer as ReadCloser
type BytesBuffer struct {
	reader *bytes.Reader
}

// NewBytesBuffer create a bytes buffer
func NewBytesBuffer(p []byte) *BytesBuffer {
	return &BytesBuffer{reader: bytes.NewReader(p)}
}

// Read from the byte buffer
func (bb *BytesBuffer) Read(p []byte) (n int, err error) {
	return bb.reader.Read(p)
}

// Close the bytes buffer
func (bb *BytesBuffer) Close() error {
	return nil
}

func (hc *HTTPClient) doFromCache(req *http.Request) (*http.Response, error) {
	if _, err := os.Stat(hc.CacheDir); os.IsNotExist(err) {
		err := os.MkdirAll(hc.CacheDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	url := req.URL.String()
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, ":", "_")
	cacheFile := path.Join(hc.CacheDir, url)
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		logrus.Debugf("Caching HTTP response of %s\n", req.URL.String())
		res, err := hc.Client.Do(req)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == 404 {
			return nil, ErrStatusNotFound
		}

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		f, err := os.Create(cacheFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		// Write content of body into file
		_, err = f.Write(b)
		if err != nil {
			return nil, err
		}
		res.Body = NewBytesBuffer(b)
		return res, nil
	}

	logrus.Debugf("Using HTTP cache for %s\n", req.URL.String())
	f, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	res := &http.Response{}
	res.Body = NewBytesBuffer(b)
	res.StatusCode = 200

	return res, nil
}

// Do perform the http request
func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	if hc.CacheDir == "" {
		return hc.Client.Do(req)
	}
	return hc.doFromCache(req)
}
