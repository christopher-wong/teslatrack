package ownerapi

import "net/http"

const (
	baseURL = "https://owner-api.teslamotors.com"
)

type Client struct {
	HttpClient           *http.Client
	OwnerAPIAuthResponse *OwnerAPIAuthResponse
}
