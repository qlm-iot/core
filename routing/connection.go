package routing

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

type Connection struct {
	Recv chan []byte
	Send chan []byte
	conn *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func (ws *Connection) receiving(db Datastore) {
	for {
		mtype, msg, err := ws.conn.ReadMessage()
		if err != nil {
			fmt.Println(err.Error)
			ws.conn.Close()
			return
		}
		switch mtype {
		case websocket.BinaryMessage:
			Process(msg, db, ws)
		case websocket.CloseMessage:
			ws.conn.Close()
			return
		default:
		}

	}
}

func (ws *Connection) sending() {
	for {
		var msg []byte
		select {
		case msg = <-ws.Send:
			// Send back to the webSocket Channel?
			if err := ws.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				// Connection has failed, we should probably close it?
				fmt.Println(err.Error)
				ws.conn.Close()
				return
			}
		}
	}
}

func QlmWsConnect(w http.ResponseWriter, r *http.Request, db Datastore) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	from := make(chan []byte) // This channel is used for data from this client
	to := make(chan []byte)   // To send data to the client

	ws := &Connection{Recv: from, Send: to, conn: conn}

	go ws.sending()
	ws.receiving(db) // Block here

	defer func() {
		close(ws.Recv)
		close(ws.Send)
		ws.conn.Close()
	}()

}
