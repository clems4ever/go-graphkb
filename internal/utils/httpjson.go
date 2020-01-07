package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"
)

var ErrStatusNotFound = errors.New("Not Found")

func HTTPGetJSON(url string, username string, password string, response interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(viper.GetString("sources.ldap.user"), viper.GetString("sources.ldap.password"))

	res, err := NewHTTPClient().Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == 404 {
		return ErrStatusNotFound
	}

	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(responseBytes, response)
}
