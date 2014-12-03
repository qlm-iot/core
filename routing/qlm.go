package routing

import (
	// "fmt"
	"sync"
	"time"
	// "github.com/qlm-iot/core"
	"github.com/qlm-iot/qlm/df"
	"github.com/qlm-iot/qlm/mi"
)

var mutex sync.Mutex
var subscriptions map[string]*Connection
var intervals map[string]chan struct{}

func init() {
	subscriptions = make(map[string]*Connection)
	intervals = make(map[string]chan struct{})
}

func payload(message *mi.Message) (*df.Objects, error) {
	return df.Unmarshal([]byte(message.Data))
}

func Process(msg []byte, db Datastore, c *Connection) {
	envelope, err := mi.Unmarshal(msg)
	if err == nil {
		if cancel := envelope.Cancel; cancel != nil {
			processCancel(cancel, db)
		} else if read := envelope.Read; read != nil {
			processRead(read, db, c)
		} else if write := envelope.Write; write != nil {
			processWrite(write, db, c)
		} else {
			msg, _ := createResponse("200")
			c.Send <- msg
		}
	}
}

func processCancel(c *mi.CancelRequest, db Datastore) {
	if len(c.RequestIds) > 0 {
		for _, rId := range c.RequestIds {
			// Notify the close channel and the datastore
			<-intervals[rId.Text]
			db.Cancel(rId.Text)
		}
	}
}

func processRead(r *mi.ReadRequest, db Datastore, c *Connection) {
	rc := make(chan Reply) // This is where the data will come from datastore

	// Support single read only for now..
	// Almost equal to write request, so refactor these..
	if r.Message != nil {
		rr, _ := payload(r.Message)
		for _, o := range rr.Objects {
			id := o.Id.Text
			mes := make([]string, len(o.InfoItems))
			for _, i := range o.InfoItems {
				mes = append(mes, i.Name)
			}
			req := &Request{Node: id, ReplyChan: rc, Measurements: mes}
			err, reply := db.Read(req) // err, reply -> requestId is the reply
			if err == nil {
				mutex.Lock()
				subscriptions[reply.RequestId] = c
				mutex.Unlock()
				if r.Interval > 0 {
					repeat(r.Interval, rc, c, reply.RequestId)
				} else {
					// Direct read?
				}
			}

		}
	} else {
		msg, _ := createResponse("200")
		c.Send <- msg
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

// Create replyPart here..? Call on read-request to empty the channel, if no callback is
// provided?
func clear(rc chan Reply) []df.Object {
	objects := make([]df.Object, 0, 5)
	for {
		select {
		case reply, open := <-rc:
			if open {
				id := &df.QLMID{Text: reply.Node}
				infoitems := to_infoitems(reply.Datapoints)
				object := df.Object{InfoItems: infoitems, Id: id}
				objects = append(objects, object)
			}
		default:
			return objects // Channel is empty
		}
	}
}

// Clean the replyChan on every interval and send the data to the websocket
// @TODO What if it's not a persistent connection? Next layer does the buffering?
func repeat(interval float64, rc chan Reply, c *Connection, requestId string) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	// add another ticker for ttl?
	quit := make(chan struct{})
	mutex.Lock()
	intervals[requestId] = quit
	mutex.Unlock()
	go func() {
		for {
			select {
			case <-ticker.C:
				// objects := clear(rc)
			case <-quit:
				ticker.Stop()
				mutex.Lock()
				delete(intervals, requestId)
				mutex.Unlock()
				return
			}
		}
	}()
}

func to_infoitems(datapoints []Data) []df.InfoItem {
	infoitems := make([]df.InfoItem, len(datapoints))
	for _, data := range datapoints {
		values := make([]df.Value, 0, 1)
		values = append(values, df.Value{UnixTime: data.Timestamp, Text: data.Value})
		infoitems = append(infoitems, df.InfoItem{Name: data.Measurement, Values: values})
	}
	return infoitems
}
