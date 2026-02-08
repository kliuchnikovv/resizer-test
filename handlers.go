package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func (s *service) resizeHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Expecting POST request"))
			return
		}

		queryValues, err := url.Parse(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to parse URL"))
			return
		}

		var isAsync bool
		if queryValues.Query().Get("async") == "true" {
			isAsync = true
		}

		request := resizeRequest{}
		if err := json.NewDecoder(io.LimitReader(r.Body, 8*1024)).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to parse request"))
			return
		}

		results, err := s.processResizes(request, isAsync)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to process request"))
			return
		}

		data, err := json.Marshal(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal response"))
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(data)
	})
}

func (s *service) getImageHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := s.processRetrieve(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Add("content-type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
}
