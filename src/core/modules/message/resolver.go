package message

import (
	"context"
	"log"
	"sample-subscription/src/pubsub"
	"strings"

	"github.com/google/uuid"
)

type MessageResolver struct {
	PubSub              *pubsub.PubSub
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
	// Subscribe to the topic
	subscription, err := r.PubSub.Subscribe(ctx, "message_topic")
	if err != nil {
		log.Println("Subscription error:", err)
		close(c)
		return c
	}

	go func() {
		defer close(c)
		for {
			select {
			case <-ctx.Done():
				r.PubSub.Unsubscribe("message_topic", subscription.ID)
				return
			case msg := <-subscription.Channel:
				if event, ok := msg.(*Message); ok {
					if filter == "" || strings.Contains(event.Msg, filter) {
						c <- event
					}
				}
			}
		}
	}()

	return c
}
func (r MessageResolver) SendMessage(ctx context.Context, input struct{ Msg string }) Message {
	msg := Message{
		Id:  uuid.New().String(),
		Msg: input.Msg,
	}

	log.Println("Publishing message:", msg)
	if err := r.PubSub.Publish("message_topic", &msg); err != nil {
		log.Println("Error publishing message:", err)
	}
	return msg
}
