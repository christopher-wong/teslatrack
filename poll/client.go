package poll

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/christopher-wong/teslatrack/ownerapi"
	"github.com/go-redis/redis"
)

const workQueueID = "WorkQueue"

type Client struct {
	RedisClient *redis.Client
	Store       *sql.DB
	HTTPClient  *http.Client
}

func (c *Client) fetchNextID() (string, error) {
	return c.RedisClient.RPop(workQueueID).Result()
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
		err := c.saveCarStatusForUserID(userID)
		if err != nil {
			fmt.Println(err)
			// TODO: log that we failed this
		}
	}
}

func (c *Client) saveCarStatusForUserID(userID string) error {
	resp, err := c.pollCarForUserID(userID)
	if err != nil {
		log.Println(fmt.Sprintf("failed to poll car for userID: %s", userID))
		return err
	}

	insertQuery := `
		INSERT INTO state (
			user_id, timestamp, data
		)
		VALUES ($1, $2, $3)
	`

	_, err = c.Store.Exec(insertQuery, userID, time.Now(), resp)
	if err != nil {
		fmt.Println("failed to insert Tesla state into the database")
		fmt.Println(err)
		return err
	}

	return nil
}

// fetches the data on tesla's servers
func (c *Client) pollCarForUserID(userID string) ([]byte, error) {
	token, err := c.tokenForUserID(userID)
	if err != nil {
		return nil, err
	}
	teslaClient := &ownerapi.Client{
		HttpClient: c.HTTPClient,
		OwnerAPIAuthResponse: &ownerapi.OwnerAPIAuthResponse{
			AccessToken: token,
		},
	}
	resp, err := teslaClient.GetVehiclesList()
	if err != nil {
		return nil, err
	}

	// for now only poll the first vehicle
	if resp.Count > 0 {
		firstID := resp.Response[0].ID
		resp, err := teslaClient.GetVehicleData(firstID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to get vehicle data for user: %s | id: %d", userID, firstID))
		}
		return resp, nil
	}
	return nil, errors.New("No available vehicles for given userID")
}

// retrieves the tesla token for the given userID
func (c *Client) tokenForUserID(userID string) (string, error) {
	var accessToken string
	err := c.Store.QueryRow(`
		SELECT access_token
		FROM tesla_auth
		WHERE user_id = $1
		ORDER BY id DESC
	`, userID).Scan(&accessToken)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(err)
			return "", err
		}
		fmt.Println(err)
		return "", err
	}

	return accessToken, nil
}
