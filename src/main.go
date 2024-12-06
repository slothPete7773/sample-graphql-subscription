package main

import (
	"fmt"
	"net/http"
	"os"
	core "sample-subscription/src/core/modules"
	"sample-subscription/src/subscription/graphqlws"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

var httpPort = 8787

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
	schema, err := os.ReadFile("./schema.graphql")
	if err != nil {
		panic(err)
	}

	// init graphQL schema
	resolver := core.NewResolver()
	s, err := graphql.ParseSchema(string(schema), resolver, graphql.UseFieldResolvers())
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
