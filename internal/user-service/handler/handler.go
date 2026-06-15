package handler

import (
	"FCH/internal/jwt"
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

type UserHandler struct {
	tmpl        *template.Template
	userService service.UserService
}

func NewUserHandler(tmpl *template.Template, userService service.UserService) *UserHandler {
	return &UserHandler{tmpl: tmpl, userService: userService}
}

func (h *UserHandler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
		"CSRFToken": r.Header.Get("X-CSRF-Token"),
	})
}

func (h *UserHandler) HandlerLoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user := h.userService.Login(req.Username, req.Password)
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

func (h *UserHandler) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
		"CSRFToken": r.Header.Get("X-CSRF-Token"),
	})
}

func (h *UserHandler) HandlerRegisterPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.userService.Register(req.Username, req.Password)
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

func (h *UserHandler) SearchPageHandler(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := h.userService.SearchByUsername(username)
	if err != nil {
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
			return
		}
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	data := SearchedUser{
		Username: user.Username,
		ID:       user.ID,
	}
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}
	h.tmpl.ExecuteTemplate(w, "search.html", data)

}
