package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/christopher-wong/teslatrack/internal/google"
)

var AGlobalMapForJohn = make(map[string]string)

type ChargingSessionDetailsQueryRow struct {
	Timestamp     time.Time
	ChargingState string
	ChargeState   interface{}
	Latitude      string
	Longitude     string
	Address       string
}

func (s *Server) GetChargingSessionDetails(w http.ResponseWriter, r *http.Request) {
	// get the user's email from the JWT
	claims, err := s.GetJWTClaims(r.Header.Get("Authorization"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID := claims["user_id"]

	fmt.Println(userID)

	query := `
		SELECT w1.timestamp,
			w1.charging_state,
			w1.charge_state,
			w1.latitude,
			w1.longitude
		FROM (SELECT w2.timestamp,
					w2.data -> 'response' -> 'drive_state' ->> 'latitude'        as latitude,
					w2.data -> 'response' -> 'drive_state' ->> 'longitude'       as longitude,
					w2.data -> 'response' -> 'charge_state'                      as charge_state,
					w2.data -> 'response' -> 'charge_state' ->> 'charging_state' as charging_state,
					lead(w2.data -> 'response' -> 'charge_state' ->> 'charging_state')
					OVER (ORDER BY w2.timestamp DESC)                            as prev_charging_state
			FROM state w2
			WHERE user_id = $1
			ORDER BY w2.timestamp DESC) as w1
		WHERE w1.charging_state IS DISTINCT FROM w1.prev_charging_state
		ORDER BY w1.timestamp DESC;
	`

	var charges []ChargingSessionDetailsQueryRow

	rows, err := s.db.Query(query, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var rowStuct ChargingSessionDetailsQueryRow
		err := rows.Scan(&rowStuct.Timestamp, &rowStuct.ChargingState, &rowStuct.ChargeState, &rowStuct.Latitude, &rowStuct.Longitude)
		if err != nil {
			log.Fatal(err)
		}
		charges = append(charges, rowStuct)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// convert all latlongs to formatted addresses
	var chargesWithAddress []ChargingSessionDetailsQueryRow
	for _, c := range charges {

		latLongString := fmt.Sprintf("%s,%s", c.Latitude, c.Longitude)

		// if latlong string in cache, grab it, if not, insert it
		if val, ok := AGlobalMapForJohn[latLongString]; ok {
			c.Address = val
		} else {
			addr := google.ReverseGeocode(&http.Client{}, c.Latitude, c.Longitude)
			AGlobalMapForJohn[latLongString] = addr
			c.Address = addr
		}

		chargesWithAddress = append(chargesWithAddress, c)
	}

	// count charges at each address
	// addrCounter := make(map[string]int)
	// for _, row := range chargesWithAddress {
	// 	if row.ChargingState == "Charging" {
	// 		addrCounter[row.Address]++
	// 	}
	// }

	// fmt.Println(addrCounter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Data []ChargingSessionDetailsQueryRow `json:"data"`
	}{
		Data: chargesWithAddress,
	})
}
