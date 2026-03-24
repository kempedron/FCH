package handler

import (
	"FCH/internal/jwt"
	"FCH/internal/user-service/service"
	"html/template"
	"net/http"
	"time"
)

func MakeHandlerForLoginPage(tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "login.html", nil)
	})
}

func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	user := service.Login(username, password)
	if user == nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func MakeHandlerForRegisterPage(tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "register.html", nil)
	})
}

func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := service.Register(username, password)
	if user == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)

}
