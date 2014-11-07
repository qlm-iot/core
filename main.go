package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qlm-iot/core/routing"
	"github.com/qlm-iot/qlm"
	"net/http"
)

type wsConn struct {
	recv chan []byte
	send chan []byte
	conn *websocket.Conn
}

var destServer = flag.String("server", "", "Destination core server IP address")
var db routing.Datastore

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func (*ws wsConn) read() {
	for {
		mtype, msg, err := ws.conn.ReadMessage()
		if err != nil {
			// Do something
		}
		switch mtype {
		case websocket.BinaryMessage:
			ws.recv <- msg
		case websocket.TextMessage:
			/*
				Handle Close messages here also, so we know when to remove someone from subscriptions (maybe?) and/or from the
				list of available agents.
			*/
		default:
		}

	}
}

func (*ws wsConn) send() {
	for {
		var msg []byte
		select {
		case msg = <-ws.send:
			// Send back to the webSocket Channel?
		}
	}
}

func qlmWsConnect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket connection failed")
		return
	}

	from := make(chan []byte) // This channel is used for data from this client
	to := make(chan []byte)   // To send data to the client

	ws := &wsConn{recv: from, send: to, conn: conn}

	go ws.send()
	ws.read() // Block here

	defer func() {
		close(ws.recv)
		close(ws.send)
		ws.conn.Close()
	}()

}

func init() {
	db = routing.NewInMemoryStore()
}

func main() {
	flag.Parse()
	http.HandleFunc("/qlmws", qlmWsConnect)
	http.ListenAndServe("localhost:8000", nil) // ignore err for now..
}
