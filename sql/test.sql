SELECT id,
       timestamp,
       data -> 'response' -> 'drive_state' ->> 'gps_as_of' as gps_as_of,
       data -> 'response' -> 'drive_state' ->> 'latitude' as latitude,
       data -> 'response' -> 'drive_state' ->> 'longitude' as longitude,
       data -> 'response' -> 'drive_state' ->> 'shift_state' as shift_state,
       data -> 'response' -> 'drive_state' ->> 'speed' as speed
FROM state
WHERE timestamp between '2019-06-16' and '2019-06-21';