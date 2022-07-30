package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gotracker/playback/format"
)

func LoadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		log.Println("/api/load: Body was nil")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("/api/load: ReadAll failed with %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	reader := bytes.NewReader(data)

	s, err := format.LoadFromReader("", reader)
	if err != nil {
		log.Printf("/api/load: LoadFromReader failed with %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(s); err != nil {
		log.Printf("/api/load: Encode failed with %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
