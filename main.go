package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"text/template"

	"github.com/gorilla/websocket"
)

var (
	graphFile  = ""
	deviceName = "Smart Mirror"
	upgrader   = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func init() {
	flag.StringVar(&graphFile, "graph", "", "graph file name")
	flag.StringVar(&deviceName, "deviceName", "Smart Mirror", "CEC Device Name")
}

type socketHandler struct {
	lock    *sync.Mutex
	sockets map[*websocket.Conn]*socket
	closed  chan *websocket.Conn
	server  *ServeInterface
}

func newSocketHandler(socketSrc interface{}) *socketHandler {
	ret := &socketHandler{
		lock:    &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*socket),
		closed:  make(chan *websocket.Conn),
		server:  &ServeInterface{In: socketSrc},
	}
	go ret.processClosed()
	return ret
}

func (handler *socketHandler) Write(msg []byte) {
	handler.lock.Lock()
	defer handler.lock.Unlock()
	for _, s := range handler.sockets {
		s.Out <- msg
	}
}

func (handler *socketHandler) processClosed() {
	for c := range handler.closed {
		handler.lock.Lock()
		delete(handler.sockets, c)
		handler.lock.Unlock()
	}
}

type socketRequest struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func (handler *socketHandler) processSocket(sock *socket) {
	for b := range sock.In {
		var dat socketRequest
		var ret interface{}
		var err error
		var msg []byte
		dd := make(map[string]interface{})

		if err = json.Unmarshal(b, &dat); err != nil {
			goto writeError
		}
		if ret, err = handler.server.Serve(dat.Path, dat.Value); err != nil {
			goto writeError
		}
		dd[dat.Path] = ret
		if msg, err = json.Marshal(dd); err != nil {
			goto writeError
		}
		goto writeMessage
	writeError:
		log.Println(err)
		msg, _ = json.Marshal(map[string]string{"error": err.Error()})
	writeMessage:
		sock.Out <- msg
	}
}

func (handler *socketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if conn, err := upgrader.Upgrade(w, r, nil); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
	} else {
		handler.lock.Lock()
		defer handler.lock.Unlock()

		sock := newSocket(conn, handler.closed)
		handler.sockets[conn] = sock
		go handler.processSocket(sock)
	}
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
