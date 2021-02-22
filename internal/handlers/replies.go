package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/sirupsen/logrus"
)

// ReplyWithSourceGraph sends a reply containing the schema graph
func ReplyWithSourceGraph(w http.ResponseWriter, sg *schema.SchemaGraph) {
	responseJSON, err := json.Marshal(sg)
	if err != nil {
		ReplyWithInternalError(w, err)
		return
	}

	if _, err := w.Write(responseJSON); err != nil {
		ReplyWithInternalError(w, err)
	}
}

// ReplyWithInternalError send response with internal error.
func ReplyWithInternalError(w http.ResponseWriter, err error) {
	logrus.Error(err)
	w.WriteHeader(http.StatusInternalServerError)
	_, werr := w.Write([]byte(err.Error()))
	if werr != nil {
		logrus.Error(werr)
	}
}

// ReplyWithBadRequest send response with bad request.
func ReplyWithBadRequest(w http.ResponseWriter, err error) {
	logrus.Error(err)
	w.WriteHeader(http.StatusBadRequest)
	_, werr := w.Write([]byte(err.Error()))
	if werr != nil {
		logrus.Error(werr)
	}
}

// ReplyWithUnauthorized send unauthorized response.
func ReplyWithUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	_, werr := w.Write([]byte("Unauthorized"))
	if werr != nil {
		logrus.Error(werr)
	}
}

// ReplyWithTooManyRequests send unauthorized response.
func ReplyWithTooManyRequests(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
	_, werr := w.Write([]byte("Too Many Requests. Retry later."))
	if werr != nil {
		logrus.Error(werr)
	}
}
