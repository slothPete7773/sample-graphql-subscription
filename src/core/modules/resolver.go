package core

import (
	"sample-subscription/src/core/modules/message"
)

type Resolver struct {
	message.MessageResolver
}

func NewResolver() *Resolver {
	r := Resolver{
		// Option 1
		MessageResolver: message.MessageResolver{
			MessageEvents:       make(chan *message.Message),
			HelloSaidSubscriber: make(chan *message.OnMessageSubscriber),
		},
	}
	// // Option 2
	// r.MessageResolver.MessageEvents = make(chan *message.Message)
	// r.MessageResolver.HelloSaidSubscriber = make(chan *message.OnMessageSubscriber)

	go r.MessageResolver.BroadcastMessageEvent()

	return &r
}
