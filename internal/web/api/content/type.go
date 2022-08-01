package content

import "net/http"

type Type interface {
	WriteHeader(w http.ResponseWriter)
	Write(w http.ResponseWriter, data []byte) (int, error)
}
