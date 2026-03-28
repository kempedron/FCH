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
	"gorm.io/gorm"
)

func MakeHandlerForChat(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := uint(middleware.GetParamByUrl("chatID", r))
		myID, _ := middleware.GetUserIDFromRequest(r)

		var chat models.Chat

		err := database.DB.Table("chats").
			Preload("Participants.User").
			Preload("Messages", func(db *gorm.DB) *gorm.DB {
				return db.Order("created_at ASC")
			}).
			First(&chat, uint(chatID)).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				MakeHanlerForNewChat(tmpl)(w, r)
				return
			}
			http.Error(w, "chat not found", http.StatusNotFound)
			return
		}

		// isPartician := false

		// for _, p := range chat.Participants {
		// 	if p.UserID == uint(myID) {
		// 		isPartician = true
		// 		break
		// 	}
		// }
		// if !isPartician {
		// 	http.Error(w, "Access denied", http.StatusForbidden)
		// 	return
		// }

		data := map[string]any{
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

	var participants []models.ChatParticipants
	database.DB.Where("chat_id = ? AND deleted_at IS NULL", chatID).Find(&participants)

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
		}

		participantIDs := make([]uint, len(participants))
		for i, p := range participants {
			participantIDs[i] = p.UserID
		}
		chatHub.BroadcastToChat(participantIDs, broadcastMsg, myID)

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

func MakeHanlerForNewChat(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opponentID := uint(middleware.GetParamByUrl("userID", r))
		myID, _ := middleware.GetUserIDFromRequest(r)

		if exist, chat := service.IsChatExist(myID, opponentID); exist {
			http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(chat.ID), 10), http.StatusFound)
		}

		chat, err := service.GetOrCreatePersonalChat(uint(myID), uint(opponentID))
		if err != nil {
			http.Error(w, "Failed to get chat", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(chat.ID), 10), http.StatusFound)
	}
}
