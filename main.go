package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"teslad/ownerapi"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/codegangsta/negroni"
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
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func main() {
	dbinit()

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

	client := ownerapi.Client{
		HttpClient: &http.Client{},
	}

	input := &ownerapi.GetAuthTokenInput{
		Email:    teslaCreds.Email,
		Password: teslaCreds.Password,
	}
	resp, err := client.GetAuthToken(input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	claims, err := GetJWTClaims(r.Header.Get("Authorization"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Println(claims)
	fmt.Println(resp)
}
