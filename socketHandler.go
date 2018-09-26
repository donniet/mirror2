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

func (handler *socketHandler) processSocket(sock *socket) {
	for b := range sock.In {
		res := socketResponse{
			Request: new(socketRequest),
		}
		var err error
		var msg []byte

		if err = json.Unmarshal(b, res.Request); err != nil {
			res.Request = nil
			res.Error = err.Error()
		} else if res.Response, err = handler.server.Serve(res.Request.Path, res.Request.Value); err != nil {
			res.Error = err.Error()
		}

		if msg, err = json.Marshal(res); err != nil {
			log.Printf("error marshalling response, %v", err)
			msg = []byte(`{"error":"error marshalling response"}`)
		}
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
