package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"teslatrack/services"
)

// Server is the translations server
type Server struct {
	cfg      *Config
	router   *mux.Router
	db       *sql.DB
	services *services.ServicesClient
}

// Config is the server configuration
type Config struct {
	ListenAddress string
	Dev           bool
	JwtKey        []byte
}

// Run kicks off background tasks, then begins serving http
// requests. Returns only when the underlying http server dies.
func (s *Server) Run() error {
	c := cors.New(cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(s.router)

	log.Printf("API server listening on %s", s.cfg.ListenAddress)

	s.TeslaOwnerTokenRefresh()

	return http.ListenAndServe(s.cfg.ListenAddress, handlers.LoggingHandler(os.Stdout, handler))
}

// New returns a configured glass server.
func New(cfg *Config, db *sql.DB, services *services.ServicesClient) (*Server, error) {
	// Default to port 8001.
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = "0.0.0.0:8001"
	}

	srv := &Server{
		cfg:      cfg,
		router:   mux.NewRouter(),
		db:       db,
		services: services,
	}

	srv.router.HandleFunc("/user/auth/token", srv.GetTokenHandler).Methods("POST")
	srv.router.HandleFunc("/user/auth/signup", srv.SignupHandler).Methods("POST")

	// protected routes
	srv.router.Handle("/user/tesla-account", wrapAuthHandler(cfg.JwtKey, srv.SetTeslaAccountHandler)).Methods("POST")
	srv.router.Handle("/vehicle/basic-summary", wrapAuthHandler(cfg.JwtKey, srv.GetVehicleBasicSummary)).Methods("GET")
	srv.router.Handle("/vehicle/charging/sessions", wrapAuthHandler(cfg.JwtKey, srv.GetChargingSessionDetails)).Methods("GET")

	srv.router.Handle("/vehicle/freqcount", wrapAuthHandler(cfg.JwtKey, srv.GetPctCompletionFreqCount)).Methods("GET")

	return srv, nil
}

func wrapAuthHandler(jwtKey []byte, f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	return negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(http.HandlerFunc(f)),
	)
}
