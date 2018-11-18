package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type API struct {
	mirror *mirrorInterface
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("path: '%s'", r.URL.Path)
	paths := strings.Split(r.URL.Path, "/")

	var body []byte

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var err error
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	log.Printf("body:\n%s", body)

	msg := (*json.RawMessage)(&body)
	if len(body) == 0 {
		msg = nil
	}

	o, err := api.mirror.ServeJSON(paths, msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
	} else if o == nil {
		w.WriteHeader(200)
	} else if _, err := w.Write(*o); err != nil {
		log.Printf("error writing response: %v", err)
		w.WriteHeader(500)
	}
}

func (api *API) ServeSocket(req []byte) (out []byte) {
	var d map[string]*json.RawMessage
	var err error

	if err = json.Unmarshal(req, &d); err != nil {
		out, _ = json.Marshal(&SocketError{Error: err.Error()})
		return
	}

	o := make(map[string]*json.RawMessage)

	for k, v := range d {
		paths := strings.Split(k, "/")
		r, err := api.mirror.ServeJSON(paths, v)
		if err != nil {
			out, _ = json.Marshal(&SocketError{Error: err.Error()})
			o[k] = (*json.RawMessage)(&out)
		} else {
			o[k] = r
		}
	}

	if out, err = json.Marshal(o); err != nil {
		out, _ = json.Marshal(&SocketError{Error: err.Error()})
	}
	return
}
