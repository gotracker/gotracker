package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

var (
	allowed bool
)

func Allowed() bool {
	return allowed
}

func ActivateRoute(router *mux.Router) error {
	if !allowed {
		return nil
	}

	router.HandleFunc("/api/play/{file}", PlayHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/play/{file}/{samplerate:[0-9]+}/{channels:[0-9]+}/{format}", PlayHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/playback/status/{file}", PlaybackStatusHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/upload", UploadHandler).Methods(http.MethodPost, http.MethodPut)
	return nil
}
