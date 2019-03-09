package websocket

import (
	"encoding/json"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*ClientHandler]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *ClientHandler

	// Unregister requests from clients.
	unregister chan *ClientHandler
}

func NewHub() *Hub {
	tmp := &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *ClientHandler),
		unregister: make(chan *ClientHandler),
		clients:    make(map[*ClientHandler]bool),
	}
	go tmp.run()
	return tmp
}

func (h *Hub) Broadcast(message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	h.broadcast <- payload
	return nil
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
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
