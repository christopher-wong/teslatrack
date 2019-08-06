package queuer

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

const workQueueID = "WorkQueue"

type Client struct {
	RedisClient *redis.Client
	Store       *sql.DB
}

func (c *Client) RunQueuer() {
	query := "SELECT * FROM tesla_auth"

	var (
		rowID        int
		userID       int
		accessToken  string
		tokenType    string
		expiresIn    int
		refreshToken string
		createdAt    int64
	)

	// run loop forever every minute
	for {
		time.Sleep(60 * time.Second)

		// query db for all users and push them onto the queue
		rows, err := c.Store.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&rowID, &userID, &accessToken, &tokenType, &expiresIn, &refreshToken, &createdAt)
			if err != nil {
				fmt.Println("Failed to query db")
				log.Fatal(err)
			}
			log.Printf("push queue id: %d\n", userID)
			c.RedisClient.LPush(workQueueID, userID)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}
