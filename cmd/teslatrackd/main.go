package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	server "github.com/christopher-wong/teslatrack/internal/httpserver"
	"github.com/christopher-wong/teslatrack/poll"
	"github.com/christopher-wong/teslatrack/queuer"
	"github.com/go-redis/redis"
	"github.com/ianschenck/envflag"

	_ "github.com/lib/pq"
)

const (
	host     = "teslatrack-stage-do-user-2432224-0.db.ondigitalocean.com"
	port     = 25060
	user     = "doadmin"
	password = "a3uot0pp9bxzcxoa"
	dbname   = "defaultdb"
	sslmode  = "require"

	redisHost     = "157.230.142.205:6379"
	redisPassword = "FkNU6btkbjp+RwIG9529yJZG+EfNboVHEC6FzhpifbNMC0fIPC/MJP0/kvo3GYuT7LgkhGDVfE1gEDch"
)

var (
	listenAddr = envflag.String("TESLATRACK_LISTEN_ADDR", "0.0.0.0:8000", "address to listen on")
	dev        = envflag.Bool("GLASS_DEV_MODE", true, "set dev to false to serve static assets in prod")

	jwtKey = []byte("my_secret_key")
)

func main() {
	db, err := dbinit(host, user, password, dbname, sslmode, port)
	if err != nil {
		log.Fatal(err)
	}

	rc, err := redisinit(redisHost, redisPassword)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &server.Config{
		ListenAddress: *listenAddr,
		Dev:           *dev,
		JwtKey:        jwtKey,
	}

	// start API server
	app, err := server.New(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	// start http server
	go func() {
		err = app.Run()
		log.Fatal("failed to start api server")
	}()

	// run background tasks to poll car
	go runBackgroundPoll(rc, db)

	// run background task to push work to Redis
	go runBackgroundQueuer(rc, db)

	// stop main thread from exiting
	select {}
}

func runBackgroundQueuer(rc *redis.Client, db *sql.DB) {
	client := &queuer.Client{
		RedisClient: rc,
		Store:       db,
	}

	client.RunQueuer()
}

func runBackgroundPoll(rc *redis.Client, db *sql.DB) {
	pollClient := &poll.Client{
		RedisClient: rc,
		Store:       db,
		HTTPClient:  &http.Client{},
	}

	pollClient.RunWorker()
}

func redisinit(host, password string) (*redis.Client, error) {
	// connect to redis
	rc := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	_, err := rc.Ping().Result()
	if err != nil {
		log.Println("failed to connect to redis!")
		return nil, err
	}

	log.Println("successfully connected to redis")

	return rc, nil
}

func dbinit(host, user, password, dbName, sslMode string, port int) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Println("failed to ping database")
		return nil, err
	}

	log.Println("successfully connected to postgres")
	return db, nil
}
