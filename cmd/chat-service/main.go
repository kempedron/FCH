package main

import (
	"FCH/internal/chat-service/handler"
	"FCH/internal/chat-service/service"
	"FCH/internal/database"
	"FCH/internal/repository" // Импортируем ваши репозитории
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func InitTemplates() *template.Template {
	funcMap := template.FuncMap{
		"firstChar": func(s string) string {
			if len(s) == 0 {
				return "?"
			}
			return string([]rune(s)[0])
		},
		"not": func(v any) bool { return v == nil },
	}
	return template.Must(
		template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html"),
	)
}

func main() {
	r := mux.NewRouter()

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	tmpl := InitTemplates()

	// REPOS
	chatRepo := repository.NewChatRepository(db)
	partRepo := repository.NewParticipantRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	chatService := service.NewService(chatRepo, partRepo, msgRepo)

	chatHandler := handler.NewChatHandler(tmpl, chatService)

	r.HandleFunc("/chat/{chatID:[0-9]+}", chatHandler.ChatPageHandler)
	r.HandleFunc("/chat/{chatID}/send", chatHandler.SendMessage)
	r.HandleFunc("/start-chat/{userID:[0-9]+}", chatHandler.NewChatPageHandler)
	r.HandleFunc("/group-chat/join-to-chat/{groupID}", chatHandler.JoinToGroupChat)
	r.HandleFunc("/group-chat/create", chatHandler.MakeHandlerForCreateNewGroupChat)
	r.HandleFunc("/my-chats", chatHandler.MyChatsPageHandler)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("DEBUG: Chat Service got %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	log.Println("Сервер чатов успешно запущен на порту :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
