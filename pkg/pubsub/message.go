package pubsub

import "time"

// Message is the structure that is published to subscribers
type Message struct {
	// ID identifies the message and could be used for idempotent processing
	ID string

	// Data represents the actual value of the message
	Data []byte

	// PublishTime contains the time the message was published
	PublishTime time.Time
}
