package main

import (
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
)

func init() {
	flag.StringVar(&graphFile, "graph", "", "graph file name")
	flag.StringVar(&deviceName, "deviceName", "Smart Mirror", "CEC Device Name")
}

func main() {
	flag.Parse()

	// s := Service{}

	// c, err := cec.Open("", deviceName)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	defer c.Destroy()
	// 	//c.PowerOn(0)
	// 	c.Standby(0)
	// }

	// d, _ := NewCECDisplay("", deviceName)
	ui := NewMirrorInterface("http://api.wunderground.com/api/52a3d65a04655627/forecast/q/MN/Minneapolis.json")

	images := make(chan ImageRequest)

	go func() {
		if err := ProcessImage(graphFile, images); err != nil {
			log.Printf("error from image processor, %v", err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if ind, err := ioutil.ReadFile("client/index.html"); err != nil {
			http.Error(w, err.Error(), 500)
		} else if tmpl, err := template.New("index").Parse(string(ind)); err != nil {
			http.Error(w, err.Error(), 500)
		} else {
			tmpl.Execute(w, struct {
				WebsocketURL string
			}{
				WebsocketURL: fmt.Sprintf("ws://%s/api/uisocket", r.Host),
			})
		}
	})

	http.Handle("/api/uisocket", newSocketHandler(ui))
	http.Handle("/api/", http.StripPrefix("/api/", &ServeInterface{ui}))

	log.Printf("serving on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))

}
