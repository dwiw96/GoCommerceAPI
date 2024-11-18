package chache

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	cfg "github.com/dwiw96/vocagame-technical-test-backend/config"
	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	pg "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/postgresql"
	rd "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/redis"
	middleware "github.com/dwiw96/vocagame-technical-test-backend/pkg/middleware"
	conv "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/converter"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cacheTest auth.ICache
	pool      *pgxpool.Pool
	client    *redis.Client
	ctx       context.Context
	key       *rsa.PrivateKey
)

func TestMain(m *testing.M) {
	os.Setenv("DB_USERNAME", "dwiw")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "technical_test")
	os.Setenv("REDIS_HOST", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "0")

	redis_db, err := conv.ConvertStrToInt(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal(err)
	}

	envConfig := &cfg.EnvConfig{
		DB_USERNAME:    os.Getenv("DB_USERNAME"),
		DB_PASSWORD:    os.Getenv("DB_PASSWORD"),
		DB_HOST:        os.Getenv("DB_HOST"),
		DB_PORT:        os.Getenv("DB_PORT"),
		DB_NAME:        os.Getenv("DB_NAME"),
		REDIS_HOST:     os.Getenv("REDIS_HOST"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		REDIS_DB:       redis_db,
	}

	pool = pg.ConnectToPg(envConfig)

	client = rd.ConnectToRedis(envConfig)
	defer client.Close()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cacheTest = NewAuthCache(client, ctx)

	os.Exit(m.Run())
}

func createToken(t *testing.T) (payload *auth.JwtPayload) {
	var err error
	key, err = middleware.LoadKey(ctx, pool)
	require.NoError(t, err)
	require.NotNil(t, key)

	user := auth.User{
		ID:       int32(generator.RandomInt(1, 100)),
		Username: generator.CreateRandomString(5) + " " + generator.CreateRandomString(7),
		Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
	}
	token, err := middleware.CreateToken(user, 5, key)
	require.NoError(t, err)
	require.NotZero(t, len(token))

	payload, err = middleware.ReadToken(token, key)
	require.NoError(t, err)

	return
}

func TestCachingBlockedToken(t *testing.T) {
	var err error
	tests := []struct {
		name    string
		payload *auth.JwtPayload
		err     bool
	}{
		{
			name:    "success",
			payload: createToken(t),
			err:     false,
		}, {
			name:    "success",
			payload: createToken(t),
			err:     false,
		}, {
			name:    "failed_duration_minus",
			payload: createToken(t),
			err:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.err {
				err = cacheTest.CachingBlockedToken(*test.payload)
				require.NoError(t, err)

				res, err := client.Get(ctx, fmt.Sprint("block ", test.payload.ID)).Result()
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprint(test.payload.UserID), res)
			} else {
				now := time.Now().UTC().Add(1)
				test.payload.Exp = now.Unix()
				err = cacheTest.CachingBlockedToken(*test.payload)
				require.NoError(t, err)

				res, err := client.Get(ctx, fmt.Sprint("block ", test.payload.ID)).Result()
				require.Error(t, err)
				assert.Empty(t, res)
			}
		})
	}
}

func TestCheckBlockedToken(t *testing.T) {
	var err error

	tests := []struct {
		name    string
		payload *auth.JwtPayload
		err     bool
	}{
		{
			name:    "valid",
			payload: createToken(t),
			err:     false,
		}, {
			name:    "blacklist",
			payload: createToken(t),
			err:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log("id:", test.payload.ID)
			if !test.err {
				err = cacheTest.CheckBlockedToken(*test.payload)
			} else {
				err = cacheTest.CachingBlockedToken(*test.payload)
				require.NoError(t, err)

				err = cacheTest.CheckBlockedToken(*test.payload)
				require.Error(t, err)
			}
		})
	}
}
