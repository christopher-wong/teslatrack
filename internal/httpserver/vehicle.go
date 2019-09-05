package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// VehicleBasicSummary represents a response object built and returned via
// the vehicle/basic-summary endpoint
type VehicleBasicSummary struct {
	Timestamp   time.Time `json:"timestamp"`
	DisplayName string    `json:"display_name"`
	Odometer    float64   `json:"odometer"`
	InsideTemp  *float32  `json:"inside_temp"`
	OutsideTemp *float32  `json:"outside_temp"`
	ChargeState struct {
		BatteryLevel          *int     `json:"battery_level"`
		BatteryRange          *float32 `json:"battery_range"`
		ChargePower           *float32 `json:"charge_power"`
		ChargerRate           *float32 `json:"charge_rate"`
		ChargerActualCurrent  *float32 `json:"charger_actual_current"`
		ChargerCurrentRequest *float32 `json:"charge_current_request"`
		ChargerVoltage        *int32   `json:"charger_voltage"`
		ChargingState         *string  `json:"charging_state"`
		EstBatteryRange       *float32 `json:"est_battery_range"`
		TimeToFullCharge      *float32 `json:"time_to_full_charge"`
	} `json:"charge_state"`
}

// GetVehicleBasicSummary retrieves basic summary data using the userID inside
// the JWT.
func (s *Server) GetVehicleBasicSummary(w http.ResponseWriter, r *http.Request) {
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
		SELECT timestamp,
			data -> 'response' ->> 'display_name'                             AS display_name,
			data -> 'response' -> 'vehicle_state' ->> 'odometer'              AS odometer,
			data -> 'response' -> 'climate_state' ->> 'inside_temp'           AS inside_temp,
			data -> 'response' -> 'climate_state' ->> 'outside_temp'          AS outside_temp,
			data -> 'response' -> 'charge_state' ->> 'battery_level'          AS battery_level,
			data -> 'response' -> 'charge_state' ->> 'battery_range'          AS battery_range,
			data -> 'response' -> 'charge_state' ->> 'charger_power'          AS charger_power,
			data -> 'response' -> 'charge_state' ->> 'charge_rate'            AS charge_rate,
			data -> 'response' -> 'charge_state' ->> 'charger_actual_current' AS charger_actual_current,
			data -> 'response' -> 'charge_state' ->> 'charge_current_request' AS charge_current_request,
			data -> 'response' -> 'charge_state' ->> 'charger_voltage'        AS charger_voltage,
			data -> 'response' -> 'charge_state' ->> 'charging_state'         AS charging_state,
			data -> 'response' -> 'charge_state' ->> 'est_battery_range'      AS est_battery_range,
			data -> 'response' -> 'charge_state' ->> 'time_to_full_charge'    AS time_to_full_charge
		FROM state
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT 1;
	`

	result := s.db.QueryRow(query, userID)

	obj := &VehicleBasicSummary{}
	err = result.Scan(
		&obj.Timestamp,
		&obj.DisplayName,
		&obj.Odometer,
		&obj.InsideTemp,
		&obj.OutsideTemp,
		&obj.ChargeState.BatteryLevel,
		&obj.ChargeState.BatteryRange,
		&obj.ChargeState.ChargePower,
		&obj.ChargeState.ChargerRate,
		&obj.ChargeState.ChargerActualCurrent,
		&obj.ChargeState.ChargerCurrentRequest,
		&obj.ChargeState.ChargerVoltage,
		&obj.ChargeState.ChargingState,
		&obj.ChargeState.EstBatteryRange,
		&obj.ChargeState.TimeToFullCharge,
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
