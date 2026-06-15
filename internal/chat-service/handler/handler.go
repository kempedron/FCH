package handler

import (
	"FCH/internal/chat-service/service"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type ChatHandler struct {
	tmpl        *template.Template
	chatService service.ChatService
}

func NewChatHandler(tmpl *template.Template, chatService service.ChatService) *ChatHandler {
	return &ChatHandler{tmpl: tmpl, chatService: chatService}
}

func (h *ChatHandler) ChatPageHandler(w http.ResponseWriter, r *http.Request) {
	chatID := uint(middleware.GetParamByUrl("chatID", r))
	myID, _ := middleware.GetUserIDFromRequest(r)

	var buf bytes.Buffer

	chat, err := h.chatService.GetChatByID(chatID, myID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.NewChatPageHandler(w, r)
			return
		}
		http.Error(w, "chat not found", http.StatusNotFound)
		return
	}

	data := map[string]any{
		"Chat": chat,
		"MyID": myID,
	}
	err = h.tmpl.ExecuteTemplate(&buf, "chatPage.html", data)
	if err != nil {
		log.Printf("html rendering error: %s", err)
		http.Error(w, "html render Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf8")
	buf.WriteTo(w)
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

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	valID, _ := middleware.GetUserIDFromRequest(r)
	myID := uint(valID)

	vars := mux.Vars(r)
	chatIDStr := vars["chatID"]
	chatID64, _ := strconv.ParseUint(chatIDStr, 10, 32)
	chatID := uint(chatID64)

	// Замена "грязного" SQL-запроса на вызов метода сервиса
	inChat, err := h.chatService.IsUserInChat(chatID, myID)
	if err != nil || !inChat {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

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

		// Логика записи перенесена в сервис
		dbMessage, err := h.chatService.CreateMessage(chatID, myID, msg.Content)
		if err != nil {
			log.Printf("failed to save message: %v", err)
			break
		}

		broadcastMsg := ChatMessage{
			ID:        dbMessage.ID,
			ChatID:    dbMessage.ChatID,
			SenderID:  dbMessage.AuthorID,
			Content:   dbMessage.Content,
			CreatedAt: dbMessage.CreatedAt,
		}

		participantIDs, err := h.chatService.GetParticipantIDs(chatID)
		if err != nil {
			log.Printf("failed to get participants: %v", err)
			continue
		}

		chatHub.BroadcastToChat(participantIDs, broadcastMsg, myID)
		conn.WriteJSON(broadcastMsg)
	}
}

func (h *ChatHandler) MyChatsPageHandler(w http.ResponseWriter, r *http.Request) {
	myID, _ := middleware.GetUserIDFromRequest(r)

	chats, err := h.chatService.GetMyChats(myID)
	if err != nil {
		http.Error(w, "Failed to get chats", http.StatusInternalServerError)
		return
	}
	data := struct {
		Chats []models.Chat
		MyID  uint
	}{
		Chats: chats,
		MyID:  myID,
	}
	h.tmpl.ExecuteTemplate(w, "myChats.html", data)
}

func (h *ChatHandler) NewChatPageHandler(w http.ResponseWriter, r *http.Request) {
	opponentID := uint(middleware.GetParamByUrl("userID", r))
	myID, _ := middleware.GetUserIDFromRequest(r)

	if exist, chat := h.chatService.IsChatExist(myID, opponentID); exist {
		http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(chat.ID), 10), http.StatusFound)
		return
	}

	chat, err := h.chatService.GetOrCreatePersonalChat(myID, opponentID)
	if err != nil {
		http.Error(w, "Failed to get chat", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(chat.ID), 10), http.StatusFound)
}

func (h *ChatHandler) MakeHandlerForCreateNewGroupChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		csrfToken := r.Header.Get("X-CSRF-Token")
		data := map[string]string{
			"CSRFToken": csrfToken,
		}
		h.tmpl.ExecuteTemplate(w, "create_group.html", data)
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

	groupID, err := h.chatService.CreateGroupChat(groupName, userID)
	if err != nil {
		http.Error(w, "Не удалось создать группу", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(groupID), 10), http.StatusFound)
}

func (h *ChatHandler) JoinToGroupChat(w http.ResponseWriter, r *http.Request) {
	inviteCode := r.URL.Query().Get("code")
	groupChatID := uint(middleware.GetParamByUrl("groupID", r))
	userID, err := middleware.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = h.chatService.JoinToGroupChat(groupChatID, inviteCode, userID)
	if err != nil {
		log.Printf("join group error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/chat/"+strconv.FormatUint(uint64(groupChatID), 10), http.StatusFound)
}
