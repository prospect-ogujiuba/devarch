package events

import (
	"fmt"
	"sync"
	"time"
)

type Bus struct {
	mu          sync.Mutex
	nextSeq     uint64
	nextSub     int
	now         func() time.Time
	subscribers map[int]chan Envelope
}

func NewBus() *Bus {
	return &Bus{
		now:         time.Now,
		subscribers: make(map[int]chan Envelope),
	}
}

func (b *Bus) SetNow(now func() time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if now == nil {
		b.now = time.Now
		return
	}
	b.now = now
}

func (b *Bus) Publish(spec Spec) (Envelope, error) {
	if b == nil {
		return Envelope{}, fmt.Errorf("events bus: nil bus")
	}
	payload, err := EncodePayload(spec.Payload)
	if err != nil {
		return Envelope{}, err
	}

	b.mu.Lock()
	b.nextSeq++
	sequence := b.nextSeq
	timestamp := spec.Timestamp
	if timestamp.IsZero() {
		timestamp = b.now()
	}
	subscribers := make([]chan Envelope, 0, len(b.subscribers))
	for _, subscriber := range b.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	b.mu.Unlock()

	envelope := Envelope{
		Sequence:  sequence,
		Workspace: spec.Workspace,
		Resource:  spec.Resource,
		Kind:      spec.Kind,
		Timestamp: timestamp,
		Payload:   payload,
	}
	for _, subscriber := range subscribers {
		subscriber <- envelope
	}
	return envelope, nil
}

func (b *Bus) Subscribe(buffer int) (<-chan Envelope, func()) {
	if b == nil {
		ch := make(chan Envelope)
		close(ch)
		return ch, func() {}
	}
	if buffer <= 0 {
		buffer = 1
	}
	ch := make(chan Envelope, buffer)
	b.mu.Lock()
	id := b.nextSub
	b.nextSub++
	b.subscribers[id] = ch
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if existing, ok := b.subscribers[id]; ok {
			delete(b.subscribers, id)
			close(existing)
		}
	}
}
