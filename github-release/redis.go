package github_release

import (
	"fmt"
	"os"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

func newRedisInstance() (*cache.Cache, error) {
	opt, err := redis.ParseURL(os.Getenv("REDIS_SERVER"))
	if err != nil {
		return nil, fmt.Errorf("error parsing redis url: %w", err)
	}

	redisClient := redis.NewClient(opt)

	mycache := cache.New(&cache.Options{
		Redis:      redisClient,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return mycache, nil
}
