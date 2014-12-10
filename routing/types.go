package routing

type Datastore interface {
	Subscribe(r *Request) (error, Reply)
	ReadImmediate(r *Request) (error, Reply)
	Write(r *Write) (error, Reply)
	Cancel(rId string) error
	NodeList() []string
	SourceList(node string) ([]string, error)
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
