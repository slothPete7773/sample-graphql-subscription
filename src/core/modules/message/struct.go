package message

type Message struct {
	Id  string
	Msg string
}

type OnMessageSubscriber struct {
	Stop   <-chan struct{}
	Events chan<- *Message
	Filter string
}
