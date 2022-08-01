package api

import (
	"encoding/json"
	"net/http"
)

type apiReadyResponse struct {
	Ready bool `json:"ready"`
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	response := apiReadyResponse{
		Ready: true,
	}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
