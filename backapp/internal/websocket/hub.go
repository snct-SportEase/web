package websocket

import "encoding/json"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	stopped chan struct{}
	onEmpty func()
}

func NewHub(onEmpty ...func()) *Hub {
	var emptyCallback func()
	if len(onEmpty) > 0 {
		emptyCallback = onEmpty[0]
	}
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		stopped:    make(chan struct{}),
		onEmpty:    emptyCallback,
	}
}

func (h *Hub) Run() {
	defer close(h.stopped)
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if len(h.clients) == 0 && h.onEmpty != nil {
					h.onEmpty()
					return
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	case <-h.stopped:
	}
}

func (h *Hub) BroadcastJSON(v interface{}) {
	msg, err := json.Marshal(v)
	if err != nil {
		// handle error
		return
	}
	h.Broadcast(msg)
}
