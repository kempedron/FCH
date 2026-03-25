package handler

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients sync.Map
}

var chatHub = &Hub{}

func (h *Hub) Register(userID uint, conn *websocket.Conn) {
	h.clients.Store(userID, conn)
}

func (h *Hub) Unregister(userID uint) {
	h.clients.Delete(userID)
}

type ChatMessage struct {
	ID        uint      `json:"id"`
	ChatID    uint      `json:"chat_id"`
	SenderID  uint      `json:"sender_id"`
	Recipient uint      `json:"recipient_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Hub) SendTo(recipient uint, msg ChatMessage) error {
	val, ok := h.clients.Load(recipient)
	if !ok {
		return nil
	}
	conn := val.(*websocket.Conn)
	return conn.WriteJSON(msg)

}
