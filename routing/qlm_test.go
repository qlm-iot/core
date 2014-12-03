package routing

import (
	"testing"
)

type TestStore struct {
	cancels []string
}

func newTestStore() *TestStore {
	db := &TestStore{make([]string, 0)}
	return db
}

func (db *TestStore) Read(r *Request) error {
	return nil
}

func (db *TestStore) Write(r *Request) error {
	return nil
}

func (db *TestStore) Cancel(requestId string) error {
	_ = append(db.cancels, requestId)
	return nil
}

func TestQlmCancel(t *testing.T) {
	/*
		db := newTestStore()
		// Read with I/O from examples and put to Process?
		db.Cancel("REQ1234567")
		t.Fail()

	*/
}
