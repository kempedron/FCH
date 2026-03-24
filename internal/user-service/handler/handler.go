package handler

import (
	"FCH/internal/database"
	"FCH/internal/jwt"
	"FCH/internal/models"
	"FCH/internal/user-service/service"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

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
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user := service.Login(req.Username, req.Password)
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

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

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
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := service.Register(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User already exists", http.StatusConflict)
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

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}

type SearchedUser struct {
	Username string
	ID       uint
}

func MakeHandlerForSearchPage(tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		username := mux.Vars(r)["username"]
		err := database.DB.Where("username=?", username).First(&user).Error
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		data := SearchedUser{
			Username: user.Username,
			ID:       user.ID,
		}
		tmpl.ExecuteTemplate(w, "search.html", data)
	})
}
