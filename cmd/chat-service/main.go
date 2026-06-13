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
		"multiply": func(a, b uint) uint { return a * b },
		"mul":      func(a, b int) int { return a * b },
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

	// 1. Инициализируем БД всего один раз
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	tmpl := InitTemplates()

	// 2. Создаем слой репозиториев
	chatRepo := repository.NewChatRepository(db)
	userRepo := repository.NewRepository(db)
	partRepo := repository.NewParticipantRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	// 3. Создаем сервис и передаем ему репозитории
	chatService := service.NewService(chatRepo, userRepo, partRepo, msgRepo)

	// 4. Инициализируем структуру хэндлеров, передавая шаблон и готовый сервис
	chatHandler := handler.NewChatHandler(tmpl, chatService)

	// 5. Регистрируем маршруты как МЕТОДЫ созданной структуры chatHandler
	r.HandleFunc("/chat/{chatID:[0-9]+}", chatHandler.MakeHandlerForChat())
	r.HandleFunc("/chat/{chatID}/send", chatHandler.SendMessage)
	r.HandleFunc("/start-chat/{userID:[0-9]+}", chatHandler.MakeHandlerForNewChat())
	r.HandleFunc("/group-chat/join-to-chat/{groupID}", chatHandler.JoinToGroupChat)
	r.HandleFunc("/group-chat/create", chatHandler.MakeHandlerForCreateNewGroupChat())
	r.HandleFunc("/my-chats", chatHandler.MakeHandlerForMyChats())

	// Middleware для отладки
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("DEBUG: Chat Service got %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	log.Println("🚀 Сервер чатов успешно запущен на порту :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
