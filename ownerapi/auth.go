package ownerapi

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	clientID          = "81527cff06843c8634fdc09e8ac0abefb46ac849f38fe1e431c2ef2106796384"
	clientSecret      = "c7257eb71a564034f9419ee651c7d0e5f7aa6bfbd18bafb5c5c033b093bb2fa3"
	grantTypePassword = "password"
	grantTypeRefresh  = "refresh_token"
)

type OwnerAPIAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int    `json:"created_at"`
}

type GetAuthTokenInput struct {
	Email        string
	Password     string
	RefreshToken string
}

func NewClient(httpClient *http.Client, input *GetAuthTokenInput) (*Client, error) {
	client := &Client{HttpClient: httpClient}
	resp, err := client.getAuthToken(input)
	if err != nil {
		return nil, err
	}
	client.OwnerAPIAuthResponse = resp
	return client, nil
}

func (c *Client) refreshToken(input *GetAuthTokenInput) (*OwnerAPIAuthResponse, error) {
	resp, err := c.HttpClient.PostForm(baseURL+"/oauth/token", url.Values{
		"grant_type":    {grantTypeRefresh},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {input.RefreshToken}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, err
	}

	var decoded OwnerAPIAuthResponse
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	return &decoded, err
}

func (c *Client) getAuthToken(input *GetAuthTokenInput) (*OwnerAPIAuthResponse, error) {
	resp, err := c.HttpClient.PostForm(baseURL+"/oauth/token", url.Values{
		"grant_type":    {grantTypePassword},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"email":         {input.Email},
		"password":      {input.Password}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, err
	}

	var decoded OwnerAPIAuthResponse
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	return &decoded, err
}
