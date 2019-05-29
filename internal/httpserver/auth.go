package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email  string `json:"email"`
	UserID int    `json:"user_id"`
	jwt.StandardClaims
}

func (s *Server) GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := s.db.QueryRow("SELECT password FROM users WHERE email=$1", creds.Email)
	storedCreds := &Credentials{}
	err = result.Scan(&storedCreds.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	// check the password
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get the user's userID from the database
	var userID int
	err = s.db.QueryRow("SELECT id FROM users WHERE email=$1", creds.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	// set expiration
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		Email:  storedCreds.Email,
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// declare token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.cfg.JwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
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

func (s *Server) SignupHandler(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err = s.db.Query("INSERT INTO users (email, password) VALUES ($1, $2)", creds.Email, string(hashedPassword)); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
}

func (s *Server) GetJWTClaims(tk string) (jwt.MapClaims, error) {
	tkReplaced := strings.Replace(tk, "Bearer ", "", -1)

	token, err := jwt.Parse(tkReplaced, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return s.cfg.JwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("failed to validate claims")
}
