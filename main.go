package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qlm-iot/core/routing"
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

func (ws *wsConn) receiving() {
	for {
		mtype, msg, err := ws.conn.ReadMessage()
		if err != nil {
			fmt.Println(err.Error)
			return
		}
		switch mtype {
		case websocket.BinaryMessage:
			routing.Process(msg, db)
		case websocket.TextMessage:
			/*
				Handle Close messages here also, so we know when to remove someone from subscriptions (maybe?) and/or from the
				list of available agents.
			*/
		default:
		}

	}
}

func (ws *wsConn) sending() {
	/*
		for {
			var msg []byte
			select {
			case msg = <-ws.send:
				// Send back to the webSocket Channel?
			}
		}
	*/
}

func qlmWsConnect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	from := make(chan []byte) // This channel is used for data from this client
	to := make(chan []byte)   // To send data to the client

	ws := &wsConn{recv: from, send: to, conn: conn}

	go ws.sending()
	ws.receiving() // Block here

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
