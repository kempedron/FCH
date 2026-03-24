package handler

import (
	"html/template"
	"net/http"
)

func MakeHandlerForMainPage(tmpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index.html", nil)
	})
}
