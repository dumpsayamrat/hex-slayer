package ws

import (
	"log"
	"sync"
)

// Hub manages topic-based pub/sub for websocket connections.
// Topics are strings like "zone:8664a4b17ffffff".
// Any part of the app can call Broadcast to send to all subscribers of a topic.
var Hub = &hub{
	topics: make(map[string]map[*Conn]bool),
}

type hub struct {
	mu     sync.RWMutex
	topics map[string]map[*Conn]bool

	// OnFirstSubscribe is called when a topic goes from 0→1 subscribers.
	OnFirstSubscribe func(topic string)
}

// Subscribe adds a conn to a topic.
func (h *hub) Subscribe(topic string, c *Conn) {
	h.mu.Lock()
	wasEmpty := len(h.topics[topic]) == 0
	if h.topics[topic] == nil {
		h.topics[topic] = make(map[*Conn]bool)
	}
	h.topics[topic][c] = true
	count := len(h.topics[topic])
	h.mu.Unlock()

	log.Printf("ws hub: conn subscribed to %s (%d subscribers)", topic, count)

	if wasEmpty && h.OnFirstSubscribe != nil {
		h.OnFirstSubscribe(topic)
	}
}

// Unsubscribe removes a conn from a topic.
func (h *hub) Unsubscribe(topic string, c *Conn) {
	h.mu.Lock()
	if conns, ok := h.topics[topic]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.topics, topic)
		}
	}
	h.mu.Unlock()
}

// UnsubscribeAll removes a conn from every topic. Call this on disconnect.
func (h *hub) UnsubscribeAll(c *Conn) {
	h.mu.Lock()
	for topic, conns := range h.topics {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.topics, topic)
		}
	}
	h.mu.Unlock()
}

// Broadcast sends a JSON payload to all conns subscribed to a topic.
// Failed sends are silently dropped (client probably disconnected).
func (h *hub) Broadcast(topic string, payload interface{}) {
	h.mu.RLock()
	conns := make([]*Conn, 0, len(h.topics[topic]))
	for c := range h.topics[topic] {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		if err := c.SendJSON(payload); err != nil {
			log.Printf("ws hub: broadcast error on topic %s: %v", topic, err)
		}
	}
}

// SubscriberCount returns how many conns are subscribed to a topic.
func (h *hub) SubscriberCount(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.topics[topic])
}
