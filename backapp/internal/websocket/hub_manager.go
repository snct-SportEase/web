package websocket

import "sync"

type HubManager struct {
	hubs map[string]*Hub
	mu   sync.Mutex
}

func NewHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[string]*Hub),
	}
}

func (m *HubManager) GetHub(topic string) *Hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[topic]; ok {
		return hub
	}

	hub := NewHub()
	m.hubs[topic] = hub
	go hub.Run()
	return hub
}

func (m *HubManager) BroadcastTo(topic string, v interface{}) {
	m.mu.Lock()
	hub, ok := m.hubs[topic]
	m.mu.Unlock()

	if ok {
		hub.BroadcastJSON(v)
	}
}
