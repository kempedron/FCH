package main

import (
	"FCH/internal/database"
	"FCH/internal/repository"
	"FCH/internal/user-service/handler"
	"FCH/internal/user-service/service"
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
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)

	userService := service.NewService(userRepo)
	userHandler := handler.NewUserHandler(tmpl, userService)

	r := mux.NewRouter()
	r.HandleFunc("/login", userHandler.HandlerLoginPost).Methods("POST")
	r.HandleFunc("/login", userHandler.LoginPageHandler).Methods("GET")

	r.HandleFunc("/register", userHandler.HandlerRegisterPost).Methods("POST")
	r.HandleFunc("/register", userHandler.RegisterPageHandler).Methods("GET")

	r.HandleFunc("/search/{username}", userHandler.SearchPageHandler).Methods("GET")

	if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
