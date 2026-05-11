package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	Conn      *websocket.Conn
	UserID    string
	Channels  map[string]bool
	mu        sync.Mutex
}

type Hub struct {
	clients    map[*Client]bool
	channels   map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		channels:   make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				// Remove from all channels
				for chID := range client.Channels {
					if clients, ok := h.channels[chID]; ok {
						delete(clients, client)
					}
				}
				client.Conn.Close()
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) SubscribeToChannel(client *Client, channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.channels[channelID]; !ok {
		h.channels[channelID] = make(map[*Client]bool)
	}
	h.channels[channelID][client] = true
	client.Channels[channelID] = true
}

func (h *Hub) UnsubscribeFromChannel(client *Client, channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.channels[channelID]; ok {
		delete(clients, client)
	}
	delete(client.Channels, channelID)
}

func (h *Hub) BroadcastToChannel(channelID string, event Event) {
	h.mu.RLock()
	clients, ok := h.channels[channelID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range clients {
		client.mu.Lock()
		if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("failed to write to client: %v", err)
		}
		client.mu.Unlock()
	}
}
