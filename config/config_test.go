package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnvConfig(t *testing.T) {
	ans := &EnvConfig{
		SERVER_PORT:    ":8080",
		DB_USERNAME:    "dwiw",
		DB_PASSWORD:    "secret",
		DB_HOST:        "localhost",
		DB_PORT:        "5432",
		DB_NAME:        "commerce_main_db",
		REDIS_HOST:     "localhost:6379",
		REDIS_PASSWORD: "",
		REDIS_DB:       0,
	}
	res := GetEnvConfig()
	require.NotNil(t, res)
	assert.Equal(t, ans, res)
}
