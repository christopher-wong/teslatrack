package ownerapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetVehicleData returns a all of the vehicle configuration
// and state data in a single response. Use VehiclesResponse.ID as input.
//
// See https://tesla-api.timdorr.com/api-basics/vehicles#vehicle_id-vs-id
// for the difference between the vehicle_id and id.
func (c *Client) GetVehicleData(id int64) ([]byte, error) {
	if id == 0 {
		return []byte{}, nil
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/1/vehicles/%d/vehicle_data", baseURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.OwnerAPIAuthResponse.AccessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	return bodyBytes, nil
}
