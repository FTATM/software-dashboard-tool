package handler

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func respondJson(w http.ResponseWriter, code int, response *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// Assuming model.Response has a JSON tag like `json:"message"`
	json.NewEncoder(w).Encode(response)
}
