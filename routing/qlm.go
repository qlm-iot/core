package routing

import (
	"github.com/qlm-iot/qlm/df"
	"github.com/qlm-iot/qlm/mi"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Waiter struct {
	conn *Connection
	temp chan []byte
	rc   chan Reply
}

var mutex sync.Mutex
var subscriptions map[string]Waiter
var intervals map[string]chan struct{}

func init() {
	subscriptions = make(map[string]Waiter)
	intervals = make(map[string]chan struct{})
}

func payload(message *mi.Message) (*df.Objects, error) {
	return df.Unmarshal([]byte(message.Data))
}

/*
 * Entry point for QLM messages, depending on the type of the message
 * (cancel/read/write), call the approriate method. This method is
 * blocking.
 */
func Process(msg []byte, db Datastore, c *Connection) {
	if envelope, err := mi.Unmarshal(msg); err == nil {
		if cancel := envelope.Cancel; cancel != nil {
			processCancel(cancel, db, c)
		} else if read := envelope.Read; read != nil {
			processRead(read, db, c)
		} else if write := envelope.Write; write != nil {
			processWrite(write, db, c)
		} else {
			msg, _ := createResponse(http.StatusOK)
			c.Send <- msg
		}
	} else {
		msg, _ := createErrorResponse(http.StatusBadRequest, "Request could not be parsed")
		c.Send <- msg
	}
}

func processCancel(c *mi.CancelRequest, db Datastore, conn *Connection) {
	if len(c.RequestIds) > 0 {
		for _, rId := range c.RequestIds {
			// Notify the close channel and the datastore
			// <-intervals[rId.Text]
			db.Cancel(rId.Text)
			if _, found := subscriptions[rId.Text]; found {
				mutex.Lock()
				delete(subscriptions, rId.Text)
				mutex.Unlock()
			} else {
				msg, _ := createErrorResponse(http.StatusNotFound, "Subscription not found")
				conn.Send <- msg
				return
			}
		}
	}
	msg, _ := createResponse(http.StatusOK)
	conn.Send <- msg
}

func processRead(r *mi.ReadRequest, db Datastore, c *Connection) {
	rc := make(chan Reply, 8192) // This is where the data will come from datastore

	if r.Message != nil {
		rr, _ := payload(r.Message)
		for _, o := range rr.Objects {
			id := o.Id.Text
			mes := make([]string, 0, len(o.InfoItems))
			for _, i := range o.InfoItems {
				mes = append(mes, i.Name)
			}
			req := &Request{Node: id, ReplyChan: rc, Measurements: mes}
			err, reply := db.Subscribe(req) // err, reply -> requestId is the reply
			if err == nil {
				mutex.Lock()
				t := make(chan []byte)
				w := Waiter{conn: c, temp: t, rc: rc}
				subscriptions[reply.RequestId] = w
				mutex.Unlock()
				if r.Interval > 0 {
					go repeat(r.Interval, rc, t, reply.RequestId)
					req, _ := createReqReply(http.StatusOK, reply.RequestId)
					c.Send <- req
				} else {
					if err, _ := db.ReadImmediate(req); err != nil {
						msg, _ := createErrorResponse(http.StatusNotFound, err.Error())
						c.Send <- msg
					} else {
						msg, _ := createMsg(clear(rc))
						envelope, _ := createMessageResponse(http.StatusOK, msg)
						c.Send <- envelope
					}
				}
			} else {
				msg, _ := createErrorResponse(http.StatusNotFound, err.Error())
				c.Send <- msg
			}

		}
	} else if len(r.RequestIds) > 0 {
		rId := r.RequestIds[0].Text
		w, found := subscriptions[rId]
		if found {
			msg, _ := createMsg(clear(w.rc))
			envelope, _ := createMessageResponse(http.StatusOK, msg)
			c.Send <- envelope
		} else {
			msg, _ := createErrorResponse(http.StatusNotFound, "Subscription not found")
			c.Send <- msg
		}
	} else {
		msg, _ := createResponse(http.StatusOK)
		c.Send <- msg
	}
}

func processWrite(w *mi.WriteRequest, db Datastore, c *Connection) {
	wr, _ := payload(w.Message)
	for _, o := range wr.Objects {
		id := o.Id.Text
		datapoints := make([]Data, 0, 1)
		for _, i := range o.InfoItems {
			for _, v := range i.Values {
				data := Data{Measurement: i.Name, Value: v.Text}
				datapoints = append(datapoints, data)
			}
		}
		write := &Write{Node: id, Datapoints: datapoints}
		err, _ := db.Write(write)
		if err == nil {
			if response, err := createResponse(http.StatusOK); err == nil {
				c.Send <- response
			}
		}
	}
}

/*
 * Assistant methods for creating QLM response messages
 */

func createResponseTemplate(code int) mi.OmiEnvelope {
	return mi.OmiEnvelope{
		Version: "1.0",
		Ttl:     0,
		Response: &mi.Response{
			Results: []mi.RequestResult{
				mi.RequestResult{
					Return: &mi.Return{ReturnCode: strconv.Itoa(code)},
				},
			},
		},
	}
}

func createErrorResponse(code int, desc string) ([]byte, error) {
	envelope := createResponseTemplate(code)
	ret := envelope.Response.Results[0].Return
	ret.Description = desc
	return mi.Marshal(envelope)
}

func createMessageResponse(code int, objects []byte) ([]byte, error) {
	envelope := createResponseTemplate(code)
	envelope.Response.Results[0].Message = &mi.Message{Data: string(objects)}
	return mi.Marshal(envelope)
}

func createResponse(code int) ([]byte, error) {
	envelope := createResponseTemplate(code)
	return mi.Marshal(envelope)
}

func createReqReply(code int, requestId string) ([]byte, error) {
	envelope := createResponseTemplate(code)
	envelope.Response.Results[0].RequestId = &mi.Id{Text: requestId}
	return mi.Marshal(envelope)
}

func createMsg(objects []df.Object) ([]byte, error) {
	return df.Marshal(df.Objects{Objects: objects})
}

// Clear the reply channel and forge a message to be sent back
func clear(rc chan Reply) []df.Object {
	objects := make([]df.Object, 0, 5)
Clear:
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
			break Clear
		}
	}
	return objects
}

func repeat(interval float64, rc chan Reply, t chan []byte, requestId string) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	// NewTimer for TTL support..
	quit := make(chan struct{})
	mutex.Lock()
	intervals[requestId] = quit
	mutex.Unlock()
	go func() {
		for {
			select {
			case <-ticker.C:
				msg, _ := createMsg(clear(rc))
				t <- msg
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
	infoitems := make([]df.InfoItem, 0, len(datapoints))
	for _, data := range datapoints {
		values := make([]df.Value, 0, 1)
		values = append(values, df.Value{UnixTime: data.Timestamp, Text: data.Value})
		infoitems = append(infoitems, df.InfoItem{Name: data.Measurement, Values: values})
	}
	return infoitems
}

/*
 * Following methods are intended for querying possible datasources from a QLM node
 */

func NodeList(db Datastore) ([]byte, error) {
	nodes := db.NodeList()
	objects := make([]df.Object, 0, len(nodes))
	for _, k := range nodes {
		id := &df.QLMID{Text: k}
		object := df.Object{Id: id}
		objects = append(objects, object)
	}
	msg, _ := createMsg(objects)
	return createMessageResponse(http.StatusOK, msg)
}

func KeyList(node string, db Datastore) ([]byte, error) {
	err, keys := db.SourceList(node)
	if err != nil {
		return nil, err
	}
	infoitems := make([]df.InfoItem, 0, len(keys))
	for _, k := range keys {
		infoitems = append(infoitems, df.InfoItem{Name: k})
	}
	object := df.Object{Id: &df.QLMID{Text: node}, InfoItems: infoitems}
	objects := make([]df.Object, 0, 1)
	objects = append(objects, object)
	msg, _ := createMsg(objects)
	return createMessageResponse(http.StatusOK, msg)
}
