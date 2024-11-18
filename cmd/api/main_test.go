package main

import (
	"context"
	"os"
	"testing"

	cfg "github.com/dwiw96/vocagame-technical-test-backend/config"
	pg "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/postgresql"
	password "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/password"
)

func TestMain(m *testing.M) {
	env := cfg.GetEnvConfig()
	pgPool := pg.ConnectToPg(env)
	defer pgPool.Close()

	ctx := context.Background()
	defer ctx.Done()

	password.JwtInit(pgPool, ctx)

	os.Exit(m.Run())
}
