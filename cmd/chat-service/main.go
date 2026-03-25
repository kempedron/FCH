package main

import (
	"FCH/internal/chat-service/handler"
	"FCH/internal/database"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func InitTemplates() *template.Template {
	funcMap := template.FuncMap{
		"multiply": func(a, b uint) uint {
			return a * b
		},
		"firstChar": func(s string) string {
			if len(s) == 0 {
				return "?"
			}
			return string([]rune(s)[0])
		},
		"not": func(v any) bool {
			return v == nil
		},
	}
	return template.Must(
		template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html"),
	)
}

func main() {

	r := mux.NewRouter()

	err := database.InitDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	tmpl := InitTemplates()
	r.HandleFunc("/chat/{userID}", handler.MakeHandlerForChat(tmpl))
	r.HandleFunc("/chat/{chatID}/send", handler.SendMessage)
	///chat/{chatID}/send
	http.ListenAndServe(":8080", r)
}
