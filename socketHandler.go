package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

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
