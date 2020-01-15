package main

import (
	"flag"
	"fmt"
	"github.com/mdanzinger/pubsublite/pkg/api"
	"github.com/mdanzinger/pubsublite/pkg/pubsub"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
)

var port *string

func init() {
	port = flag.String("port", "8111", "port to run server")
}

func main() {
	flag.Parse()

	l := log.New(os.Stdout, "[pubsublite] ", 0)
	b := pubsub.NewBroker(l)
	h := api.NewHandler(b, l)

	// use negroni for logs, panic recovering, etc
	n := negroni.Classic()
	n.UseHandler(h)

	l.Printf("server listening on port :%s", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), n))
}
