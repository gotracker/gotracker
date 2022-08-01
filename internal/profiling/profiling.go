package profiling

import (
	"net/http"

	"github.com/gorilla/mux"
)

var (
	allowed bool
	Enabled bool
)

func Allowed() bool {
	return allowed
}

func ActivateRoute(router *mux.Router) error {
	if !allowed || !Enabled {
		return nil
	}
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	return nil
}
