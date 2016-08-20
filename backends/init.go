package backends

import "gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"

func MakeRedis(host, password string) *redis.Redis {
	return redis.MakeRedis(host, password)
}
