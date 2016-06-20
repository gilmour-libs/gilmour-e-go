package backends

import (
	"log"
	"sync"
	"time"

	"gopkg.in/redis.v3"
)

func newClient(server, password string) *redis.Client {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	options := &redis.Options{
		Network: "tcp",
		Addr:    server,

		// Dialer creates new network connection and has priority over
		// Network and Addr options.
		// Dialer func() (net.Conn, error)

		// An optional password. Must match the password specified in the
		// requirepass server configuration option.
		Password: password,
		// A database to be selected after connecting to server.
		DB: 0,

		// The maximum number of retries before giving up.
		// Default is to not retry failed commands.
		// MaxRetries int

		// Sets the deadline for establishing new connections. If reached,
		// dial will fail with a timeout.
		// Default is 5 seconds.
		// DialTimeout time.Duration
		// Sets the deadline for socket reads. If reached, commands will
		// fail with a timeout instead of blocking.
		// ReadTimeout time.Duration
		// Sets the deadline for socket writes. If reached, commands will
		// fail with a timeout instead of blocking.
		// WriteTimeout time.Duration

		// The maximum number of socket connections.
		// Default is 10 connections.
		PoolSize: 1000,
		// Specifies amount of time client waits for connection if all
		// connections are busy before returning an error.
		// Default is 1 second.
		// PoolTimeout time.Duration
		// Specifies amount of time after which client closes idle
		// connections. Should be less than server's timeout.
		// Default is to not close idle connections.
		IdleTimeout: 240 * time.Second,
		// The frequency of idle checks.
		// Default is 1 minute.
		// IdleCheckFrequency time.Duration
	}

	return redis.NewClient(options)
}

var cached = struct {
	sync.RWMutex
	client *redis.Client
}{}

func getClient(redis_host, password string) *redis.Client {

	cached.Lock()
	if cached.client == nil {
		cached.client = newClient(redis_host, password)
	}
	cached.Unlock()

	return cached.client
}
