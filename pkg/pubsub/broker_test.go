package pubsub

import (
	"context"
	"io/ioutil"
	"log"
	"sync"
	"testing"
)

const testTopic = "testing_topic"
const testData = "testing_message"

func TestBroker_Subscribe(t *testing.T) {
	b := Broker{
		topics: make(map[string][]*Subscriber),
		rwm:    sync.RWMutex{},
		l:      log.New(ioutil.Discard, "", 0),
	}

	subscribers := []*Subscriber{
		{
			topic: testTopic,
		},
		{
			topic: testTopic,
		},
	}

	for i := range subscribers {
		if err := b.Subscribe(context.Background(), subscribers[i]); err != nil {
			t.Errorf("unexpected error subscribing to broker : %s", err)
		}
	}

	subs, ok := b.topics[testTopic]
	if !ok {
		t.Error("expected topic \"test\" to be registered with the broker")
	}

	if len(subs) != len(subscribers) {
		t.Errorf("expected topic \"test\" to contain %v subscribers, but got %v", len(subscribers), len(subs))
	}

}

func TestBroker_Broadcast(t *testing.T) {
	// waitgroup is used to ensure messages are being handled with the handleFn below
	wg := sync.WaitGroup{}
	handleFn := func(ctx context.Context, msg *Message) error {
		wg.Done()
		return nil
	}

	subscribers := []*Subscriber{
		NewSubscriber(testTopic, handleFn),
		NewSubscriber(testTopic, handleFn),
	}

	// Start listeners to ensure messages are being processed
	for _, s := range subscribers {
		wg.Add(1)
		go s.Listen()
	}

	b := &Broker{
		topics: map[string][]*Subscriber{testTopic: subscribers},
		rwm:    sync.RWMutex{},
		l:      log.New(ioutil.Discard, "", 0),
	}
	if err := b.Broadcast(context.Background(), []byte(testData), testTopic); err != nil {
		t.Errorf("Broadcast() received unexpected error: %s", err)
	}

	// Wait for messages to get handled
	wg.Wait()
}

func TestBroker_Unsubscribe(t *testing.T) {
	subscribers := []*Subscriber{
		NewSubscriber(testTopic, func(ctx context.Context, msg *Message) error { return nil }),
		NewSubscriber(testTopic, func(ctx context.Context, msg *Message) error { return nil }),
	}

	b := &Broker{
		topics: map[string][]*Subscriber{testTopic: subscribers},
		rwm:    sync.RWMutex{},
		l:      log.New(ioutil.Discard, "", 0),
	}

	if err := b.Unsubscribe(context.Background(), subscribers[0]); err != nil {
		t.Errorf("Unsubscribe() received unexpected error: %s", err)
	}

	if subs := b.topics[testTopic]; len(subs) != 1 {
		t.Errorf("expected length of subscribers to be 1 after calling Unsubscribe() but got %v", len(subs))
	}

}
