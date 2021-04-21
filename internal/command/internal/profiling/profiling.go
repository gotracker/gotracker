package profiling

import (
	"log"
	"net/http"
)

var (
	Allowed bool
)

func Activate(profilerBindAddress string) {
	if !Allowed {
		return
	}
	go func() {
		log.Println(http.ListenAndServe(profilerBindAddress, nil))
	}()
}
