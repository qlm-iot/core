package routing

import (
	//	"github.com/qlm-iot/core/"
	"github.com/qlm-iot/qlm/df"
	"github.com/qlm-iot/qlm/mi"
)

func payload(message mi.Message) (*df.Objects, error) {
	return df.Unmarshal([]byte(message.Data))
}

func Process(msg []byte, db Datastore, ws wsConn) {
	envelope, err := mi.Unmarshal(msg)
	if err == nil {
		if cancel := &envelope.Cancel; cancel != nil {
			processCancel(cancel, db)
		}

		if read := &envelope.Read; read != nil {
		}

		if write := &envelope.Write; write != nil {
		}
	}
}

func processCancel(c *mi.CancelRequest, db Datastore) {
	if len(c.RequestIds) > 0 {
		for _, rId := range c.RequestIds {
			db.Cancel(rId.Text)
		}
	}
}

func processRead(r *mi.ReadRequest, db Datastore) {
	// Support single read only for now..
	// Almost equal to write request, so refactor these..
	rr, _ := payload(r.Message)
	for _, o := range wr.Objects {
		id := o.Id.Text
		for _, i := range o.InfoItems {
			data := &Data{Node: id, Measurement: i.Name}
			req := &Request{Reply: ws.reply, ToWrite: data}
			db.Read(req)
		}
	}
}

// Where's the reply channel..?

func processWrite(w *mi.WriteRequest, db Datastore, ws wsConn) {
	wr, _ := payload(w.Message)
	for _, o := range wr.Objects {
		id := o.Id.Text
		for _, i := range o.InfoItems {
			for _, v := range i.Values {
				data := &Data{Node: id, Measurement: i.Name, Value: v.Text}
				req := &Request{Reply: ws.reply, ToWrite: data}
				db.Write(req)
			}
		}
	}
}
