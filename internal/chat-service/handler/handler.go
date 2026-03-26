package handler

import (
	"FCH/internal/chat-service/service"
	"FCH/internal/database"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func MakeHandlerForChat(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		opponentIDStr := vars["userID"]
		opponentID, _ := strconv.ParseUint(opponentIDStr, 10, 32)
		myID, _ := middleware.GetUserIDFromRequest(r)

		var chat models.Chat

		err := database.DB.Table("chats").
			Joins("JOIN chat_particiants p1 ON p1.chat_id = chats.id").
			Joins("JOIN chat_particiants p2 ON p2.chat_id = chats.id").
			Where("p1.user_id = ? AND p2.user_id = ? AND p1.is_group = ?", myID, opponentID, false).
			First(&chat).Error

		if err != nil {
			chat = models.Chat{Name: "Private Chat"}
			database.DB.Create(&chat)

			database.DB.Create(&models.ChatParticipants{ChatID: chat.ID, UserID: uint(myID)})
			database.DB.Create(&models.ChatParticipants{ChatID: chat.ID, UserID: uint(opponentID)})
		}

		database.DB.Preload("Participants.User").Preload("Messages").First(&chat, chat.ID)

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
	valID, _ := middleware.GetUserIDFromRequest(r)
	myID := uint(valID)

	vars := mux.Vars(r)
	chatIDStr := vars["chatID"]
	chatID, _ := strconv.ParseUint(chatIDStr, 10, 32)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	chatHub.Register(myID, conn)

	defer func() {
		chatHub.Unregister(myID)
		conn.Close()
	}()

	for {
		var msg ChatMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		dbMessage := models.Message{
			ChatID:   uint(chatID),
			AuthorID: myID,
			Content:  msg.Content,
		}

		database.DB.Create(&dbMessage)

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

func MakeHandlerForMyChats(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		MyID, _ := middleware.GetUserIDFromRequest(r)

		chats, err := service.GetMyChats(MyID)
		if err != nil {
			http.Error(w, "Failed to get chats", http.StatusInternalServerError)
			return
		}
		data := struct {
			Chats []models.Chat
			MyID  uint
		}{
			Chats: chats,
			MyID:  MyID,
		}
		fmt.Printf("Chats found: %d for UserID: %d\n", len(chats), MyID)
		tmpl.ExecuteTemplate(w, "myChats.html", data)
	}
}
