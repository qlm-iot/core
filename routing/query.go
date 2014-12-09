package routing

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func QlmQuery(w http.ResponseWriter, r *http.Request, db Datastore) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/xml")
	var resp []byte
	var err error
	if node, found := vars["node"]; found {
		// Request node details
		resp, err = KeyList(node, db)
	} else {
		// Request list of available nodes
		resp, err = NodeList(db)
	}
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not fetch requested info %s", err)
		return
	}
	w.Write(resp)
}

func QlmInterface(w http.ResponseWriter, r *http.Request, db Datastore) {
	req := r.FormValue("msg")
	wait := make(chan []byte, 5)
	c := &Connection{Send: wait}

	if req != "" {
		Process([]byte(req), db, c)
	}
	msg := <-wait
	w.Write(msg)
}
