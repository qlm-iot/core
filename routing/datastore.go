package routing

import (
	"errors"
	"sync"
	"time"
)

type Datastore interface {
	Read(r *Request)
	Write(r *Request)
	Cancel(rId string)
}

var mu = &sync.Mutex{}      // Datastore Mutex
var trackMu = &sync.Mutex{} // Tracking map Mutex
var datastore = make(map[string]string)
var tracking = make(map[string]chan struct{})

type Data struct {
	Node        string
	Measurement string
	Value       string // Or something else.. do we even need this struct?
}

type Request struct {
	RequestId string
	Reply     chan string // Pointer?
	Interval  int32
	ToWrite   Data
}

func repeat(r *Request) chan struct{} {
	ticker := time.NewTicker(time.Duration(r.Interval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				read(r.ToWrite.Node, r.ToWrite.Measurement, r.Reply)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}

// Do we need to synchronize with channels the access to this routing store?
// Yes, but later.. either with mutex or with just single channel ..

func read(node string, measurement string, reply chan string) {
	mu.Lock()
	reply <- datastore[keyFormat(node, measurement)]
	mu.Unlock()
}

func keyFormat(node string, measurement string) string {
	return node + ":" + measurement
}

func key(r *Request) string {
	return keyFormat(r.ToWrite.Node, r.ToWrite.Measurement)
}

func Read(r *Request) error {
	// At the end / defer, create new go func() { sleep; fetch } which listens for cancel?
	if _, found := datastore[key(r)]; found {
		// Value was found, return Data
		if r.Interval > 0 {
			reply := repeat(r)
			trackMu.Lock()
			tracking[r.RequestId] = reply // Store the quit channel
			trackMu.Unlock()
		} else {
			read(r.ToWrite.Node, r.ToWrite.Measurement, r.Reply)
		}
	} else {
		// Value was not found, return error
		return errors.New("Could not fetch requested data")
	}
	return nil
}

func Write(r *Request) error {
	key := key(r)

	mu.Lock()
	datastore[key] = r.ToWrite.Value
	mu.Unlock()
	return nil
}

func Cancel(requestId string) error {
	if _, found := tracking[requestId]; found {
		// it was a valid key
		trackMu.Lock()
		close(tracking[requestId])
		delete(tracking, requestId)
		trackMu.Unlock()
	} else {
		errors.New("No subscription found for " + requestId)
	}
	return nil
}
