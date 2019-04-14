package poll

import (
	"database/sql"
	"net/http"

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

// Poll loops forever, pulling work from the work queue and sending userID's to the given channel
func (c *Client) Poll(to chan<- string) {
	for {
		userID, err := c.fetchNextID()
		if err != nil {
			// TODO: should we rate-limit this?
		} else {
			to <- userID
		}
	}
}

func (c *Client) saveCarStatusForUserID(userID string) error {
	c.pollCarForUserID(userID)
	panic("Unimplemented!")
}

// fetches the data on tesla's servers
func (c *Client) pollCarForUserID(userID string) (string, error) {
	token, err := c.tokenForUserID(userID)
	if err != nil {
		return "", err
	}

	panic("Unimplemented!")
}

// retrieves the tesla token for the given userID
func (c *Client) tokenForUserID(userID string) (string, error) {
	//TODO: Implement this
	panic("Unimplemented!")
}
