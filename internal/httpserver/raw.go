package server

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type PropertyMap map[string]interface{}

func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (a *PropertyMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type RawResult struct {
	Timestamp time.Time
	Data      PropertyMap
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
