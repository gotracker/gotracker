package web

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gotracker/gotracker/internal/profiling"
	webApi "github.com/gotracker/gotracker/internal/web/api"
)

var (
	allowed bool
	Enabled bool
	webCtx  context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
)

func Allowed() bool {
	return allowed
}

func Shutdown() {
	if cancel != nil {
		cancel()
	}
}

func WaitForShutdown() {
	wg.Wait()
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	Shutdown()
}

type RouteActivator func(router *mux.Router) error

func UnhandledHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("web server: unhandled route %v\n", r.URL.Path)
}

func Activate(ctx context.Context, webBindAddress string, additionalRoutes ...RouteActivator) {
	if !allowed || !Enabled {
		return
	}

	webCtx, cancel = context.WithCancel(ctx)

	router := mux.NewRouter()

	// activate profiling (if enabled)
	// activate webApi (if allowed)
	additionalRoutes = append([]RouteActivator{profiling.ActivateRoute, webApi.ActivateRoute}, additionalRoutes...)

	for _, additionalRoute := range additionalRoutes {
		additionalRoute(router)
	}

	// add shutdown handler
	router.HandleFunc("/shutdown", ShutdownHandler)

	router.PathPrefix("/").HandlerFunc(UnhandledHandler)

	srv := &http.Server{
		Handler: router,
		Addr:    webBindAddress,
		BaseContext: func(l net.Listener) context.Context {
			return webCtx
		},
		WriteTimeout: 0,
		ReadTimeout:  0,
		IdleTimeout:  15 * time.Second,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("web server listening on %s...\n", webBindAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("web server: %v\n", err)
		}
		log.Printf("web server closed.\n")
	}()

	go func() {
		<-webCtx.Done()
		if err := srv.Shutdown(webCtx); err != nil && !errors.Is(err, context.Canceled) {
			// failure/timeout shutting down the server gracefully
			log.Panic(err)
		}
	}()
}
