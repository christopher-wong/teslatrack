package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// GetPctCompletionFreqCount generaetes a frequency count map for all the
// completed charges at given percentages
func (s *Server) GetPctCompletionFreqCount(w http.ResponseWriter, r *http.Request) {
	// get the user's email from the JWT
	claims, err := s.GetJWTClaims(r.Header.Get("Authorization"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	userID := claims.UserID

	pctCompletionFreqCount := map[int32]int32{}

	results, err := s.services.GetCompletedChargeRows(userID)
	if err != nil {
		log.Fatal("failed to fetch completed charge data")
	}

	start := time.Now()
	for _, row := range *results {
		// frequency number of completed charges at each battery level.
		// key = completed charge %
		// value = count
		if _, ok := pctCompletionFreqCount[row.BatteryLevel]; ok {
			pctCompletionFreqCount[row.BatteryLevel]++
		} else {
			pctCompletionFreqCount[row.BatteryLevel] = 1
		}
	}

	elapsed := time.Since(start)
	log.Printf("took %s", elapsed)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pctCompletionFreqCount)
}
