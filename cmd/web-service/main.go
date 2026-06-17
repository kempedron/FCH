package main

import (
	"FCH/internal/web-service/handler"
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
	r := mux.NewRouter()
	r.Handle("/", handler.MakeHandlerForMainPage(tmpl))
	if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
