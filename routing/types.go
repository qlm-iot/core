package routing

type Datastore interface {
	NewStore()
	Read(r *Request)
	Write(r *Request)
	Cancel(rId string)
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
