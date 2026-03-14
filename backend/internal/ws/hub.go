package ws

import (
	"log"
	"sync"
)

// Hub manages topic-based pub/sub for websocket connections.
// Topics are strings like "zone:8664a4b17ffffff".
type Hub struct {
	mu     sync.RWMutex
	topics map[string]map[*Conn]bool

	// OnFirstSubscribe is called when a topic goes from 0→1 subscribers.
	OnFirstSubscribe func(topic string)
	// OnSubscribe is called on every subscribe.
	OnSubscribe func(topic string)
}

func NewHub() *Hub {
	return &Hub{
		topics: make(map[string]map[*Conn]bool),
	}
}

// Subscribe adds a conn to a topic.
func (h *Hub) Subscribe(topic string, c *Conn) {
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
	if h.OnSubscribe != nil {
		h.OnSubscribe(topic)
	}
}

// Unsubscribe removes a conn from a topic.
func (h *Hub) Unsubscribe(topic string, c *Conn) {
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
func (h *Hub) UnsubscribeAll(c *Conn) {
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
func (h *Hub) Broadcast(topic string, payload interface{}) {
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
func (h *Hub) SubscriberCount(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.topics[topic])
}
