package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/qlm-iot/core/routing"
	"net/http"
)

var destServer = flag.String("server", "", "Destination core server IP address")
var db routing.Datastore

func init() {
	db = routing.NewInMemoryStore()
}

func qlmPoller(w http.ResponseWriter, r *http.Request) {
	routing.QlmQuery(w, r, db)
}

func qlmHandler(w http.ResponseWriter, r *http.Request) {
	routing.QlmWsConnect(w, r, db)
}

func main() {
	flag.Parse()
	r := mux.NewRouter()
	s := r.PathPrefix("/qlm/Objects").Subrouter()
	s.HandleFunc("/{node}/", qlmPoller)
	s.HandleFunc("/", qlmPoller)
	r.HandleFunc("/qlmws", qlmHandler)
	http.Handle("/", r)
	fmt.Println("Waiting for connections on port 8000")
	http.ListenAndServe("localhost:8000", nil) // ignore err for now..
}
