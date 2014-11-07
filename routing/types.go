package routing

type Datastore interface {
	Read(r *Request) error
	Write(r *Request) error
	Cancel(rId string) error
}

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
