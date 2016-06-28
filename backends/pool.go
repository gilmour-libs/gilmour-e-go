package backends

import (
	"log"
	"sync"
	"time"

	"gopkg.in/redis.v4"
)

var once sync.Once

func newClient(server, password string) *redis.Client {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	options := &redis.Options{
		Network:     "tcp",
		Addr:        server,
		Password:    password,
		DB:          0,
		PoolSize:    1000,
		IdleTimeout: 240 * time.Second,
	}

	return redis.NewClient(options)
}

var cached = struct {
	client *redis.Client
}{}

func getClient(redis_host, password string) *redis.Client {
	once.Do(func() {
		cached.client = newClient(redis_host, password)
	})

	return cached.client
}
