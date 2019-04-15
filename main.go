package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/christopher-wong/teslatrack/ownerapi"
	"github.com/codegangsta/negroni"
	"github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	jwt "github.com/dgrijalva/jwt-go"

	_ "github.com/lib/pq"
)

var db *sql.DB
var jwtKey = []byte("my_secret_key")

const (
	host     = "teslatrack-stage-do-user-2432224-0.db.ondigitalocean.com"
	port     = 25060
	user     = "doadmin"
	password = "a3uot0pp9bxzcxoa"
	dbname   = "defaultdb"

	redisHost     = "157.230.142.205:6379"
	redisPassword = "FkNU6btkbjp+RwIG9529yJZG+EfNboVHEC6FzhpifbNMC0fIPC/MJP0/kvo3GYuT7LgkhGDVfE1gEDch"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func runAPI() {
	// connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword, // no password set
		DB:       0,             // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("failed to connect to redis!")
		panic(err)
	}

	r := mux.NewRouter()

	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	// TODO: swap out router for Mux or Gin
	r.HandleFunc("/user/auth/token", GetTokenHandler).Methods("POST")
	r.HandleFunc("/user/auth/signup", SignupHandler).Methods("POST")
	r.Handle("/user/tesla-account", negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(http.HandlerFunc(SetTeslaAccountHandler)),
	)).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", handlers.LoggingHandler(os.Stdout, r)))
}

func main() {
	dbinit()

	// start API server
	go runAPI()

	// stop main thread from executing
	select {}
}

func SetTeslaAccountHandler(w http.ResponseWriter, r *http.Request) {
	// These are a user's Tesla creds.
	// NEVER store these, just grab their token.
	teslaCreds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(teslaCreds)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := &ownerapi.GetAuthTokenInput{
		Email:    teslaCreds.Email,
		Password: teslaCreds.Password,
	}
	// create an ownerapi client and auth to Tesla
	client, err := ownerapi.NewClient(&http.Client{}, input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// get the user's email from the JWT
	claims, err := GetJWTClaims(r.Header.Get("Authorization"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	userEmail := claims["email"]

	// get the user's userID from the database
	var userID int
	err = db.QueryRow("SELECT id FROM user WHERE email=$1", userEmail).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// write their tesla auth object to the database
	query := `
		INSERT INTO tesla_auth (
			user_id,
			access_token,
			token_type,
			expires_in,
			refresh_token,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	resp := client.OwnerAPIAuthResponse
	if _, err = db.Query(query, userID, resp.AccessToken, resp.TokenType, resp.ExpiresIn, resp.RefreshToken, resp.CreatedAt); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// return the Tesla auth credentials
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
