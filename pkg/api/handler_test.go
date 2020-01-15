package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mdanzinger/pubsublite/pkg/pubsub"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"nhooyr.io/websocket"
	"strings"
	"testing"
)

const testTopic = "test_topic"
const testMessage = "test_message"

// golden path e2e test
func TestHandle(t *testing.T) {
	// Build dependencies
	l := log.New(ioutil.Discard, "", 0)
	h := NewHandler(pubsub.NewBroker(l), l)
	s := httptest.NewServer(h)
	defer s.Close()

	// build url, replace http with ws
	u := "ws" + strings.TrimPrefix(s.URL, "http")
	u = fmt.Sprintf("%s/ws/subscribe/%s", u, testTopic)

	// Subscribe to a topic
	conn, _, err := websocket.Dial(context.Background(), u, nil)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Publish
	p := fmt.Sprintf("%s/publish/%s", s.URL, testTopic)
	r, err := http.NewRequest(http.MethodPost, p, bytes.NewBuffer([]byte(testMessage)))
	if err != nil {
		t.Errorf("%v", err)
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer resp.Body.Close()

	// Ensure subscriber received message
	_, d, err := conn.Read(context.Background())
	if err != nil {
		t.Errorf("conn.Read received unexpected error : %s", err)
	}

	if string(d) != testMessage {
		t.Errorf("expected to published message %s, but got %s", testMessage, string(d))
	}

}
