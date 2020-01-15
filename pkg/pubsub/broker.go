package pubsub

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"log"
	"sync"
	"time"
)

var (
	ErrInvalidSubscriber = errors.New("invalid subscriber")
	ErrTopicNotFound     = errors.New("unable to find topic")
)

// Broker is used to manage pubsub state and forward messages to the subscribers of a topic
type Broker struct {
	topics map[string][]*Subscriber

	rwm sync.RWMutex
	l   *log.Logger
}

// Subscribe is used to register a subscriber to the broker
func (b *Broker) Subscribe(ctx context.Context, subscriber *Subscriber) error {
	if subscriber == nil {
		return ErrInvalidSubscriber
	}

	b.rwm.Lock()
	defer b.rwm.Unlock()

	b.l.Printf("adding subscriber to topic %s", subscriber.topic)

	go b.setupSubscriber(subscriber)

	b.topics[subscriber.topic] = append(b.topics[subscriber.topic], subscriber)
	return nil
}

func (b *Broker) setupSubscriber(s *Subscriber) {
	sig := s.Listen()
	_ = <-sig

	if err := b.Unsubscribe(context.Background(), s); err != nil {
		b.l.Printf("error unsubscribing from topic: %s", err)
	}

}

// Broadcast broadcasts the message to all listening subscribers
func (b *Broker) Broadcast(ctx context.Context, data []byte, topic string) error {
	b.rwm.RLock()
	defer b.rwm.RUnlock()
	subs, ok := b.topics[topic]
	if !ok {
		return ErrTopicNotFound
	}

	m := &Message{
		ID:          uuid.Must(uuid.NewV4()).String(),
		Data:        data,
		PublishTime: time.Now(),
	}

	for _, s := range subs {
		go func(sub *Subscriber) {
			b.l.Printf("sending message %s to subscriber %s", m.Data, sub.id)
			sub.ch <- m
		}(s)
	}

	return nil
}

// Unsubscribe will remove a subscriber from a given topic
func (b *Broker) Unsubscribe(ctx context.Context, subscriber *Subscriber) error {
	b.rwm.Lock()
	defer b.rwm.Unlock()
	subs, ok := b.topics[subscriber.topic]
	if !ok {
		return ErrTopicNotFound
	}

	for i, sub := range subs {
		if sub.id == subscriber.id {
			// avoid out of range panic
			if len(subs) == 1 {
				b.topics[subscriber.topic] = []*Subscriber{}
				return nil
			}
			// swap with the last element and truncate slice
			subs[i], subs[len(subs)-1] = subs[len(subs)-1], subs[i]
			b.topics[subscriber.topic] = subs[:len(subs)-1]
			return nil
		}
	}

	return nil
}

// NewBroker is used to initialize a new broker
func NewBroker(l *log.Logger) *Broker {
	return &Broker{
		topics: make(map[string][]*Subscriber),
		rwm:    sync.RWMutex{},
		l:      l,
	}
}
