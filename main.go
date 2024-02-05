package nostrpipeline

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// Only connect to a single relay for the time being.
// We have to look at the Pool impl in go-nostr.
// The problem is pulling the same event from multiple relays.

func StringEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("env variable \"%s\" not set, usual", key)
	}
	return value
}

type kv struct {
	Key   string
	Value int
}

type Events struct {
	EventList []*nostr.Event `json:"events"`
}

func (s *Events) Serialize() []byte {
	b := []byte("{\"events\":[")
	for i, evt := range s.EventList {
		if i > 0 {
			b = append(b, ',')
		}

		be, err := json.Marshal(evt)
		if err != nil {
			log.Fatalln(err)
		}

		b = append(b, be...)
	}
	b = append(b, ']')
	b = append(b, '}')
	return b
}

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

func (s *Pipeline) Filter(filter nostr.Filter) *Pipeline {

	data, err := json.Marshal(filter)
	if err != nil {
		fmt.Println("error serializing filter:", err)
		return nil
	}
	var b bytes.Buffer
	b.Write(data)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) Kinds(kinds []int) *Pipeline {

	var filter nostr.Filter
	input := bufio.NewScanner(s.Reader)
	for input.Scan() {
		err := json.Unmarshal(input.Bytes(), &filter)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}
	}

	filter.Kinds = kinds

	data, err := json.Marshal(filter)
	if err != nil {
		fmt.Println("error serializing filter:", err)
		return nil
	}
	var b bytes.Buffer
	b.Write(data)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) Authors(npubs []string) *Pipeline {

	var authors []string

	for _, npub := range npubs {
		_, pk, err := nip19.Decode(npub)
		if err != nil {
			panic(err)
		}
		authors = append(authors, pk.(string))
	}

	f := nostr.Filter{
		Kinds:   []int{nostr.KindTextNote},
		Authors: authors,
		Limit:   100,
	}

	data, err := json.Marshal(f)
	if err != nil {
		fmt.Println("error serializing filter:", err)
		return nil
	}
	var b bytes.Buffer
	b.Write(data)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) Query() *Pipeline {

	var filter nostr.Filter
	input := bufio.NewScanner(s.Reader)
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

	el := Events{
		EventList: events,
	}
	var b bytes.Buffer
	b.Write(el.Serialize())

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) Titles() (titles []string) {

	var events Events
	input := bufio.NewScanner(s.Reader)
	for input.Scan() {
		err := json.Unmarshal(input.Bytes(), &events)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}
	}

	for _, e := range events.EventList {
		for _, t := range e.Tags {
			if t.Key() == "title" {
				titles = append(titles, t.Value())
			}
		}
	}

	return titles
}

func (s *Pipeline) Ids() (ids []string) {

	var events Events
	input := bufio.NewScanner(s.Reader)
	for input.Scan() {
		err := json.Unmarshal(input.Bytes(), &events)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}
	}

	for _, e := range events.EventList {

		title := ""
		for _, t := range e.Tags {
			if t.Key() == "title" {
				title = t.Value()
			}
		}

		naddr, err := nip19.EncodeEntity(
			e.PubKey,
			nostr.KindArticle,
			title,
			[]string{},
		)
		if err != nil {
			log.Fatalln(err)
		}
		ids = append(ids, naddr)
	}

	return ids
}

func (s *Pipeline) Tags() *Pipeline {

	var events Events
	input := bufio.NewScanner(s.Reader)
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

	data, err := json.Marshal(tags)
	if err != nil {
		log.Fatalf("Error marshaling map to JSON: %v", err)
	}
	var b bytes.Buffer
	b.Write(data)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) SortByCount() *Pipeline {

	tags := make(map[string]int)
	input := bufio.NewScanner(s.Reader)
	for input.Scan() {
		err := json.Unmarshal(input.Bytes(), &tags)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}
	}

	// Convert map to slice of kv structs.
	var ss []kv
	for k, v := range tags {
		ss = append(ss, kv{k, v})
	}

	// Sort slice by Value.
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value // For descending order, use ss[i].Value > ss[j].Value
	})

	// Optionally, serialize the sorted map to JSON and print.
	data, err := json.Marshal(ss)
	if err != nil {
		log.Fatalf("Error marshaling sorted map to JSON: %v", err)
	}
	var b bytes.Buffer
	b.Write(data)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) SortByName() *Pipeline {

	tags := make(map[string]int)
	input := bufio.NewScanner(s.Reader)
	for input.Scan() {
		err := json.Unmarshal(input.Bytes(), &tags)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}
	}

	// Extract keys.
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}

	// Sort keys.
	sort.Strings(keys)

	// Iterate over sorted keys and build a sorted map.
	sortedMap := make(map[string]int)
	for _, k := range keys {
		sortedMap[k] = tags[k]
		fmt.Printf("%s: %d\n", k, tags[k]) // Print each key-value pair.
	}

	// Optionally, serialize the sorted map to JSON and print.
	data, err := json.Marshal(sortedMap)
	if err != nil {
		log.Fatalf("Error marshaling sorted map to JSON: %v", err)
	}
	var b bytes.Buffer
	b.Write(data)

	return &Pipeline{
		Relay:  s.Relay,
		Reader: &b,
		Output: s.Output,
	}
}

func (s *Pipeline) Publish(relay string) {

	ctx := context.Background()

	r, err := nostr.RelayConnect(ctx, relay)
	if err != nil {
		panic(err)
	}

	events := Events{}
	input := bufio.NewScanner(s.Reader)
	for input.Scan() {
		err := json.Unmarshal(input.Bytes(), &events)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}
	}

	// There has to be events cached in the buffer.
	// It makes no sense to publish an empty buffer no then does it?
	// Impl proper cheks and balances
	for _, e := range events.EventList {

		event := &nostr.Event{
			Kind:      e.Kind,
			Content:   e.Content,
			CreatedAt: nostr.Now(),
		}

		private_key := StringEnv("PRIVATE_KEY")
		_, sk, err := nip19.Decode(private_key)
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
	pipeline.Authors([]string{npub}).Kinds([]int{nostr.KindArticle}).Query().Tags().SortByCount().Stdout()
	//pipeline.Authors([]string{npub}).Query().Publish("wss://relay.damus.io/")

	//p.Author(npub).Kind(nostr.KindTextNote).Query().Stdout()
	//p.Author(npub).Kind(nostr.KindTextNote).Query().Publish("wss://relay.damus.io/")

	//p.Author(npub).Query().Tags().Stdout()
}
