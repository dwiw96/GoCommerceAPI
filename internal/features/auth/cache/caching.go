package chache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
)

type authCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewAuthCache(client *redis.Client, ctx context.Context) auth.ICache {
	return &authCache{
		client: client,
		ctx:    ctx,
	}
}

func (c *authCache) CachingBlockedToken(payload auth.JwtPayload) error {
	now := time.Now().UTC()
	exp := time.Unix(payload.Exp, 0)

	duration := time.Duration(exp.Sub(now).Nanoseconds())
	if duration <= 0 {
		return nil
	}

	err := c.client.Set(c.ctx, fmt.Sprint("block ", payload.ID), payload.UserID, duration).Err()
	if err != nil {
		return fmt.Errorf("failed to caching token, msg: %v", err)
	}

	return nil
}

func (r *authCache) CheckBlockedToken(payload auth.JwtPayload) error {
	check, err := r.client.Exists(r.ctx, "block "+payload.ID.String()).Result()
	if err != nil {
		return err
	}
	if check != 0 {
		return fmt.Errorf("token is blacklist")
	}

	return nil
}
