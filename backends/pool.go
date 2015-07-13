package gilmour

import (
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

var cached = struct {
	sync.RWMutex
	pool *redis.Pool
}{}

func GetConn(redis_host string) redis.Conn {

	cached.Lock()
	if cached.pool == nil {
		cached.pool = newPool(redis_host)
	}
	cached.Unlock()

	conn := cached.pool.Get()
	return conn
}
