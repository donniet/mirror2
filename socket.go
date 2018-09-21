package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type socket struct {
	closed  chan<- *websocket.Conn
	conn    *websocket.Conn
	In      chan []byte
	Out     chan []byte
	closing chan bool
	lock    *sync.Mutex
}

func newSocket(conn *websocket.Conn, closed chan<- *websocket.Conn) *socket {
	ret := &socket{
		conn:    conn,
		In:      make(chan []byte, 10),
		Out:     make(chan []byte, 10),
		closing: make(chan bool),
		closed:  closed,
		lock:    &sync.Mutex{},
	}
	go ret.reader()
	go ret.writer()
	return ret
}

func (s *socket) reader() {
	var msg []byte
	var err error
	for {
		// log.Printf("started reader")
		if _, msg, err = s.conn.ReadMessage(); err != nil {
			log.Print(err)
			break
		}
		// log.Printf("message: %#v %#v", msg, s.In)
		s.In <- msg
	}

	log.Printf("closing reader")

	s.Close()
}

func (s *socket) writer() {
	for done := false; !done; {
		select {
		case msg := <-s.Out:
			if err := s.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("websocket error, %#v, %s", msg, err.Error())
				done = true
			}
		case <-s.closing:
			done = true
		}
	}
	log.Printf("closing writer")

	s.Close()
}

func (s *socket) Close() {
	// log.Printf("closing socket")

	closing := func() (closing bool) {
		s.lock.Lock()
		defer s.lock.Unlock()

		select {
		case <-s.closing:
			closing = true
		default:
			closing = false
			close(s.closing) // this will stop the writer and prevent future closes
		}

		return
	}()

	if closing {
		return
	}
	// don't close s.Out in case there are outstanding references to it
	// close In to stop the websocket reader gofunc in the server
	close(s.In)    // this will end the socket listener
	s.conn.Close() // this will end the reader

	s.closed <- s.conn
	s.conn = nil
}
