package server

import (
	"encoding/json"
	"sync"
)

// SSEEvent is the payload published to team subscribers.
type SSEEvent struct {
	Type        string `json:"type"`
	StageNumber int    `json:"stageNumber,omitempty"`
	PlayerName  string `json:"playerName,omitempty"`
	IsCorrect   bool   `json:"isCorrect,omitempty"`
}

// Broker is an in-process pub/sub for SSE events, keyed by team ID.
type Broker struct {
	mu   sync.RWMutex
	subs map[string]map[chan []byte]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		subs: make(map[string]map[chan []byte]struct{}),
	}
}

// Subscribe returns a channel that receives JSON-encoded SSE events for the given team.
func (b *Broker) Subscribe(teamID string) chan []byte {
	ch := make(chan []byte, 16)
	b.mu.Lock()
	if b.subs[teamID] == nil {
		b.subs[teamID] = make(map[chan []byte]struct{})
	}
	b.subs[teamID][ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a channel from the team's subscribers.
func (b *Broker) Unsubscribe(teamID string, ch chan []byte) {
	b.mu.Lock()
	delete(b.subs[teamID], ch)
	if len(b.subs[teamID]) == 0 {
		delete(b.subs, teamID)
	}
	b.mu.Unlock()
}

// Publish sends an event to all subscribers of the given team.
func (b *Broker) Publish(teamID string, event SSEEvent) {
	data, _ := json.Marshal(event)
	b.mu.RLock()
	for ch := range b.subs[teamID] {
		select {
		case ch <- data:
		default:
			// Drop if subscriber is slow.
		}
	}
	b.mu.RUnlock()
}
