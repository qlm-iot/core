package routing

type Datastore interface {
	Read(r *Request) (error, Reply)
	Write(r *Write) (error, Reply)
	Cancel(rId string) error
}

type Data struct {
	// Node        string
	Measurement string
	Value       string // Or something else.. do we even need this struct?
}

type Write struct {
	Node       string
	Datapoints []Data
}

type Request struct {
	Node         string
	Measurements []string
	ReplyChan    chan Reply // Pointer?
	// Interval     int32 // Move to QLM part
}

type Reply struct {
	RequestId  string
	Node       string
	Datapoints []Data
}
