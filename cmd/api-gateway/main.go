package main

import (
	"FCH/internal/middleware"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	_ "FCH/docs/api-gateway"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/koding/websocketproxy"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           FCH API
// @version         1.0
// @description     API Gateway for FCH messenger. All protected routes require JWT cookie.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey JWT
// @in cookie
// @name jwt

type Config struct {
	Port           string
	ChatServiceURL string
	UserServiceURL string
	WebServiceURL  string
}

type ServiceProxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

type APIGateway struct {
	config   *Config
	router   *mux.Router
	services map[string]*ServiceProxy
}

func (g *APIGateway) initService() {
	services := map[string]string{
		"chat-service": g.config.ChatServiceURL,
		"user-service": g.config.UserServiceURL,
		"web-service":  g.config.WebServiceURL,
	}
	g.services = make(map[string]*ServiceProxy)

	for name, serviceURL := range services {
		target, err := url.Parse(serviceURL)
		if err != nil {
			log.Fatalf("failed to parse %s service url: %v", name, err)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		g.services[name] = &ServiceProxy{
			target: target,
			proxy:  proxy,
		}
	}
}

func (g *APIGateway) proxyToService(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("proxing request to %s service: %s", serviceName, r.URL.Path)
		servis, exist := g.services[serviceName]
		if !exist {
			http.Error(w, "Service unavailable", http.StatusInternalServerError)
			return
		}

		if r.Method == http.MethodPost && strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err == nil {
				bodyStr := r.PostForm.Encode()
				r.Body = io.NopCloser(strings.NewReader(bodyStr))
				r.ContentLength = int64(len(bodyStr))
			}
		}

		if token := csrf.Token(r); token != "" {
			r.Header.Set("X-CSRF-Token", token)
		}
		r.Host = servis.target.Host

		servis.proxy.ServeHTTP(w, r)
	}
}

// @Summary      Chat page
// @Description  Renders chat page for a specific chat
// @Tags         chat
// @Produce      html
// @Param        userID path int true "Chat ID"
// @Success      200 {string} string "Chat page"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Router       /chat/{userID} [get]
//
// @Summary      My chats list
// @Description  Renders list of user's chats
// @Tags         chat
// @Produce      html
// @Success      200 {string} string "My chats page"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Router       /my-chats [get]
//
// @Summary      Start personal chat
// @Description  Creates or retrieves a personal chat with another user
// @Tags         chat
// @Produce      html
// @Param        userID path int true "Target user ID"
// @Success      302 "Redirect to /chat/{chatID}"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Router       /start-chat/{userID} [get]
//
// @Summary      Create group chat form
// @Description  Group chat creation page (GET) or submission (POST)
// @Tags         chat
// @Produce      html
// @Param        group_name formData string false "Group name (POST only)"
// @Success      200 {string} string "Group create form"
// @Success      302 "Redirect to /chat/{groupID} after creation"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Router       /group-chat/create [get]
// @Router       /group-chat/create [post]
//
// @Summary      Join group chat
// @Description  Join a group chat via invite code
// @Tags         chat
// @Produce      html
// @Param        groupID path int true "Group chat ID"
// @Param        code query string true "Invite code"
// @Success      302 "Redirect to /chat/{groupID}"
// @Failure      400 "Invalid or expired invite code"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Router       /group-chat/join-to-chat/{groupID} [get]
//
// @Summary      WebSocket chat connection
// @Description  Upgrades to WebSocket for real-time messaging
// @Tags         chat
// @Param        chatID path int true "Chat ID"
// @Success      101 "Switching protocols to WebSocket"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Failure      403 "Access denied — user not in chat"
// @Router       /chat/{chatID}/send [get]
func (g *APIGateway) proxyToChatService(w http.ResponseWriter, r *http.Request) {
	g.proxyToService("chat-service")(w, r)
}

// @Summary      Login form
// @Description  Login page (GET) or authentication (POST)
// @Tags         user
// @Produce      html
// @Param        username formData string false "Username (POST only)"
// @Param        password formData string false "Password (POST only)"
// @Success      200 {string} string "Login page"
// @Success      200 {object} map[string]string "Login success: {\"status\":\"ok\"}"
// @Failure      400 "Invalid request body"
// @Failure      401 "Invalid username or password"
// @Router       /login [get]
// @Router       /login [post]
//
// @Summary      Register form
// @Description  Registration page (GET) or submission (POST)
// @Tags         user
// @Produce      html
// @Param        username formData string false "Username (POST only)"
// @Param        password formData string false "Password (POST only)"
// @Success      200 {string} string "Register page"
// @Success      200 {object} map[string]string "Register success: {\"status\":\"ok\"}"
// @Failure      400 "Invalid request body"
// @Failure      409 "User already exists"
// @Router       /register [get]
// @Router       /register [post]
//
// @Summary      Search users
// @Description  Search users by username
// @Tags         user
// @Produce      html
// @Param        username path string true "Username to search"
// @Success      200 {string} string "Search results page"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Failure      404 "User not found"
// @Router       /search/{username} [get]
func (g *APIGateway) proxyToUserService(w http.ResponseWriter, r *http.Request) {
	g.proxyToService("user-service")(w, r)
}

// @Summary      Main page
// @Description  Serves main search page
// @Tags         web
// @Produce      html
// @Success      200 {string} string "Main page"
// @Failure      401 "Unauthorized — missing or invalid JWT"
// @Router       / [get]
func (g *APIGateway) proxyToWebService(w http.ResponseWriter, r *http.Request) {
	g.proxyToService("web-service")(w, r)
}
func (g *APIGateway) proxyWebSocket(serviceName string) http.HandlerFunc {
	target, _ := url.Parse(g.config.ChatServiceURL)

	return func(w http.ResponseWriter, r *http.Request) {

		wsTarget := *target

		wsTarget.Scheme = "ws"
		if target.Scheme == "https" {
			wsTarget.Scheme = "wss"
		}

		wsTarget.Path = r.URL.Path

		log.Printf("Proxying WebSocket to: %s", wsTarget.String())

		proxy := websocketproxy.NewProxy(&wsTarget)

		proxy.ServeHTTP(w, r)
	}
}

func (g *APIGateway) setRoutes() {
	csrfMiddleware := csrf.Protect(
		[]byte(os.Getenv("CSRF_KEY")),
		csrf.Secure(false),
		csrf.Path("/"),
		csrf.TrustedOrigins([]string{"localhost:8080", "127.0.0.1:8080"}),
		csrf.RequestHeader("X-CSRF-Token"),
	)

	authRouter := g.router.PathPrefix("/").Subrouter()
	authRouter.Use(csrfMiddleware)

	authRouter.HandleFunc("/login", g.proxyToUserService).Methods("GET", "POST")
	authRouter.HandleFunc("/register", g.proxyToUserService).Methods("GET", "POST")

	g.router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	g.router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	protected := g.router.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTAuth)
	protected.Use(csrfMiddleware)

	protected.HandleFunc("/", g.proxyToWebService).Methods("GET")
	protected.HandleFunc("/my-chats", g.proxyToChatService).Methods("GET")
	protected.HandleFunc("/search/{username}", g.proxyToUserService).Methods("GET")
	protected.HandleFunc("/chat/{userID}", g.proxyToChatService).Methods("GET")
	protected.HandleFunc("/start-chat/{userID}", g.proxyToChatService).Methods("GET")
	protected.HandleFunc("/chat/{chatID}/send", g.proxyWebSocket("chat-service"))
	protected.HandleFunc("/group-chat/join-to-chat/{groupID}", g.proxyToChatService).Methods("GET")
	protected.HandleFunc("/group-chat/create", g.proxyToChatService)
}

func NewAPIGateway(cfg *Config) *APIGateway {
	router := mux.NewRouter()
	gateway := &APIGateway{
		config: cfg,
		router: router,
	}
	gateway.initService()
	gateway.setRoutes()
	return gateway
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment")
	}

	cfg := &Config{
		Port:           os.Getenv("PORT"),
		ChatServiceURL: os.Getenv("CHAT_SERVICE_URL"),
		UserServiceURL: os.Getenv("USER_SERVICE_URL"),
		WebServiceURL:  os.Getenv("WEB_SERVICE_URL"),
	}

	gateway := NewAPIGateway(cfg)

	log.Printf("API Gateway starting on port %s <-", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, gateway.router); err != nil {
		log.Fatalf("Failed to start API Gateway: %s", err)
	}
}
