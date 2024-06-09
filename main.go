package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
)

const schema = `
	schema {
		subscription: Subscription
		mutation: Mutation
		query: Query
	}

	type Query {
		hello: String!
	}

	type Subscription {
		onMessage(filter: String): Message!
	}

	type Mutation {
		sendMessage(msg: String!): Message!
	}

	type Message {
		id: String!
		msg: String!
	}
`

var httpPort = 8080

func init() {
	port := os.Getenv("HTTP_PORT")
	if port != "" {
		var err error
		httpPort, err = strconv.Atoi(port)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	// init graphQL schema
	s, err := graphql.ParseSchema(schema, newResolver(),graphql.UseFieldResolvers())
	if err != nil {
		panic(err)
	}

	// graphQL handler
	graphQLHandler := graphqlws.NewHandlerFunc(s, &relay.Handler{Schema: s})
	http.HandleFunc("/graphql", graphQLHandler)

	// start HTTP server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
		panic(err)
	}
}


type message struct {
	id  string
	msg string
}

func (r *message) Msg() string {
	return r.msg
}

func (r *message) ID() string {
	return r.id
}

type resolver struct {
	messageEvents     chan *message
	helloSaidSubscriber chan *onMessageSubscriber
}

func newResolver() *resolver {
	r := &resolver{
		messageEvents:     make(chan *message),
		helloSaidSubscriber: make(chan *onMessageSubscriber),
	}

	go r.broadcastMessageEvent()

	return r
}

func (r *resolver) Hello() string {
	return "Hello world!"
}

func (r *resolver) SendMessage(args struct{ Msg string }) *message {
	e := &message{msg: args.Msg, id: uuid.New().String()}
	go func() {
		select {
		case r.messageEvents <- e:
		case <-time.After(1 * time.Second):
		}
	}()
	return e
}

type onMessageSubscriber struct {
	stop   <-chan struct{}
	events chan<- *message
	filter string
}

func (r *resolver) broadcastMessageEvent() {
	subscribers := map[string]*onMessageSubscriber{}
	unsubscribe := make(chan string)

	// NOTE: subscribing and sending events are at odds.
	for {
		select {
			case id := <-unsubscribe:
				delete(subscribers, id)
			case s := <-r.helloSaidSubscriber:
				subscribers[uuid.NewString()] = s
			case e := <-r.messageEvents:
				for id, s := range subscribers {
					go func(id string, s *onMessageSubscriber) {
						select {
						case <-s.stop:
							unsubscribe <- id
							return
						default:
						}

						if s.filter != "" && !strings.Contains(e.msg, s.filter) {
							// Event does not match filter, skip sending
							return
						}

						select {
							case <-s.stop:
								unsubscribe <- id
							case s.events <- e:
							case <-time.After(time.Second):
						}
					}(id, s)
				}
		}
	}
}

func (r *resolver) OnMessage(ctx context.Context, args struct{ Filter *string }) <-chan *message {
	c := make(chan *message)
	filter := ""
	if args.Filter != nil {
		filter = *args.Filter
	}
	// NOTE: this could take a while
	r.helloSaidSubscriber <- &onMessageSubscriber{events: c, stop: ctx.Done(), filter: filter}

	return c
}