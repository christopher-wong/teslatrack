package ownerapi

import (
	"encoding/json"
	"fmt"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

type WakeUpResponse struct {
	Response struct {
		ID                     int64       `json:"id"`
		UserID                 int         `json:"user_id"`
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
}

func (c *Client) WakeUp(id int64) (*WakeUpResponse, error) {

	retryClient := retryablehttp.NewClient()

	req, err := retryablehttp.NewRequest("POST", fmt.Sprintf("%s/api/1/vehicles/%d/wake_up", baseURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.OwnerAPIAuthResponse.AccessToken)
	resp, err := retryClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var decoded WakeUpResponse
	err = json.NewDecoder(resp.Body).Decode(&decoded)

	return &decoded, nil
}
