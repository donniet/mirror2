package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

var (
	graphFile  = ""
	deviceName = "Smart Mirror"
	videoFifo  = ""
	motionFifo = ""
)

func init() {
	flag.StringVar(&graphFile, "graph", "", "graph file name")
	flag.StringVar(&deviceName, "deviceName", "Smart Mirror", "CEC Device Name")
	flag.StringVar(&videoFifo, "videoFifo", "", "path to the video fifo")
	flag.StringVar(&motionFifo, "motionFifo", "", "path to the motion vectors fifo")
}

func main() {
	flag.Parse()

	changed := make(chan socketResponse)

	ui := NewMirrorInterface(
		"http://api.wunderground.com/api/52a3d65a04655627/forecast/q/MN/Minneapolis.json",
		changed)

	socketHandler := newSocketHandler(ui)

	go func() {
		for obj := range changed {
			if b, err := json.Marshal(obj); err != nil {
				log.Printf("error marshalling changed message from %v", ui)
			} else {
				socketHandler.Write(b)
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if ind, err := ioutil.ReadFile("client/index.html"); err != nil {
			http.Error(w, err.Error(), 500)
		} else if tmpl, err := template.New("index").Delims("[[", "]]").Parse(string(ind)); err != nil {
			http.Error(w, err.Error(), 500)
		} else {
			tmpl.Execute(w, struct {
				WebsocketURL string
			}{
				WebsocketURL: fmt.Sprintf("ws://%s/api/uisocket", r.Host),
			})
		}
	})

	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("client/"))))
	http.Handle("/api/uisocket", socketHandler)
	http.Handle("/api/", http.StripPrefix("/api/", &ServeInterface{ui}))

	log.Printf("serving on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))

}
