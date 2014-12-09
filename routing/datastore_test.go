package routing

import (
	"testing"
)

func TestReadWrite(t *testing.T) {
	/*
		s := NewInMemoryStore()

		c := make(chan string)
		d := Data{"node", "measurement", "value"}
		r := &Request{"1", c, 0, d}
		err := s.Write(r)
		if err != nil {
			t.Fail()
		}

		go s.Read(r)

		value := <-c
		if len(value) != 5 {
			t.Fail()
		}
	*/
}

func TestCancel(t *testing.T) {
	//	t.Fail()
}
