#PubSubLite 
PubSubLite is a simple in-memory implementation of PubSub with an HTTP API server. This is **NOT** production ready as 
it has some glaring issues.

---

### Installation
`go get github.com/mdanzinger/pubsublite`

### Usage
Simply run ```go run main.go -port=<PORT>``` from within the `cmd/pubsublite` directory.

Once running, visit the `http://localhost:<PORT>/subscribe/<TOPIC>` url to subscribe to a topic, 
and start broadcasting by POSTing to `http://localhost:<PORT>/publish/<TOPIC>` like so:

```
 curl -d 'Some Message' -H "Content-Type: application/json" -X POST http://localhost:8111/publish/some_topic
```
