package services

import (
	"time"
)

type CompletedChargeDBResponse struct {
	Timestamp         time.Time
	Latitude          float32
	Longitude         float32
	Odometer          float32
	CarVersion        string
	ChargingState     string
	ChargeEnergyAdded float32
	BatteryRange      float32
	BatteryLevel      int32
	EstBatteryRange100Pct float64
}

func (s *ServicesClient) GetCompletedChargeRows(userID int32) (*[]CompletedChargeDBResponse, error) {
	query := `
SELECT w1.timestamp,
       w1.latitude,
       w1.longitude,
       w1.odometer,
       w1.car_version,
       w1.charging_state,
       w1.charge_energy_added,
       w1.battery_range,
       w1.battery_level
FROM (SELECT w2.timestamp,
             w2.data -> 'response' -> 'drive_state' ->> 'latitude'             as latitude,
             w2.data -> 'response' -> 'drive_state' ->> 'longitude'            as longitude,
             w2.data -> 'response' -> 'charge_state' ->> 'charging_state'      as charging_state,
             w2.data -> 'response' -> 'charge_state' ->> 'charge_energy_added' as charge_energy_added,
             w2.data -> 'response' -> 'charge_state' ->> 'battery_range'       as battery_range,
             w2.data -> 'response' -> 'charge_state' ->> 'battery_level'       as battery_level,
             w2.data -> 'response' -> 'vehicle_state' ->> 'odometer'           as odometer,
             w2.data -> 'response' -> 'vehicle_state' ->> 'car_version'        as car_version,
             lead(w2.data -> 'response' -> 'charge_state' ->> 'charging_state')
             OVER (ORDER BY w2.timestamp DESC)                                 as prev_charging_state
      FROM state w2
      WHERE user_id = $1
      ORDER BY w2.timestamp DESC) as w1
WHERE w1.charging_state IS DISTINCT FROM w1.prev_charging_state
  and w1.charging_state = 'Complete'
ORDER BY w1.timestamp DESC;`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		s.logger.Println("failed to query for completed charging session data")
		return nil, err
	}

	var results []CompletedChargeDBResponse

	defer rows.Close()
	for rows.Next() {
		var result CompletedChargeDBResponse
		err = rows.Scan(
			&result.Timestamp,
			&result.Latitude,
			&result.Longitude,
			&result.Odometer,
			&result.CarVersion,
			&result.ChargingState,
			&result.ChargeEnergyAdded,
			&result.BatteryRange,
			&result.BatteryLevel,
		)
		if err != nil {
			s.logger.Println("failed to scan row in results for completed charging session data")
			s.logger.Println(err.Error())
		}

		results = append(results, result)
	}
	err = rows.Err()
	if err != nil {
		s.logger.Println("failed to read db results")
		s.logger.Println(err.Error())
	}
	return &results, nil
}
