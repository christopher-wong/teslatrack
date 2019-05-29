package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/christopher-wong/teslatrack/ownerapi"
)

func (s *Server) SetTeslaAccountHandler(w http.ResponseWriter, r *http.Request) {
	// These are a user's Tesla creds.
	// NEVER store these, just grab their token.
	teslaCreds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(teslaCreds)
	if err != nil {
		log.Println(err)
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
		json.NewEncoder(w).Encode(err)
		return
	}

	// get the user's email from the JWT
	claims, err := s.GetJWTClaims(r.Header.Get("Authorization"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	userID := claims["user_id"]

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
		ON CONFLICT (user_id)
		DO UPDATE SET
			access_token = EXCLUDED.access_token,
			token_type = EXCLUDED.token_type,
			expires_in = EXCLUDED.expires_in,
			refresh_token = EXCLUDED.refresh_token,
			created_at = EXCLUDED.created_at;
	`
	resp := client.OwnerAPIAuthResponse
	if _, err = s.db.Query(query, userID, resp.AccessToken, resp.TokenType, resp.ExpiresIn, resp.RefreshToken, resp.CreatedAt); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)

		return
	}

	// return the Tesla auth credentials
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
