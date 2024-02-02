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
	s.Reader.filter.Limit = 1
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

	npub := "npub14ge829c4pvgx24c35qts3sv82wc2xwcmgng93tzp6d52k9de2xgqq0y4jk"
	p.Author(npub).Kind(nostr.KindTextNote).Query().Stdout()
}
