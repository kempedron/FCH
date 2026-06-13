package main

import (
	"FCH/internal/database"
	"FCH/internal/user-service/handler"
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
		"mul": func(a, b int) int {
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
	tmpl := InitTemplates()
	_, err := database.InitDB()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/login", handler.HandlerLogin).Methods("POST")
	r.HandleFunc("/login", handler.MakeHandlerForLoginPage(tmpl)).Methods("GET")

	r.HandleFunc("/register", handler.HandlerRegister).Methods("POST")
	r.HandleFunc("/register", handler.MakeHandlerForRegisterPage(tmpl)).Methods("GET")

	r.HandleFunc("/search/{username}", handler.MakeHandlerForSearchPage(tmpl)).Methods("GET")

	if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
