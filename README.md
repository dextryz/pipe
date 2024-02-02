# Atreides

Nostr CLI for managing long form content

- Pipe notes from one relay to another via a filter.

```go

p := pipeline.New("wss://ixian.relay", "wss://damus.relay")
p.Authors("npub...", "npub...").String()
p.Authors("npub...", "npub...").Publish("wss://")

```

```
pipeline.Relays("wss://damus.relay", "wss...").Kinds(30023).Authors("npub...").Publish("wss://ixian.relay")
pipeline.Relays("wss://damus.relay", "wss...").Filter(filter).Publish("wss://ixian.relay")

pipeline.Relays("wss://damus.relay", "wss...").Filter(filter).ToArticle()
```

## Command Line Interface

```shell

> art -title npub
Concurrency in Go (naddr)
Channels in Go (naddr)

> art -tag npub
go (33)
coding (19)
nostr (23)

```

## Piping in Go

```go

// 1. Pull all articles for npub
// 2. List all tags in pulled articles
// 3. Sort by highest frequency (most used)
// 4. Only show the first 10 tags
// 5. Print to stdout

pipeline.Article("npub...").Tag().Freq().First(10).ShowTitle().Stdout()

// 2. Only give articles that contains the words "context" in the article content.

pipeline.Article("npub...").Search("context").Stdout()

// 2. Send a prompt to ChatGPT.

pipeline.Article("npub...").Prompt("Give me all articles related to nostr").ShowContent().Save("nostr.db")

```

## Kind 1

```go

pipeline.TextNote("npub...").NoImage().ShowContent().Stdout()

```
