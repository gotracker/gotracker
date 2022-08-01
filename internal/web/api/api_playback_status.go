package api

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gotracker/gotracker/internal/web/api/broker"
)

var (
	playbackStatusServerMap sync.Map // map[string]playbackStatusServerSetup
)

type playbackStatusServerSetup struct {
	once   sync.Once
	broker *broker.Broker
}

func playbackStatusGetBroker(filename string) *broker.Broker {
	value, _ := playbackStatusServerMap.LoadOrStore(filename, &playbackStatusServerSetup{})
	srv := value.(*playbackStatusServerSetup)
	srv.once.Do(func() {
		srv.broker = broker.NewServer(fmt.Sprintf("/api/playback/status/%s", filename))
	})
	return srv.broker
}

func PlaybackStatusHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	name, ok := vars["file"]
	if !ok || name == "" {
		log.Println("/api/playback/status: file was not provided in vars")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	playbackStatusGetBroker(name).Stream(w, r)
}
