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
	rd "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/redis"
	middleware "github.com/dwiw96/vocagame-technical-test-backend/pkg/middleware"
	conv "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/converter"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"
	password "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/password"
	testUtils "github.com/dwiw96/vocagame-technical-test-backend/testutils"

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
	pool = testUtils.GetPool()
	defer pool.Close()
	ctx = testUtils.GetContext()
	defer ctx.Done()

	schemaCleanup := testUtils.SetupDB("test_cache_auth")

	password.JwtInit(pool, ctx)

	os.Setenv("REDIS_HOST", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "0")

	redis_db, err := conv.ConvertStrToInt(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal(err)
	}

	env := &cfg.EnvConfig{
		REDIS_HOST:     os.Getenv("REDIS_HOST"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		REDIS_DB:       redis_db,
	}

	client = rd.ConnectToRedis(env)
	defer client.Close()

	cacheTest = NewAuthCache(client, ctx)

	exitTest := m.Run()

	schemaCleanup()

	os.Exit(exitTest)
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

	err = testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

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
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

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
