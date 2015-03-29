Open Group IoT implementation
===
This core implementation is a proxy / gathering layer for Open Group IoT nodes. At the start of the project, the specification was known as QLM (Quantum Lifecycle Mechanism) and this term is referred in the comments and function names of the application. It is capable of receiving messages of version 1.0 of the specification.

* Scope

This core implementation expects nodes to push data to it using write messages and does not itself generate any meaningful statistics. It is also capable of accepting subscription requests from other nodes, but it does not itself produce any data. 

* Motivation

** Choice of Golang as implementation language

** WebSockets as the main communication protocol

Why golang etc.
Why websockets?

* Core's architectural structure

The project uses a chain of responsibility design pattern, where each layer is responsible for single thing. There are three layers in the core project, one responsible for communication (WebSockets and HTTP), one for message parsing and Open Group functionality, and the last one is the datastore for metrics. This modularity enhances the reusability of the code even if the requirements change. While this example is only used as a proxy layer for the requests, by writing a new datastore it would be possible to add event generation here also - without any rewrite necessary for the connection or messaging layers.

** Connection layer

The connection layer has two implementations, basic HTTP messaging as well as WebSockets implementation. These protocols are both provided by the standard library of Golang, so it was relatively easy to implement both features as all the message parsing is done independently of the used communication layer. No request is bound to certain protocol, so it is possible for the same subscription to use all the communication protocols in parallel as the connection layer's only purpose is to transfer the message between messaging layer and the socket.

** Messaging layer

The messaging layer's function is to understand the incoming messages and act based on the information. It understands the semantics of the Open Group IoT messages and does the necessary marshalling and unmarshalling of messages from input document format to internal structures. Unlike the connection layer, the messaging layer is not stateless, but instead it stores necessary data for tracking the related queries. 

The implementation is in the qlm.go source file.

** In-memory datastore

This sample implementation is not intended to be used as a persistent datastore and as such does not provide any persistence. All the information is kept in the memory and only the last value is saved. Heap based hashmap is used for this purpose as there's no problem with GC in the scenarios this implementation is intended for. Access to the datastore is encapsulated through an interface, so it is possible to use other layers with a new backend storage just by writing another implementation of Datastore-interface and use new implementation in the main.go

To get persistence the simplest way would be to write snapshots of internals maps as memory mapped files, but as a proxy layer it's questionable if there's any need for such functionality. The datastore implementation is not aware of anything that's related to the IoT messages, instead it implements a very simple interface:

```golang
type Datastore interface {
	Subscribe(r *Request) (error, Reply)
	ReadImmediate(r *Request) (error, Reply)
	Write(r *Write) (error, Reply)
	Cancel(rId string) error
	NodeList() []string
	SourceList(node string) (error, []string)
}
```

* Core usage

** Installation instructions

** Demo instance?
