package handler

import (
	"log"
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
	log.Printf("Hub: Attempting to send message to UserID %d", recipient)
	val, ok := h.clients.Load(recipient)
	if !ok {
		log.Printf("Hub: User %d not found in active clients map", recipient)
		return nil
	}
	conn := val.(*websocket.Conn)
	err := conn.WriteJSON(msg)
	if err != nil {
		log.Printf("Hub: Error writing to JSON for user %d: %v", recipient, err)
		h.clients.Delete(recipient)
	}
	return err
}

func (h *Hub) BroadcastToChat(particiantsIDs []uint, msg ChatMessage, excludeSender uint) {
	for _, participantID := range particiantsIDs {
		if participantID == excludeSender {
			continue
		}
		go h.SendTo(participantID, msg)
	}
}
