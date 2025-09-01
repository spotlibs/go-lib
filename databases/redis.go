package databases

import (
	"sync"

	"github.com/goravel/framework/facades"
	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	RedisOnce   sync.Once
)

func GetRedisClient() *redis.Client {
	RedisOnce.Do(func() {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     facades.Config().GetString("database.redis.default.host") + ":" + facades.Config().GetString("database.redis.default.port"),
			Password: facades.Config().GetString("database.redis.default.password"),
			DB:       facades.Config().GetInt("database.redis.default.database"),
		})
	})
	return RedisClient
}
