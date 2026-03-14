package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Conn wraps a websocket.Conn with a write mutex to prevent concurrent writes.
type Conn struct {
	Raw *websocket.Conn
	mu  sync.Mutex
}

func NewConn(raw *websocket.Conn) *Conn {
	return &Conn{Raw: raw}
}

// SendJSON safely writes a JSON message to the websocket.
func (c *Conn) SendJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Raw.WriteJSON(v)
}

func (c *Conn) Close() error {
	return c.Raw.Close()
}
