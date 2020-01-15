package pubsub

import (
	"context"
	"github.com/gofrs/uuid"
	"log"
)

// Handlefunc is used to handle the message. This can be used to deliver to a websocket connection, or passed along
// to some sqs client, etc
type Handlefunc func(ctx context.Context, msg *Message) error

// Subscriber represents a pubsub subscriber
type Subscriber struct {
	// id is used to identify the subscriber
	id string

	// topic subscriber is subscribed to
	topic string

	ch chan *Message

	handler Handlefunc

	l *log.Logger
}

// Listen is responsible for handling the subscriber's messages and will emit an "unsubscribe" signal on error
func (s *Subscriber) Listen() chan struct{} {
	signal := make(chan struct{})

	go func() {
		for msg := range s.ch {
			if err := s.handler(context.Background(), msg); err != nil {
				s.l.Printf("subscriber %s received error handling message: %s", s.id, err)
				s.l.Printf("removing subscriber %s from topic %s", s.id, s.topic)
				signal <- struct{}{}
			}
		}
	}()

	return signal
}

// SubscriberOption is used to add functional params to the NewSubscriber constructor
type SubscriberOption func(*Subscriber)

// NewSubscriber initializes a new subscriber ready to subscribe to a topic
func NewSubscriber(topic string, handler Handlefunc, opts ...SubscriberOption) *Subscriber {
	s := &Subscriber{
		id:      uuid.Must(uuid.NewV4()).String(),
		topic:   topic,
		ch:      make(chan *Message, 0),
		handler: handler,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithCapacity is a functional param and is used to set the capacity for the Subscriber channel.
func WithCapacity(cap int) func(sub *Subscriber) {
	return func(sub *Subscriber) {
		sub.ch = make(chan *Message, cap)
	}
}

func WithLogger(logger *log.Logger) func(sub *Subscriber) {
	return func(sub *Subscriber) {
		sub.l = logger
	}
}
