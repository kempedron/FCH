package main

import (
	"FCH/internal/middleware"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	ChatServiceURL string
	UserServiceURl string
	WebServiceURl  string
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
		"user-service": g.config.UserServiceURl,
		"web-service":  g.config.WebServiceURl,
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
			http.Error(w, "Service unavaible", http.StatusInternalServerError)
			return
		}

		servis.proxy.ServeHTTP(w, r)
	}
}

func (g *APIGateway) proxyToChatService(w http.ResponseWriter, r *http.Request) {
	g.proxyToService("chat-service")(w, r)
}

func (g *APIGateway) proxyToUserService(w http.ResponseWriter, r *http.Request) {
	g.proxyToService("user-service")(w, r)
}

func (g *APIGateway) proxyToWebService(w http.ResponseWriter, r *http.Request) {
	g.proxyToService("web-service")(w, r)
}

func (g *APIGateway) setRoutes() {
	g.router.HandleFunc("/login", g.proxyToUserService).Methods("GET")
	g.router.HandleFunc("/register", g.proxyToUserService).Methods("GET")
	g.router.HandleFunc("/login", g.proxyToUserService).Methods("POST")
	g.router.HandleFunc("/register", g.proxyToUserService).Methods("POST")
	g.router.HandleFunc("/search/{username}", g.proxyToUserService).Methods("GET")

	protected := g.router.PathPrefix("/").Subrouter()
	protected.HandleFunc("/", g.proxyToWebService).Methods("GET")
	protected.HandleFunc("/chat/{userID}", g.proxyToChatService).Methods("GET")

	protected.Use(middleware.JWTAuth)

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
		UserServiceURl: os.Getenv("USER_SERVICE_URL"),
		WebServiceURl:  os.Getenv("WEB_SERVICE_URL"),
	}

	gateway := NewAPIGateway(cfg)

	log.Printf("API Gateway starting on port %s <-", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, gateway.router); err != nil {
		log.Fatalf("Failed to start API Gateway: %s", err)
	}
}
