package realtime

import (
	"encoding/json"
	"sync"
)

// Event is pushed to connected clients via SSE.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub manages SSE subscriber channels per user.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{subscribers: make(map[string]map[chan []byte]struct{})}
}

func (h *Hub) Subscribe(userID string) chan []byte {
	ch := make(chan []byte, 16)
	h.mu.Lock()
	if h.subscribers[userID] == nil {
		h.subscribers[userID] = make(map[chan []byte]struct{})
	}
	h.subscribers[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(userID string, ch chan []byte) {
	h.mu.Lock()
	if subs, ok := h.subscribers[userID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(h.subscribers, userID)
		}
	}
	h.mu.Unlock()
	close(ch)
}

func (h *Hub) Publish(userID string, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subscribers[userID] {
		select {
		case ch <- data:
		default:
		}
	}
}

func (h *Hub) PublishRole(role string, event Event) {
	// Role-wide fan-out uses synthetic channel keys "role:ROLE_NAME".
	h.Publish("role:"+role, event)
}
