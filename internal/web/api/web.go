package api

import (
	"github.com/gorilla/mux"
)

var (
	allowed bool
)

func Allowed() bool {
	return allowed
}

func ActivateRoute(router *mux.Router) {
	if !allowed {
		return
	}

	router.HandleFunc("/api/load", LoadHandler)
}
