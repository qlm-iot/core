package routing

import (
	// "github.com/qlm-iot/core"
	"github.com/qlm-iot/qlm/df"
	"github.com/qlm-iot/qlm/mi"
)

func payload(message *mi.Message) (*df.Objects, error) {
	return df.Unmarshal([]byte(message.Data))
}

func Process(msg []byte, db Datastore, c *Connection) {
	envelope, err := mi.Unmarshal(msg)
	if err == nil {
		if cancel := envelope.Cancel; cancel != nil {
			processCancel(cancel, db)
		}

		if read := envelope.Read; read != nil {
			processRead(read, db, c)
		}

		if write := envelope.Write; write != nil {
			go processWrite(write, db, c)
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

func processRead(r *mi.ReadRequest, db Datastore, c *Connection) {

	rc := make(chan Reply) // This is where the data will go

	// Support single read only for now..
	// Almost equal to write request, so refactor these..
	rr, _ := payload(r.Message)
	for _, o := range rr.Objects {
		id := o.Id.Text
		mes := make([]string, len(o.InfoItems))
		for _, i := range o.InfoItems {
			mes = append(mes, i.Name)
		}
		req := &Request{Node: id, ReplyChan: rc, Measurements: mes}
		db.Read(req) // err, reply -> requestId is the reply
	}
}

func processWrite(w *mi.WriteRequest, db Datastore, c *Connection) {
	wr, _ := payload(w.Message)
	for _, o := range wr.Objects {
		id := o.Id.Text
		datapoints := make([]Data, 1)
		for _, i := range o.InfoItems {
			for _, v := range i.Values {
				data := Data{Measurement: i.Name, Value: v.Text}
				datapoints = append(datapoints, data)
			}
		}
		write := &Write{Node: id, Datapoints: datapoints}
		err, _ := db.Write(write)
		if err == nil {
			if response, err := createResponse("200"); err == nil {
				c.Send <- response
			}
		}
	}
}

func createResponse(code string) ([]byte, error) {
	envelope := mi.OmiEnvelope{
		Version: "1.0",
		Ttl:     0,
		Response: &mi.Response{
			Results: []mi.RequestResult{
				mi.RequestResult{
					Return: &mi.Return{ReturnCode: code},
				},
			},
		},
	}
	return mi.Marshal(envelope)
}

func repeat(interval int32, rc chan Reply) {
	// Repeat read.. subscription type of thingie..
	/*
		ticker := time.NewTicker(time.Duration(r.Interval) * time.Second)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
		return quit
	*/
}
