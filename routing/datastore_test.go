package routing

import (
	"testing"
)

func TestReadWrite(t *testing.T) {
	c := make(chan string)
	d := Data{"node", "measurement", "value"}
	r := &Request{"1", c, 0, d}
	err := Write(r)
	if err != nil {
		t.Fail()
	}

	go Read(r)

	value := <-c
	if len(value) != 5 {
		t.Fail()
	}
}

func TestRead(t *testing.T) {

}

func TestCancel(t *testing.T) {
	t.Fail()
}

// Wait a moment..

/*
We need:
 - Function that reads data and passes it to a chan
 - Fetch that will repeat the above function
 - Cancel that would remove the subscription.. so:
   - We need to map those functions (well, their quit channels to requestIds)
*/
