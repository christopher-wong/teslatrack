package ownerapi

import (
	"encoding/json"
	"net/http"
)

type VehiclesResponse struct {
	Response []struct {
		ID                     int64       `json:"id"`
		VehicleID              int         `json:"vehicle_id"`
		Vin                    string      `json:"vin"`
		DisplayName            string      `json:"display_name"`
		OptionCodes            string      `json:"option_codes"`
		Color                  interface{} `json:"color"`
		Tokens                 []string    `json:"tokens"`
		State                  string      `json:"state"`
		InService              bool        `json:"in_service"`
		IDS                    string      `json:"id_s"`
		CalendarEnabled        bool        `json:"calendar_enabled"`
		APIVersion             int         `json:"api_version"`
		BackseatToken          interface{} `json:"backseat_token"`
		BackseatTokenUpdatedAt interface{} `json:"backseat_token_updated_at"`
	} `json:"response"`
	Count int `json:"count"`
}

func (c *Client) GetVehiclesList() (*VehiclesResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/api/1/vehicles", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.OwnerAPIAuthResponse.AccessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 && resp.StatusCode > 299 {
		return nil, err
	}

	var decoded VehiclesResponse
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	return &decoded, err
}
