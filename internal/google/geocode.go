package google

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	googleMapsAPIKey = "AIzaSyDvsfaS7xpdlLw_ONQsujaRYKxRPkXAVts"
)

func ReverseGeocode(httpClient *http.Client, lat, long string) string {
	req, err := http.NewRequest("GET", fmt.Sprint("https://maps.googleapis.com/maps/api/geocode/json"), nil)
	if err != nil {
		return ""
	}

	q := req.URL.Query()
	q.Add("latlng", fmt.Sprintf("%s,%s", lat, long))
	q.Add("key", googleMapsAPIKey)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.Status
	}

	var decoded ReverseGeocodeResponse
	json.NewDecoder(resp.Body).Decode(&decoded)

	return decoded.Results[0].FormattedAddress
}
