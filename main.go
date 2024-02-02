package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// Only connect to a single relay for the time being.
// We have to look at the Pool impl in go-nostr.
// The problem is pulling the same event from multiple relays.

func StringEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("address env variable \"%s\" not set, usual", key)
	}
	return value
}

var (
	PRIVATE_KEY = StringEnv("PRIVATE_KEY")
)

type Pipeline struct {
	Relay  *nostr.Relay
	Reader *EventBuffer
	Output io.Writer
	Error  error
}

func New(relay string) *Pipeline {

	ctx := context.Background()

	r, err := nostr.RelayConnect(ctx, relay)
	if err != nil {
		panic(err)
	}

	eb := &EventBuffer{
		filter: &nostr.Filter{},
	}

	return &Pipeline{
		Relay:  r,
		Output: os.Stdout,
		Reader: eb,
	}
}

func (s *Pipeline) Author(npub string) *Pipeline {
	_, pk, err := nip19.Decode(npub)
	if err != nil {
		panic(err)
	}
	s.Reader.filter.Authors = []string{pk.(string)}
	s.Reader.filter.Limit = 1000
	return s
}

func (s *Pipeline) Authors(npubs []string) *Pipeline {
	pk := []string{}
	for _, npub := range npubs {
		_, v, err := nip19.Decode(npub)
		if err != nil {
			panic(err)
		}
		pk = append(pk, v.(string))
	}
	s.Reader.filter.Authors = pk
	s.Reader.filter.Limit = 1
	return s
}

func (s *Pipeline) Kind(kind int) *Pipeline {
	s.Reader.filter.Kinds = []int{kind}
	return s
}

func (s *Pipeline) Kinds(kinds []int) *Pipeline {
	s.Reader.filter.Kinds = kinds
	return s
}

func (s *Pipeline) Publish(relay string) {

	ctx := context.Background()

	r, err := nostr.RelayConnect(ctx, relay)
	if err != nil {
		panic(err)
	}

	// There has to be events cached in the buffer.
	// It makes no sense to publish an empty buffer no then does it?
	// Impl proper cheks and balances
	for _, e := range s.Reader.events {

		event := &nostr.Event{
			Kind:      e.Kind,
			Content:   e.Content,
			CreatedAt: nostr.Now(),
		}

		_, sk, err := nip19.Decode(PRIVATE_KEY)
		if err != nil {
			panic(err)
		}

		event.Sign(sk.(string))

		err = r.Publish(ctx, *event)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Published")
	}
}

func (s *Pipeline) Query() *Pipeline {

	ctx := context.Background()

	events, err := s.Relay.QuerySync(ctx, *s.Reader.filter)
	if err != nil {
		log.Fatalln(err)
	}

	s.Reader.events = events

	s.Reader.SerializeEvents(s.Reader.events)

	return s
}

func (s *Pipeline) Tags() {

	tags := make(map[string]int)

	for _, e := range s.Reader.events {
		for _, t := range e.Tags {
			if t.Key() == "t" {
				_, ok := tags[t.Value()]
				if ok {
					tags[t.Value()] += 1
				} else {
					tags[t.Value()] = 1
				}
			}
		}
	}

	for tag, count := range tags {
		fmt.Printf("%s (%d)\n", tag, count)
	}
}

func (s *Pipeline) Stdout() {
	if s.Error != nil {
		return
	}
	io.Copy(s.Output, s.Reader)
}

func (p *Pipeline) String() (string, error) {
	if p.Error != nil {
		return "", p.Error
	}
	data, err := io.ReadAll(p.Reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
	npub := "npub14ge829c4pvgx24c35qts3sv82wc2xwcmgng93tzp6d52k9de2xgqq0y4jk"

	pipeline := New("wss://relay.damus.io/")
	pipeline.Author(npub).Kind(nostr.KindArticle).Query().Tags()

	//p.Author(npub).Kind(nostr.KindTextNote).Query().Stdout()
	//p.Author(npub).Kind(nostr.KindTextNote).Query().Publish("wss://relay.damus.io/")

	//p.Author(npub).Query().Tags().Stdout()
}
