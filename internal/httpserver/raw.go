package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type RawResult struct {
	Timestamp time.Time
	Data      map[string]interface{}
}

// GetLatestRawEntries retrieves the latest raw entries.
func (s *Server) GetLatestRawEntries(w http.ResponseWriter, r *http.Request) {
	// get the user's email from the JWT
	claims, err := s.GetJWTClaims(r.Header.Get("Authorization"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	userID := claims.UserID

	query := `
		SELECT timestamp, data
		FROM state
		WHERE user_id = $1
		ORDER BY timestamp DESC
		limit 10;
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	defer rows.Close()

	var data []RawResult

	for rows.Next() {
		var row RawResult
		err = rows.Scan(&row.Timestamp, &row.Data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}

		data = append(data, row)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Data []RawResult `json:"data"`
	}{
		Data: data,
	})
}
