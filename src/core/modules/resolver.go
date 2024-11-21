package core

import (
	"sample-subscription/src/core/modules/message"
	"strings"
	"time"

	"github.com/google/uuid"
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

	go r.BroadcastMessageEvent()

	return &r
}

func (r *Resolver) BroadcastMessageEvent() {
	subscribers := map[string]*message.OnMessageSubscriber{}
	unsubscribe := make(chan string)

	// NOTE: subscribing and sending events are at odds.
	for {
		select {
		case id := <-unsubscribe:
			delete(subscribers, id)
		case s := <-r.HelloSaidSubscriber:
			subscribers[uuid.NewString()] = s
		case e := <-r.MessageEvents:
			for id, s := range subscribers {
				go func(id string, s *message.OnMessageSubscriber) {
					select {
					case <-s.Stop:
						unsubscribe <- id
						return
					default:
					}

					if s.Filter != "" && !strings.Contains(e.Msg, s.Filter) {
						// Event does not match filter, skip sending
						return
					}

					select {
					case <-s.Stop:
						unsubscribe <- id
					case s.Events <- e:
					case <-time.After(time.Second):
					}
				}(id, s)
			}
		}
	}
}
