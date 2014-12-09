package routing

type Datastore interface {
	Read(r *Request) (error, Reply)
	Write(r *Write) (error, Reply)
	Cancel(rId string) error
	NodeList() []string
	SourceList(node string) []string
}

type Data struct {
	Measurement string
	Value       string
	Timestamp   int64
}

type Write struct {
	Node       string
	Datapoints []Data
}

type Request struct {
	Node         string
	Measurements []string
	ReplyChan    chan Reply
}

type Reply struct {
	RequestId  string
	Node       string
	Datapoints []Data
}
