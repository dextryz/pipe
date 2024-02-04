package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Events struct {
	EventList []*nostr.Event `json:"events"`
}

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
	Reader io.Reader
	Output io.Writer
	Error  error
}

func New(relay string) *Pipeline {

	ctx := context.Background()

	r, err := nostr.RelayConnect(ctx, relay)
	if err != nil {
		panic(err)
	}

	return &Pipeline{
		Relay:  r,
		Output: os.Stdout,
	}
}

func (s *Pipeline) Author(npub string) *Pipeline {
	_, pk, err := nip19.Decode(npub)
	if err != nil {
		panic(err)
	}

	f := nostr.Filter{
		Kinds:   []int{nostr.KindArticle},
		Authors: []string{pk.(string)},
		Limit:   10,
	}

	bb, err := json.Marshal(f)
	if err != nil {
		fmt.Println("Error serializing filter:", err)
		return nil
	}

	var b bytes.Buffer // A Buffer needs no initialization.
	b.Write(bb)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) Query() *Pipeline {

	//result := &bytes.Buffer{}
	input := bufio.NewScanner(s.Reader)

	var filter nostr.Filter
	for input.Scan() {

		err := json.Unmarshal(input.Bytes(), &filter)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}
	}

	ctx := context.Background()
	events, err := s.Relay.QuerySync(ctx, filter)
	if err != nil {
		log.Fatalln(err)
	}

	serialized := []byte("{\"events\":[")
	for i, evt := range events {
		if i > 0 {
			serialized = append(serialized, ',')
		}

		be, err := json.Marshal(evt)
		if err != nil {
			log.Fatalln(err)
		}

		serialized = append(serialized, be...)
	}
	serialized = append(serialized, ']')
	serialized = append(serialized, '}')

	var b bytes.Buffer // A Buffer needs no initialization.
	b.Write(serialized)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) Tags() *Pipeline {

	input := bufio.NewScanner(s.Reader)

	var events Events
	for input.Scan() {

		err := json.Unmarshal(input.Bytes(), &events)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}

	}

	tags := make(map[string]int)

	for _, e := range events.EventList {
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

	return &Pipeline{}
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
	pipeline.Author(npub).Query().Tags()

	//p.Author(npub).Kind(nostr.KindTextNote).Query().Stdout()
	//p.Author(npub).Kind(nostr.KindTextNote).Query().Publish("wss://relay.damus.io/")

	//p.Author(npub).Query().Tags().Stdout()
}
