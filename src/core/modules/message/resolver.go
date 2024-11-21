package message

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MessageResolver struct {
	MessageEvents       chan *Message
	HelloSaidSubscriber chan *OnMessageSubscriber
}

func (MessageResolver) Hello() string {
	return "Hello"
}

func (r MessageResolver) OnMessage(ctx context.Context, input struct{ Filter *string }) <-chan *Message {
	c := make(chan *Message)
	filter := ""
	if input.Filter != nil {
		filter = *input.Filter
	}
	// NOTE: this could take a while
	r.HelloSaidSubscriber <- &OnMessageSubscriber{Events: c, Stop: ctx.Done(), Filter: filter}

	return c
}
func (MessageResolver) SendMessage(ctx context.Context, input struct{ Msg string }) Message {
	return Message{
		Id:  uuid.New().String(),
		Msg: input.Msg,
	}

func (r *MessageResolver) BroadcastMessageEvent() {
	subscribers := map[string]*OnMessageSubscriber{}
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
				go func(id string, s *OnMessageSubscriber) {
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
