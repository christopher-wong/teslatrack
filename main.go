package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	jwt "github.com/dgrijalva/jwt-go"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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

func initDB() {
	var err error

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

func main() {
	r := mux.NewRouter()

	// TODO: swap out router for Mux or Gin
	r.HandleFunc("/user/auth/token", GetTokenHandler).Methods("POST")
	r.HandleFunc("/signup", SignupHandler).Methods("POST")

	initDB()

	log.Fatal(http.ListenAndServe(":8000", handlers.LoggingHandler(os.Stdout, r)))
}

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := db.QueryRow("SELECT password FROM USERS WHERE email=$1", creds.Email)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	storedCreds := &Credentials{}
	err = result.Scan(&storedCreds.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check the password
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
	}

	// set expiration
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		Email: creds.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// declare token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tr := struct {
		Name    string `json:"name"`
		Value   string `json:"value"`
		Expires int64  `json:"expires"`
		Type    string `json:"type"`
	}{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime.Unix(),
		Type:    "bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tr)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err = db.Query("INSERT INTO users (email, password) VALUES ($1, $2)", creds.Email, string(hashedPassword)); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
