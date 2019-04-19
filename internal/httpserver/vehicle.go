package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type VehicleBasicSummary struct {
	Timestamp       time.Time `json:"timestamp"`
	DisplayName     string    `json:"display_name"`
	Odometer        float64   `json:"odometer"`
	ChargingState   string    `json:"charging_state"`
	BatteryLevel    int       `json:"battery_level"`
	BatteryRange    float32   `json:"battery_range"`
	EstBatteryRange float32   `json:"est_battery_range"`
	InsideTemp      float32   `json:"inside_temp"`
	OutsideTemp     float32   `json:"outside_temp"`
}

func (s *Server) GetVehicleBasicSummary(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT timestamp,
       data -> 'response' ->> 'display_name'                        AS display_name,
       data -> 'response' -> 'vehicle_state' ->> 'odometer'         AS odometer,
       data -> 'response' -> 'charge_state' ->> 'charging_state'    AS charging_state,
       data -> 'response' -> 'charge_state' ->> 'battery_level'     AS battery_level,
       data -> 'response' -> 'charge_state' ->> 'battery_range'     AS battery_range,
       data -> 'response' -> 'charge_state' ->> 'est_battery_range' AS est_battery_range,
       data -> 'response' -> 'climate_state' ->> 'inside_temp'      AS inside_temp,
       data -> 'response' -> 'climate_state' ->> 'outside_temp'     AS outside_temp
	FROM state
	WHERE user_id = 1
	ORDER BY timestamp DESC
	LIMIT 1;
	`

	result := s.db.QueryRow(query)

	obj := &VehicleBasicSummary{}
	err := result.Scan(
		&obj.Timestamp,
		&obj.DisplayName,
		&obj.Odometer,
		&obj.ChargingState,
		&obj.BatteryLevel,
		&obj.BatteryRange,
		&obj.EstBatteryRange,
		&obj.InsideTemp,
		&obj.OutsideTemp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(obj)
}
