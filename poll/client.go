package poll

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/christopher-wong/teslatrack/ownerapi"
	"github.com/go-redis/redis"
)

const workQueueID = "WorkQueue"

type Client struct {
	RedisClient *redis.Client
	Store       *sql.DB
	hHtpClient  *http.Client
	WorkQueueID string
}

func (c *Client) fetchNextID() (string, error) {
	return c.RedisClient.RPop(c.WorkQueueID).Result()
}

// PullWorkTo loops forever, pulling work from the work queue and sending userID's to the given channel
func (c *Client) PullWorkTo(to chan<- string) {
	for {
		userID, err := c.fetchNextID()
		if err != nil {
			// TODO: should we rate-limit this?
		} else {
			to <- userID
		}
	}
}

// RunWorker blocks, continuously polling the redis work queue and saving car status's
func (c *Client) RunWorker() {
	queue := make(chan string, 2)
	go c.PullWorkTo(queue)
	for userID := range queue {
		c.saveCarStatusForUserID(userID)
	}
}

func (c *Client) saveCarStatusForUserID(userID string) error {
	c.pollCarForUserID(userID)
	// TODO: implement saving
	panic("Unimplemented!")
}

// fetches the data on tesla's servers
func (c *Client) pollCarForUserID(userID string) (string, error) {
	token, err := c.tokenForUserID(userID)
	if err != nil {
		return "", err
	}
	teslaClient := &ownerapi.Client{
		HttpClient:           c.hHtpClient,
		OwnerAPIAuthResponse: token,
	}
	resp, err := teslaClient.GetVehiclesList()
	if err != nil {
		return "", err
	}

	// for now only poll the first vehicle
	if resp.Count > 0 {
		// TODO: poll the vehicle and save the data
		panic("Unimplemented!")
	}
	return "", errors.New("No available vehicles for given userID")
}

// retrieves the tesla token for the given userID
func (c *Client) tokenForUserID(userID string) (*ownerapi.OwnerAPIAuthResponse, error) {
	//TODO: Implement this
	panic("Unimplemented!")
}
