package handler

import (
	"FCH/internal/database"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func MakeHandlerForChat(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		chatID := vars["userID"]
		myID, _ := middleware.GetUserIDFromRequest(r)

		var chat models.Chat
		database.DB.Preload("Participants.User").Preload("Messages").First(&chat, chatID)

		data := map[string]interface{}{
			"Chat": chat,
			"MyID": myID,
		}
		tmpl.ExecuteTemplate(w, "chatPage.html", data)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
	myID, _ := middleware.GetUserIDFromRequest(r)

	vars := mux.Vars(r)
	chatIDStr := vars["chatID"]
	chatID, _ := strconv.ParseUint(chatIDStr, 10, 32)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	chatHub.Register(uint(myID), conn)

	defer func() {
		chatHub.Unregister(uint(myID))
		conn.Close()
	}()

	for {

		var msg ChatMessage

		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Read error: %s", err)
			break
		}
		dbMessage := models.Message{
			ChatID:   uint(chatID),
			AuthorID: uint(myID),
			Content:  msg.Content,
		}
		if err := database.DB.Create(&dbMessage).Error; err != nil {
			log.Printf("error save to DB: %s", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			continue
		}
		broadcastMsg := ChatMessage{
			ID:        dbMessage.ID,
			ChatID:    dbMessage.ChatID,
			SenderID:  dbMessage.AuthorID,
			Content:   dbMessage.Content,
			CreatedAt: dbMessage.CreatedAt,
			Recipient: msg.Recipient,
		}
		if broadcastMsg.Recipient != 0 {
			chatHub.SendTo(broadcastMsg.Recipient, broadcastMsg)
		}
		conn.WriteJSON(broadcastMsg)

	}

}
