package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qlm-iot/qlm"
	"net/http"
)

var destServer = flag.String("server", "", "Destination core server IP address")
var messageChan = make(chan []byte)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func processMessages(mchan chan []byte) {
	var msg []byte
	for {
		select {
		case msg = <-mchan:
			answer, _ := qlm.Unmarshal(msg)
			for _, data := range answer.Objects {
				for _, info := range data.InfoItems {
					fmt.Println(info.Name)
				}
			}
			fmt.Println(answer.Version)
		}
	}
}

// We should have channel here.. This is single agent connecting
func qlmWsConnect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket connection failed")
		return
	}
	defer conn.Close()
	// Create channel for input messages
	//	mchan := make(chan []byte)
	for {
		mtype, msg, err := conn.ReadMessage()
		if err != nil {
			// Do something
		}
		switch mtype {
		case websocket.BinaryMessage:
			// This should be sent to the channel and then parsed with unmarshaller?
			// How to share channels between these two functions? Create one here and pass it on a function call? Only..?
			messageChan <- msg
		case websocket.TextMessage:
			/*
				Handle Close messages here also, so we know when to remove someone from subscriptions (maybe?) and/or from the
				list of available agents.
			*/
		default:
		}

	}
}

func init() {
}

func main() {
	flag.Parse()
	go processMessages(messageChan)
	http.HandleFunc("/qlmws", qlmWsConnect)
	http.ListenAndServe("localhost:8000", nil) // ignore err for now..
}
