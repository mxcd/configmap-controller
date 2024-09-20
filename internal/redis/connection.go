package redis

import (
	"fmt"

	"github.com/mxcd/go-config/config"
	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	Client *redis.Client
}

type RedisConnectionOptions struct {
	Host          string
	Port          int
	Password      string
	DatabaseIndex int
	Sentinel      bool
}

func NewRedisConnection(options *RedisConnectionOptions) (*RedisConnection, error) {
	redisHost := config.Get().String("REDIS_HOST")
	redisPort := config.Get().Int("REDIS_PORT")
	redisPassword := config.Get().String("REDIS_PASSWORD")
	redisDatabaseIndex := config.Get().Int("REDIS_DATABASE_INDEX")

	redisAddress := fmt.Sprintf("%s:%d", redisHost, redisPort)

	redisConnection := &RedisConnection{
		// TODO: add support for sentinel
		// TODO: add support for OTEL
		Client: redis.NewClient(&redis.Options{
			Addr:     redisAddress,
			Password: redisPassword,
			DB:       redisDatabaseIndex,
		}),
	}
	return redisConnection, nil
}

func (c *RedisConnection) Close() {
	c.Client.Close()
}
