package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Client struct {
	ID     uuid.UUID
	Role   string
	Send   chan []byte
	Closed bool
}

type SSEHub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

func NewSSEHub() *SSEHub {
	hub := &SSEHub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
	go hub.run()
	go hub.heartbeat()
	return hub
}

func (h *SSEHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			slog.Debug("Cliente SSE conectado", slog.String("id", client.ID.String()), slog.String("role", client.Role))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				client.Closed = true
			}
			h.mu.Unlock()
			slog.Debug("Cliente SSE desconectado", slog.String("id", client.ID.String()))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Cliente no lee a tiempo; cerramos para no bloquear
					close(client.Send)
					client.Closed = true
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// heartbeat envía un ping cada 15 segundos para evitar timeouts de red móvil
func (h *SSEHub) heartbeat() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		pingMsg := []byte(": ping\n\n")
		for client := range h.clients {
			select {
			case client.Send <- pingMsg:
			default:
			}
		}
		h.mu.RUnlock()
	}
}

func (h *SSEHub) Broadcast(eventType string, payload interface{}) {
	event := SSEEvent{
		Type: eventType,
		Data: payload,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		slog.Error("Error serializando evento SSE", slog.Any("error", err))
		return
	}

	formattedMsg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(bytes)))
	h.broadcast <- formattedMsg
}

func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request, role string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := &Client{
		ID:     uuid.New(),
		Role:   role,
		Send:   make(chan []byte, 64),
		Closed: false,
	}

	h.register <- client

	// Notificar conexión inicial
	initialEvent := fmt.Sprintf("event: connected\ndata: {\"client_id\":\"%s\"}\n\n", client.ID.String())
	w.Write([]byte(initialEvent))
	flusher.Flush()

	notify := r.Context().Done()
	go func() {
		<-notify
		h.unregister <- client
	}()

	for msg := range client.Send {
		w.Write(msg)
		flusher.Flush()
	}
}
