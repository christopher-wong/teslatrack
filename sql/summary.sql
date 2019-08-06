-- SELECT timestamp,
--        data -> 'response' -> 'charge_state' ->> 'charging_state'                                      AS charging_state,
-- FROM (SELECT )
-- WHERE user_id = 1
-- ORDER BY timestamp DESC;

SELECT
    w1.day, w1.rainy
FROM
    (SELECT
        w2.day,
        w2.rainy,
        lead(w2.rainy) OVER (ORDER BY w2.day DESC) as prev_rainy
     FROM
        weather w2
     ORDER BY
        w2.day DESC) as w1
WHERE
    w1.rainy IS DISTINCT FROM w1.prev_rainy
AND w1.user_id = 1
ORDER BY
    w1.day DESC;