package routing

import (
	"github.com/gorilla/mux"
	"net/http"
)

func QlmQuery(w http.ResponseWriter, r *http.Request, db Datastore) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/xml")
	if node, found := vars["node"]; found {
		// Request node details
		resp, _ := KeyList(node, db)
		w.Write(resp)
	} else {
		// Request list of available nodes
		resp, _ := NodeList(db)
		w.Write(resp)
	}
}
