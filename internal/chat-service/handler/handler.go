package handler

import (
	"FCH/internal/chat-service/service"
	"FCH/internal/database"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slices"
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
			Where("id IN (?)", database.DB.Table("chat_participants").
				Select("chat_id").
				Where("user_id = ? AND deleted_at IS NULL", myID),
			).
			First(&chat, uint(chatID)).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				MakeHanlerForNewChat(tmpl)(w, r)
				return
			}
			http.Error(w, "chat not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Chat": chat,
			"MyID": myID,
		}
		tmpl.ExecuteTemplate(w, "chatPage.html", data)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:8080",
			"http://127.0.0.1:8080",
		}
		return slices.Contains(allowedOrigins, origin)
	},
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

		var count int64
		database.DB.Model(&models.ChatParticipants{}).
			Where("chat_id = ? AND user_id = ? AND deleted_at IS NULL", chatID, myID).
			Count(&count)
		if count == 0 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
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

		var participants []models.ChatParticipants
		database.DB.Where("chat_id = ? AND deleted_at IS NULL", chatID).Find(&participants)

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
		fmt.Println("OPPONENT ID: -> ", opponentID, " <- !!!")

		if exist, chat := service.IsChatExist(myID, opponentID); exist {
			http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(chat.ID), 10), http.StatusFound)
			return
		}

		chat, err := service.GetOrCreatePersonalChat(uint(myID), uint(opponentID))
		if err != nil {
			http.Error(w, "Failed to get chat", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(chat.ID), 10), http.StatusFound)
	}

}

// handle GET and POST methods
func MakeHanderForCreateNewGroupChat(tmp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {

			csrfToken := r.Header.Get("X-CSRF-Token")

			data := map[string]string{
				"CSRFToken": csrfToken,
			}

			tmp.ExecuteTemplate(w, "create_group.html", data)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
			return
		}

		groupName := r.FormValue("group_name")

		userID, err := middleware.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		var groupID uint

		err = database.DB.Transaction(func(tx *gorm.DB) error {
			newChat := models.Chat{
				Name:       groupName,
				IsGroup:    true,
				CreatorID:  userID,
				InviteCode: service.GenerateInviteCode(),
			}

			if err := tx.Create(&newChat).Error; err != nil {
				return err
			}

			participant := models.ChatParticipants{
				ChatID: newChat.ID,
				UserID: userID,
				Role:   "admin",
			}

			if err := tx.Create(&participant).Error; err != nil {
				return err
			}

			groupID = newChat.ID

			return nil
		})

		if err != nil {
			http.Error(w, "Не удалось создать группу", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(groupID), 10), http.StatusFound)
	}
}

func JoinToGroupChat(w http.ResponseWriter, r *http.Request) {
	invite_code := r.URL.Query().Get("code")
	groupChatID := uint(middleware.GetParamByUrl("groupID", r))
	userID, err := middleware.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var chat models.Chat

	err = database.DB.Where("id = ? AND invite_code = ? AND is_group = ? AND deleted_at IS NULL", groupChatID, invite_code, true).First(&chat).Error
	if err != nil {
		http.Error(w, "invalid or exparid invite link", http.StatusNotFound)
		return
	}

	err = service.AddUserToGroupChat(userID, groupChatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			log.Printf("error: %s", err)
			http.Error(w, "Failed to join", http.StatusInternalServerError)
		}
		return

	}
	http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(groupChatID), 10), http.StatusFound)
}
