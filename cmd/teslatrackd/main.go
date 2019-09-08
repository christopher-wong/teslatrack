package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"

	"teslatrack/poll"
	"teslatrack/queuer"
	"teslatrack/services"

	log "github.com/sirupsen/logrus"

	server "teslatrack/internal/httpserver"

	"github.com/go-redis/redis"
	"github.com/ianschenck/envflag"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var (
	listenAddr = envflag.String("TESLATRACK_LISTEN_ADDR", "0.0.0.0:8000", "address to listen on")
	dev        = envflag.Bool("DEV_MODE", true, "set dev to false to serve static assets in prod")

	jwtKey = []byte("my_secret_key")
)

func main() {
	viper.SetConfigName("config") // name of config file (without extension)

	viper.AddConfigPath(".") // optionally look for config in the working directory
	viper.AddConfigPath("../../")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	fetch := viper.GetBool("teslatrack.fetch")

	host := viper.GetString("postgres.host")
	user := viper.GetString("postgres.user")
	password := viper.GetString("postgres.password")
	dbname := viper.GetString("postgres.dbname")
	sslmode := viper.GetString("postgres.sslmode")
	port := viper.GetInt("postgres.port")

	redisHost := viper.GetString("redis.host")
	redisPassword := viper.GetString("redis.password")

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

	// run services jobs
	svcClient := services.New(db, log.New())
	go svcClient.CalculateBatteryDegradationStats()

	// start API server
	app, err := server.New(cfg, db, svcClient)
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

	if fetch {
		// run background task to push work to Redis
		go runBackgroundQueuer(rc, db)
	}

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
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
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
