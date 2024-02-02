# Pipeline

```go
// Create a new pipeline connect to a Relay
pipeline := New("wss://relay.damus.io/")

// Create a query pipeline to view all Tags used in Articles by Author
pipeline.Author(npub).Kind(nostr.KindArticle).Query().Tags()
```
