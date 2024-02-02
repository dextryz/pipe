package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// Only connect to a single relay for the time being.
// We have to look at the Pool impl in go-nostr.
// The problem is pulling the same event from multiple relays.

type Pipeline struct {
	Relay  *nostr.Relay
	Reader *EventBuffer
	Output io.Writer
	Error  error
}

func New() *Pipeline {

	ctx := context.Background()

	r, err := nostr.RelayConnect(ctx, "wss://relay.damus.io/")
	if err != nil {
		panic(err)
	}

	return &Pipeline{
		Relay:  r,
		Output: os.Stdout,
	}
}

func (s *Pipeline) Authors(npubs []string) *Pipeline {

	ctx := context.Background()

	pk := []string{}
	for _, npub := range npubs {
		_, v, err := nip19.Decode(npub)
		if err != nil {
			panic(err)
		}
		pk = append(pk, v.(string))
	}

	filter := nostr.Filter{
		Kinds:   []int{nostr.KindArticle},
		Authors: pk,
		Limit:   1,
	}

	events, err := s.Relay.QuerySync(ctx, filter)
	if err != nil {
		log.Fatalln(err)
	}

	eb := EventBuffer{
		events: events,
	}

	eb.SerializeEvents(eb.events)

	return &Pipeline{
        Output: s.Output,
		Reader: &eb,
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
	log.Println("Hello")

	p := New()

	p.Authors([]string{"npub14ge829c4pvgx24c35qts3sv82wc2xwcmgng93tzp6d52k9de2xgqq0y4jk"}).Stdout()
}
