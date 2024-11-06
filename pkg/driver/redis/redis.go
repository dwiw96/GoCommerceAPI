package redisPkg

import (
	"github.com/redis/go-redis/v9"

	cfg "github.com/dwiw96/vocagame-technical-test-backend/config"
)

func ConnectToRedis(envCfg *cfg.EnvConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     envCfg.REDIS_HOST,
		Password: envCfg.REDIS_PASSWORD, // no password set
		DB:       envCfg.REDIS_DB,       // use default DB
	})

	return client
}
