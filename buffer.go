package main

import (
	"io"

	"github.com/nbd-wtf/go-nostr"
)

type EventBuffer struct {
	events     []*nostr.Event
	serialized []byte
	readPos    int // Track the current position in the serialized slice.
}

// SerializeEvents takes a slice of pointers to Event objects and serializes them into a single []byte.
func (s *EventBuffer) SerializeEvents(events []*nostr.Event) {
	if s.serialized == nil {
		s.serialized = []byte("[")
		for i, evt := range s.events {
			if i > 0 {
				s.serialized = append(s.serialized, ',')
			}
			s.serialized = append(s.serialized, evt.Serialize()...)
		}
		s.serialized = append(s.serialized, ']')
	}
}

func (s *EventBuffer) Read(p []byte) (n int, err error) {

	if s.serialized == nil || s.readPos >= len(s.serialized) {
		return 0, io.EOF // Indicate end of file if there's nothing left to read.
	}
	n = copy(p, s.serialized[s.readPos:])
	s.readPos = s.readPos + n
	return n, nil
}
