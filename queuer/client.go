package queuer

import (
	"database/sql"

	"github.com/go-redis/redis"
)

type Client struct {
	RedisClient *redis.Client
	Store       *sql.DB
}
