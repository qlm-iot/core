Open Group IoT implementation
===
This core implementation is a proxy / gathering layer for Open Group IoT nodes. At the start of the project, the specification was known as QLM (Quantum Lifecycle Mechanism) and this term is referred in the comments and function names of the application. It is capable of receiving messages of version 1.0 of the specification.

This project was done for the T-106.5700 course at Aalto University under the supervision of Kary Främling. 

## Scope

The project scope was to implement a working node that would accept and return Open Group IoT messages (DF and MI). To access the core node some clients were also required to be written that could push data to the node and also retrieve the data. Multi-node message passing and routing were not in the scope of this project.

This core implementation expects nodes to push data to it using write messages and does not itself generate any meaningful statistics. It is also capable of accepting subscription requests from other nodes and store the updated data for the clients to fetch. It is however relatively easy to use the building blocks of this project to create a node that pushes statistics by implementing a new datastore interface as described below in the architectural section. 

## Motivation

Some of the project members' choices affected the architectural design of the system. While our choices might not have been the same as the Open Group working group intended, we had from our perspective good reasons for our choices.

### Choice of Golang as implementation language

The project members had different backgrounds on software engineering side. Some were more experienced than others, but the main obstacle in choosing a tool was that our best known programming languages were different. It was then chosen that we would choose a language that no one had experience previously, but something that everyone would like to learn. Golang was the choice, but also it was a good language for this sort of project as the language is simple in syntax, compiles with static linking (easy to distribute), has fast but lean runtime and is designed for systems programming.

### WebSockets as the main communication protocol

At the specification design there was a wish for a requirement that nodes would interract and push data through sockets instead of only using REST-interface. We opted to use WebSockets as the basis of our sockets connection because it has few distinct advantages. It has very low overhead over clean socket implementation, however it provides standard semantics on how to negotiate connection and encode messages. There was no need to reimplement or redesign such aspects. The other major advantage it does have over many socket implementations is the ability to use HTTP connections for the negotiation part, which means that firewalls are more likely to accept the communication and the nodes do not necessarily require any additional firewall openings.

## Future work

This project is not a finished product and should not be used as such. While it was stable in our tests, there has been no stress testing or long-term experience on using it. For complete reference implementation of the Open Group IoT platform, it would be required to implement routing and multi-node features. Also some features might not be exactly like they've been specified as some small details might have been omitted or missed while developing. There has been no formal requirements checking as this would fall in to the QE-process which was limited during our project.

Other notable small improvements that would be good to start with would be changing packaging infrastructure by moving at least datastore from the routing package to somewhere else. This would allow referencing the messaging layer and communication layers separately in another project instead of pulling also datastore implemention as a dependency (the datastore does not have to be used in the referencing project, but the sources would be pulled - compiler itself will ignore them).

Some usability improvements would be to export internal statistics through the APIs. The amount of subscriptions, currently stored data amounts and client details would help in debugging as well as monitoring how to product works when used in the real life. Also a very good datastore implemention (or a secondary - it would not be difficult to run several at the same time) would be to send statistics to a graph interface such as Grafana or Graphite.

There are plenty of opportunities to extend the project and use it, as the Golang is supported under many different platforms. 

## Core's architectural structure

The project uses a chain of responsibility design pattern, where each layer is responsible for single thing. There are three layers in the core project, one responsible for communication (WebSockets and HTTP), one for message parsing and Open Group functionality, and the last one is the datastore for metrics. This modularity enhances the reusability of the code even if the requirements change. While this example is only used as a proxy layer for the requests, by writing a new datastore it would be possible to add event generation here also - without any rewrite necessary for the connection or messaging layers.

### Connection layer

The connection layer has two implementations, basic HTTP messaging as well as WebSockets implementation. These protocols are both provided by the standard library of Golang, so it was relatively easy to implement both features as all the message parsing is done independently of the used communication layer. No request is bound to certain protocol, so it is possible for the same subscription to use all the communication protocols in parallel as the connection layer's only purpose is to transfer the message between messaging layer and the socket.

### Messaging layer

The messaging layer's function is to understand the incoming messages and act based on the information. It understands the semantics of the Open Group IoT messages and does the necessary marshalling and unmarshalling of messages from input document format to internal structures. Unlike the connection layer, the messaging layer is not stateless, but instead it stores necessary data for tracking the related queries. 

The implementation is in the qlm.go source file.

### In-memory datastore

This sample implementation is not intended to be used as a persistent datastore and as such does not provide any persistence. All the information is kept in the memory and only the last value is saved. Heap based hashmap is used for this purpose as there's no problem with GC in the scenarios this implementation is intended for. Access to the datastore is encapsulated through an interface, so it is possible to use other layers with a new backend storage just by writing another implementation of Datastore-interface and use new implementation in the main.go

As an exception to the last-value rule, the datastore supports subscription based accessing, where data is written for a subscription each time it is otherwise updated to the datastore. This is done with a help of go channels. This works fine for small queue amounts, but is not intended for a rare-polling scenario where millions of datapoints would be fetched at a later time. For such requirements, another datastructure should be chosen. 

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

## Core usage

After the compilation, the core server can be started with ```./core```. If you want to modify the listening port, use the flag ```--port``` with a valid integer value. To modify listening address, use ```--addr```

### Compiling instructions

Compilation is simple using the go get command to fetch everything that is necessary.

```
[michael@burmanmgo ~]$ go get github.com/qlm-iot/core
[michael@burmanmgo ~]$ cd $GOPATH/src/github.com/qlm-iot/core/
[michael@burmanmgo ~/go/src/github.com/qlm-iot/core]$ go build
[michael@burmanmgo ~/go/src/github.com/qlm-iot/core]$ ./core
Listening for connections at localhost:8000
^C[michael@burmanmgo ~/go/src/github.com/qlm-iot/core]$ ./core --port 8001
Listening for connections at localhost:8001
^C[michael@burmanmgo ~/go/src/github.com/qlm-iot/core]$
```

### Demo instance?

Demo instance is available at http://69.28.83.51:8000/qlm/Objects/

### Known bugs

At the end of the project, following bugs were noticed.

* The core prints when a websocket connection is closed a small hex-string to the stdout.
* There's a small potential memory leak if user constantly creates a new subscription, cancels it and chooses another datapoint to follow. The amount that could be (potentially not available for cleaning for the GC) leaked by a buggy delete function is however very small. The bug is in the datastore.go lines 145-154 and the code was commented out because it failed as the demo. 
