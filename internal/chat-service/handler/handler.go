package handler

import (
	"FCH/internal/chat-service/service"
	"FCH/internal/middleware"
	"html/template"
	"net/http"
)

func MakeHandlerForChat(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetParamByUrl("userID", r)
		myID, err := middleware.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		chat, err := service.GetChatInfo(middleware.GetUsernameByID(uint(userID)), uint(myID))
		if err != nil {
			http.Error(w, "chat not found", http.StatusNotFound)
			return
		}
		tmpl.ExecuteTemplate(w, "chatPage.html", chat)
	}
}
