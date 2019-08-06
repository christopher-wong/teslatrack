SELECT DISTINCT (tb.car_version)
FROM (
         SELECT timestamp,
                data -> 'response' -> 'vehicle_state' ->> 'car_version' AS car_version
         FROM state
         WHERE user_id = 1
         ORDER BY timestamp DESC
     ) as tb