package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gotracker/gotracker/internal/web/api/files"
)

type apiUploadResponse struct {
	Statuses []apiUploadStatus `json:"statuses"`
}

type apiUploadStatus struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		log.Println("/api/upload: body was nil")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	reader, err := r.MultipartReader()
	if err != nil {
		log.Printf("/api/upload: multipartreader failed with %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var response apiUploadResponse

	var anySuccess bool
	f, err := reader.NextPart()
	for f != nil && err == nil {
		defer f.Close()
		var data []byte
		data, err = ioutil.ReadAll(f)
		if err != nil {
			log.Printf("/api/upload: readall failed with %v\n", err)
			response.Statuses = append(response.Statuses, apiUploadStatus{
				Filename: f.FileName(),
				Status:   "error",
				Error:    err.Error(),
			})
			continue
		}

		if err := files.AddFile(f.FileName(), data); err != nil {
			log.Printf("/api/upload: addfile failed with %v\n", err)
			response.Statuses = append(response.Statuses, apiUploadStatus{
				Filename: f.FileName(),
				Status:   "error",
				Error:    err.Error(),
			})
			continue
		}

		log.Printf("/api/upload: added %q successfully.\n", f.FileName())
		anySuccess = true

		response.Statuses = append(response.Statuses, apiUploadStatus{
			Filename: f.FileName(),
			Status:   "success",
		})

		f, err = reader.NextPart()
	}

	if !anySuccess {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
