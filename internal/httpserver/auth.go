package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Credentials stores the email and password used to login to the Tesla API or
// the teslatrack api
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims is a custom claims object that wraps jwt.StandardClaims
type Claims struct {
	Email  string `json:"email"`
	UserID int32  `json:"user_id"`
	jwt.StandardClaims
}

// calculate when a given token expires and if it's within 7 days of the
// current time, return true.
func needsRefresh(createdAt, expiresIn int64) bool {
	expiresAtTime := time.Unix(createdAt+expiresIn, 0)

	fmt.Println(expiresAtTime)

	return expiresAtTime.After(time.Now()) && expiresAtTime.Before(time.Now().Add(24*7*time.Hour))
}

// look at all the tokens in the database, refresh any that are about to expire.
func (s *Server) TeslaOwnerTokenRefresh() error {
	rows, err := s.db.Query("SELECT expires_in, refresh_token, created_at FROM tesla_auth")
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var expiresIn int64
		var refreshToken string
		var createdAt int64
		err = rows.Scan(&expiresIn, &refreshToken, &createdAt)
		if err != nil {
			log.Println(err)
			return err
		}

		if needsRefresh(createdAt, expiresIn) {
			// hit tesla auth endpoint and pass refresh token
		}
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetTokenHandler creates a JWT after verifying a user's credentials.
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
	var userID int32
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

// SignupHandler creates a new account and hashes the user password
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

// GetJWTClaims retrieves a JWT from a cplains struct.
func (s *Server) GetJWTClaims(tk string) (*Claims, error) {
	tkReplaced := strings.Replace(tk, "Bearer ", "", -1)

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tkReplaced, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return s.cfg.JwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}
