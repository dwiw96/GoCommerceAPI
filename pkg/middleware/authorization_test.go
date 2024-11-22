package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"os"
	"testing"
	"time"

	cfg "github.com/dwiw96/GoCommerceAPI/config"
	auth "github.com/dwiw96/GoCommerceAPI/internal/features/auth"
	pg "github.com/dwiw96/GoCommerceAPI/pkg/driver/postgresql"
	rd "github.com/dwiw96/GoCommerceAPI/pkg/driver/redis"
	generator "github.com/dwiw96/GoCommerceAPI/pkg/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	pool   *pgxpool.Pool
	ctx    context.Context
	client *redis.Client
)

func TestMain(m *testing.M) {
	os.Setenv("DB_USERNAME", "dwiw")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "commerce_main_db")
	os.Setenv("REDIS_HOST", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")

	envConfig := &cfg.EnvConfig{
		DB_USERNAME:    os.Getenv("DB_USERNAME"),
		DB_PASSWORD:    os.Getenv("DB_PASSWORD"),
		DB_HOST:        os.Getenv("DB_HOST"),
		DB_PORT:        os.Getenv("DB_PORT"),
		DB_NAME:        os.Getenv("DB_NAME"),
		REDIS_HOST:     os.Getenv("REDIS_HOST"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
	}

	pool = pg.ConnectToPg(envConfig)

	client = rd.ConnectToRedis(envConfig)

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	os.Exit(m.Run())
}

func createTokenAndKey(t *testing.T) (string, auth.User, *rsa.PrivateKey) {
	key, err := LoadKey(ctx, pool)
	require.NoError(t, err)
	require.NotNil(t, key)

	firstname := generator.CreateRandomString(5)
	payload := auth.User{
		Username: firstname + " " + generator.CreateRandomString(7),
		Email:    generator.CreateRandomEmail(firstname),
	}

	token, err := CreateToken(payload, 5, key)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	return token, payload, key
}

func TestCreateToken(t *testing.T) {
	createTokenAndKey(t)
}

func TestVerifyToken(t *testing.T) {
	token, _, key := createTokenAndKey(t)

	t.Run("success", func(t *testing.T) {
		res, err := VerifyToken(token, key)
		require.NoError(t, err)
		require.True(t, res)
	})

	t.Run("failed", func(t *testing.T) {
		res, err := VerifyToken(token+"b", key)
		require.Error(t, err)
		require.False(t, res)
	})
}

func TestReadToken(t *testing.T) {
	token, payloadInput, key := createTokenAndKey(t)

	t.Run("success", func(t *testing.T) {
		payload, err := ReadToken(token, key)
		require.NoError(t, err)
		assert.Equal(t, payloadInput.Username, payload.Name)
		assert.Equal(t, payloadInput.Email, payload.Email)
	})

	t.Run("failed", func(t *testing.T) {
		payload, err := ReadToken(token+"b", key)
		require.Error(t, err)
		assert.Nil(t, payload)
	})
}

func TestLoadKey(t *testing.T) {
	res, err := LoadKey(ctx, pool)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestCheckBlockedToken(t *testing.T) {
	token, _, key := createTokenAndKey(t)

	payload, err := ReadToken(token, key)
	require.NoError(t, err)

	t.Run("valid", func(t *testing.T) {
		err = CheckBlockedToken(client, ctx, payload.ID)
		require.NoError(t, err)
	})

	t.Run("blacklist", func(t *testing.T) {
		iat := time.Unix(payload.Iat, 0)
		exp := time.Unix(payload.Exp, 0)
		duration := time.Duration(exp.Sub(iat).Nanoseconds())
		err = client.Set(ctx, "block "+payload.ID.String(), payload.UserID, duration).Err()
		require.NoError(t, err)

		err = CheckBlockedToken(client, ctx, payload.ID)
		require.Error(t, err)
	})
}

func TestGetHeaderToken(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", nil)
	require.NoError(t, err)

	authHeader := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImYyNGZiYmExLTE5NDctNGNhYy05ODA4LTM2ZDY2YzQ2NzIwMCIsInVzZXJfaWQiOjc2LCJpc3MiOiIiLCJuYW1lIjoiR3JhY2UgRG9lIEp1bmlvciIsImVtYWlsIjoiZ3JhY2VAbWFpbC5jb20iLCJhZGRyZXNzIjoiQ2lyY2xlIFN0cmVldCwgTm8uMSwgQmFuZHVuZywgV2VzdCBKYXZhIiwiaWF0IjoxNzI3ODM3MjY5LCJleHAiOjE3Mjc4NDA4Njl9.mdQtJ22xRT5n8xYp5dGdVIzBo-OOocnaE6F054C0LEImf1rA_Fo0_fd3IGVa3XW5kDdpobqB8K6hDFm-XCPbkxvIfXjsjAwGqDrlzsjLiNmSvRwUj6FFWUkIpS_4Nl7Szcc2dEXe7n75LOs9yIhzNmuNjyC9Ago8BJiTYL0_jAkzxlHUwSaRj6naxbsLpiRhpjAW14-ema0wdbbHkaPkv0cj6rOQlsRTCW6R6i_2lrew5eOHIR750gBdImJ8HGtzB29yUA3A9P0-rGjITwZTanoqtOdv5d6lSMJ7eYMEACe4Lj3-k93V65e2ZJEFCnutk0H2ZPSaMBZwTx9B32S8JQ"

	r.Header.Set("Authorization", "Bearer "+authHeader)

	token, err := GetTokenHeader(r)
	require.NoError(t, err)
	assert.Equal(t, authHeader, token)
}

func TestPayloadVerification(t *testing.T) {
	var user auth.User
	user.Username = generator.CreateRandomString(int(generator.RandomInt(3, 13)))
	user.Email = generator.CreateRandomEmail(user.Username)
	user.HashedPassword = generator.CreateRandomString(int(generator.RandomInt(5, 10)))

	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.HashedPassword)

	query := `
	INSERT INTO users(
		email,
		username,
		hashed_password
	) VALUES 
		($1, $2, $3) 
	RETURNING id;`

	row := pool.QueryRow(ctx, query, user.Email, user.Username, user.HashedPassword)
	err := row.Scan(&user.ID)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)

	err = PayloadVerification(ctx, pool, user.Email, user.Username)
	require.NoError(t, err)
}
