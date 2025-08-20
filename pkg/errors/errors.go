package errors

import (
	"encoding/json"
	"net/http"

	"csjk-bk/models"
)

// ServeError writes an HTTP 500 response in StandardResponse JSON format.
func ServeError(rw http.ResponseWriter, _ *http.Request, err error) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(rw).Encode(&models.StandardResponse{Detail: err.Error()})
}
