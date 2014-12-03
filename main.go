package main

import (
	"flag"
	"fmt"
	"github.com/qlm-iot/core/routing"
	"net/http"
)

var destServer = flag.String("server", "", "Destination core server IP address")
var db routing.Datastore

func init() {
	db = routing.NewInMemoryStore()
}

func qlmHandler(w http.ResponseWriter, r *http.Request) {
	routing.QlmWsConnect(w, r, db)
}

func main() {
	flag.Parse()
	http.HandleFunc("/qlmws", qlmHandler)
	fmt.Println("Waiting for connections on port 8000")
	http.ListenAndServe("localhost:8000", nil) // ignore err for now..
}
