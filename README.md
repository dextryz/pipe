# Pipeline

- Manipulating tags

```go
// Create a new pipeline connect to a Relay
pipeline := New("wss://relay.damus.io/")

// Create a query pipeline to view all Tags used in Articles by Author and sort by name.
pipeline.Authors([]string{npub}).Kind([]int{nostr.KindArticle}).Query().Tags().SortByName().Stdout()

// Create a query pipeline to view all Tags used in Articles by Author and sort by tag count.
pipeline.Authors([]string{npub}).Kind([]int{nostr.KindArticle}).Query().Tags().SortByCount().Stdout()
```

- Pipeline between relays

```go
// Pull article from relay for a specific user and publish to new relay with your secret key
export SECRET_KEY=nsec...
pipeline := New("wss://relay.damus.io/")
pipeline.Authors([]string{npub}).Kind([]int{nostr.KindArticle}).Query().Publish("wss://mynew.relay.io/")
```
