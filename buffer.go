package main

import (
	"github.com/nbd-wtf/go-nostr"
)

type EventBuffer struct {
	events  []*nostr.Event
	filter  *nostr.Filter
	readPos int // Track the current position in the serialized slice.
}

// SerializeEvents takes a slice of pointers to Event objects and serializes them into a single []byte.
func SerializeEvents(events []*nostr.Event) []byte {
	serialized := []byte("[")
	for i, evt := range events {
		if i > 0 {
			serialized = append(serialized, ',')
		}
		serialized = append(serialized, evt.Serialize()...)
	}
	serialized = append(serialized, ']')
	return serialized
}

func (s *EventBuffer) Read(p []byte) (n int, err error) {

	// 	if s.serialized == nil || s.readPos >= len(s.serialized) {
	// 		return 0, io.EOF // Indicate end of file if there's nothing left to read.
	// 	}
	// 	n = copy(p, s.serialized[s.readPos:])
	// 	s.readPos = s.readPos + n
	// 	return n, nil

	return 0, nil
}
