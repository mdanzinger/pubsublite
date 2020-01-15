package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/mdanzinger/pubsublite/pkg/pubsub"
	"github.com/urfave/negroni"
	"io/ioutil"
	"log"
	"net/http"
	"nhooyr.io/websocket"
)

type Handler struct {
	r      *mux.Router
	broker *pubsub.Broker
	l      *log.Logger
}

func (h *Handler) routes() {

	h.r.HandleFunc("/subscribe/{topic}", h.renderSubscribe())
	h.r.HandleFunc("/publish/{topic}", h.publishMessage()).Methods("POST")

	// Websocket routes
	ws := h.r.PathPrefix("/ws").Subrouter()
	ws.HandleFunc("/subscribe/{topic}", h.registerSubscriber())
}

func (h *Handler) registerSubscriber() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topic := vars["topic"]

		// Create ws connection
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		s := pubsub.NewSubscriber(topic, wsHandler(conn), pubsub.WithLogger(h.l))

		if err := h.broker.Subscribe(r.Context(), s); err != nil {
			h.l.Printf("got error adding subscription: %s", err)
		}
	}
}

func (h *Handler) publishMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topic := vars["topic"]

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			h.l.Printf("error reading response body : %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err := h.broker.Broadcast(r.Context(), data, topic); err != nil {
			h.l.Printf("got error adding subscription: %s", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// wsHandler is responsible for publishing the published messages to the websocket conn
func wsHandler(conn *websocket.Conn) func(ctx context.Context, msg *pubsub.Message) error {
	return func(ctx context.Context, msg *pubsub.Message) error {
		return conn.Write(ctx, websocket.MessageText, msg.Data)
	}
}

func (h *Handler) renderSubscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mux.Vars(r)
		w.Write([]byte(subscribeTmpl))
	}
}

// NewHandler returns a handler with the API routes initialized
func NewHandler(broker *pubsub.Broker, logger *log.Logger) http.Handler {
	h := &Handler{
		r:      mux.NewRouter(),
		broker: broker,
		l:      logger,
	}
	h.routes()

	n := negroni.Classic()
	n.UseHandler(h.r)

	return n
}
