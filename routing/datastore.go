package routing

import (
	"errors"
	"sync"
	"time"
)

/*

  This is the InMemory store, that will store only the latest available data on a Map.

*/

type InMemoryStore struct {
	mu        sync.RWMutex
	trackMu   sync.Mutex
	datastore map[string]string
	tracking  map[string]chan struct{}
}

// Methods required by the Datastore interface
func NewInMemoryStore() *InMemoryStore {
	m := new(InMemoryStore)
	m.datastore = make(map[string]string)
	m.tracking = make(map[string]chan struct{})
	return m
}

func (m *InMemoryStore) Read(r *Request) error {
	if _, found := m.datastore[m.key(r)]; found {
		if r.Interval > 0 {
			reply := m.repeat(r)
			m.trackMu.Lock()
			m.tracking[r.RequestId] = reply // Store the quit channel
			m.trackMu.Unlock()
		} else {
			m.read(r.ToWrite.Node, r.ToWrite.Measurement, r.Reply)
		}
	} else {
		return errors.New("Could not fetch requested data")
	}
	return nil
}

func (m *InMemoryStore) Write(r *Request) error {
	key := m.key(r)

	m.mu.Lock()
	m.datastore[key] = r.ToWrite.Value
	m.mu.Unlock()
	return nil
}

func (m *InMemoryStore) Cancel(requestId string) error {
	m.trackMu.Lock()
	defer m.trackMu.Unlock()

	if _, found := m.tracking[requestId]; found {
		close(m.tracking[requestId])
		delete(m.tracking, requestId)
	} else {
		errors.New("No subscription found for " + requestId)
	}
	return nil
}

// Internal methods

// @TODO Refactor this to some sort of util class to accept datastore & function to use
// This is general purpose for any datastore implementation
func (m *InMemoryStore) repeat(r *Request) chan struct{} {
	ticker := time.NewTicker(time.Duration(r.Interval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				m.read(r.ToWrite.Node, r.ToWrite.Measurement, r.Reply)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}

func (m *InMemoryStore) read(node string, measurement string, reply chan string) {
	m.mu.RLock()
	reply <- m.datastore[keyFormat(node, measurement)]
	m.mu.RUnlock()
}

func keyFormat(node string, measurement string) string {
	return node + ":" + measurement
}

func (m *InMemoryStore) key(r *Request) string {
	return keyFormat(r.ToWrite.Node, r.ToWrite.Measurement)
}
