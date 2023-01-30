package common

import (
	"encoding/json"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func RespondError(w http.ResponseWriter, message string, status int, logger log.Logger) {
	b, err := json.Marshal(&response{
		Error:  message,
		Status: "error",
	})
	if err != nil {
		level.Error(logger).Log("msg", "error marshaling json response", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if n, err := w.Write(b); err != nil {
		level.Error(logger).Log("msg", "error writing response", "bytesWritten", n, "err", err)
	}
}
