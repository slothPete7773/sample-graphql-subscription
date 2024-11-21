package message

import (
	"context"

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
}
