package routing

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

/*
  This is the InMemory store which will store only the latest available data.
  To store historical data also, change datastore to []Data and adjust some queries
*/
type Tracking struct {
	requestId string
	reply     chan Reply
	subs      []Key
	// What else? Needs to support both ways..
}

type Key struct {
	Node, Measurement string
}

type InMemoryStore struct {
	mu           sync.RWMutex
	trackMu      sync.Mutex
	subsMu       sync.RWMutex
	datastore    map[string]map[string]Data
	tracking     map[string]*Tracking
	subscription map[Key][]*Tracking
}

// Methods required by the Datastore interface
func NewInMemoryStore() *InMemoryStore {
	m := new(InMemoryStore)
	m.datastore = make(map[string]map[string]Data)
	m.tracking = make(map[string]*Tracking)
	m.subscription = make(map[Key][]*Tracking)
	return m
}

/*
  Equals subscription request.. missing immediate read
*/
func (m *InMemoryStore) Read(r *Request) (error, Reply) {
	rId := m.requestId()
	m.mu.RLock()
	defer m.mu.RUnlock()

	reply := Reply{RequestId: rId}

	if _, found := m.datastore[r.Node]; found {
		// Create requestId for this subscription request
		t := &Tracking{requestId: rId, reply: r.ReplyChan, subs: make([]Key, 10)}

		// Add subscriptions for each key
		m.subsMu.Lock()
		for _, me := range r.Measurements {
			k := Key{Node: r.Node, Measurement: me}
			t.subs = append(t.subs, k)
			m.subscription[k] = append(m.subscription[k], t)
		}
		m.subsMu.Unlock()

		// Add for tracking purposes
		m.trackMu.Lock()
		m.tracking[rId] = t
		m.trackMu.Unlock()
	} else {
		return errors.New("Could not fetch requested data, node does not exists"), reply
	}
	return nil, reply
}

func (m *InMemoryStore) Write(w *Write) (error, Reply) {
	datapoints := keymap(w.Datapoints)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check that we have node registered..
	node, found := m.datastore[w.Node]
	if !found {
		m.datastore[w.Node] = make(map[string]Data)
	}

	for _, datas := range node {
		if _, ok := datapoints[datas.Measurement]; !ok {
			delete(m.subscription, Key{Node: w.Node, Measurement: datas.Measurement})
			delete(m.datastore[w.Node], datas.Measurement)
		}
	}

	for _, data := range w.Datapoints {
		if data.Timestamp < 1 {
			// Add current timestamp if none was given
			data.Timestamp = time.Now().Unix()
		}
		node[data.Measurement] = data
		trackKey := Key{Node: w.Node, Measurement: data.Measurement}
		m.publish(trackKey, data.Value)
	}

	return nil, Reply{RequestId: m.requestId()}
}

func (m *InMemoryStore) Cancel(requestId string) error {
	m.trackMu.Lock()
	defer func() {
		m.trackMu.Unlock()
		m.subsMu.Unlock()
	}()
	m.subsMu.Lock()

	if _, found := m.tracking[requestId]; found {
		/*
			for _, k := range t.subs {
				for i, tt := range m.subscription[k] {
					if tt.requestId == requestId {
						d := append(tt[:i], tt[i+1:])
						m.subscription[k] = d
					}
				}
			}
		*/
		delete(m.tracking, requestId)
	} else {
		return errors.New("No subscription found for " + requestId)
	}

	return nil
}

/*
TODO:
If no call-back address is provided, then the responding QLM node SHOULD store the
data acquired for the subscription for later retrieval with a read message that contains the
corresponding requestID.

-> subscriptions should always create backing queue (chan), instead of our polling method only..

TODO2: Create and store requestIds.. they're not coming in the subscriptions. They must not come..
*/

// Internal methods

// @TODO Refactor this to some sort of util class to accept datastore & function to use
// This is general purpose for any datastore implementation

// This should be only for interval requests.. if there's listener for every active, we should
// block on chan read (which is filled on every write request, if there's a listener for it)
// @TODO This should be in the QLM layer.. not here!
/*
func (m *InMemoryStore) repeat(r *Request) chan struct{} {
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
}
*/
func (m *InMemoryStore) read(node string, measurement string, reply chan string) {
	/*
		m.mu.RLock()
		reply <- m.datastore[keyFormat(node, measurement)]
		m.mu.RUnlock()
	*/
}

var nextId int = 0

func (m *InMemoryStore) requestId() string {
	nextId++
	return fmt.Sprintf("REQ%07d", nextId)
}

func (m *InMemoryStore) publish(key Key, value string) error {
	m.subsMu.RLock()
	defer m.subsMu.RUnlock()

	if ts, found := m.subscription[key]; found {
		for _, t := range ts {
			d := Data{Measurement: key.Measurement, Value: value}
			data := []Data{d}
			r := Reply{RequestId: t.requestId, Node: key.Node, Datapoints: data}
			t.reply <- r
		}
	}
	return nil
}

func keymap(dataslice []Data) map[string]struct{} {
	var datapoints map[string]struct{} = make(map[string]struct{})
	for _, d := range dataslice {
		datapoints[d.Measurement] = struct{}{}
	}
	return datapoints
}
