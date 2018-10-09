package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/donniet/mvnc"
)

var (
	graphFile                  = ""
	deviceName                 = "Smart Mirror"
	videoFifo                  = "-"
	motionFifo                 = ""
	mbx                        = 120
	mby                        = 68
	magnitude                  = 60
	totalMotion                = 10
	detectionThreshold float64 = 0.75
	addr               string  = ":8080"
)

func init() {
	flag.StringVar(&graphFile, "graph", graphFile, "graph file name")
	flag.StringVar(&deviceName, "deviceName", deviceName, "CEC Device Name")
	flag.StringVar(&videoFifo, "video", videoFifo, "path to the video fifo")
	flag.StringVar(&motionFifo, "motion", motionFifo, "path to the motion vectors fifo")
	flag.Float64Var(&detectionThreshold, "detectionThreshold", detectionThreshold, "threshold to constitute detection")
	flag.IntVar(&mbx, "mbx", mbx, "motion vector X")
	flag.IntVar(&mby, "mby", mby, "motion vector Y")
	flag.IntVar(&magnitude, "magnitude", magnitude, "magnitude of motion vector")
	flag.IntVar(&totalMotion, "totalMotion", totalMotion, "total motion vectors to trigger screen")
	flag.StringVar(&addr, "addr", addr, "address to host")
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Parse()

	log.Printf("starting")

	var vid, mot *os.File
	var err error

	if videoFifo == "-" {
		vid = os.Stdin
	} else if vid, err = os.OpenFile(videoFifo, os.O_RDONLY, 0600); err != nil {
		log.Fatal(err)
	}

	changed := make(chan socketResponse)

	ui := NewMirrorInterface(
		"http://api.wunderground.com/api/52a3d65a04655627/forecast/q/MN/Minneapolis.json",
		changed)

	if motionFifo == "" {
		log.Printf("disabling motion detection")
	} else {
		log.Printf("opening motion fifo")
		if mot, err = os.OpenFile(motionFifo, os.O_RDONLY, 0600); err != nil {
			log.Fatal(err)
		}
		log.Printf("motion processor")
		motionDetected := MotionProcessor{
			mbx:       mbx,
			mby:       mby,
			magnitude: magnitude,
			total:     totalMotion,
			throttle:  500 * time.Millisecond,
		}.Process(mot)

		log.Printf("starting motion detector")
		go func() {
			for t := range motionDetected {
				log.Printf("motion detected at %v", t)
				ui.Display().Wake("10m")
			}
		}()
	}

	personDetected := mvnc.Graph{
		GraphFile: graphFile,
		Names:     map[int]string{0: "donnie", 1: "lauren"},
		Threshold: float32(detectionThreshold),
		Throttle:  200 * time.Millisecond,
	}.Process(vid)

	log.Printf("starting person detector")
	go func() {
		for p := range personDetected {
			log.Printf("person detected: %s", p)
		}
		log.Printf("person detector ended, exiting")
		os.Exit(-1)
	}()

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

	log.Printf("serving on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))

}
