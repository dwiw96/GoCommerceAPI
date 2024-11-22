package main

import (
	"os"
	"testing"

	password "github.com/dwiw96/GoCommerceAPI/pkg/utils/password"
	testUtils "github.com/dwiw96/GoCommerceAPI/testutils"
)

func TestMain(m *testing.M) {
	pgPool := testUtils.GetPool()
	defer testUtils.ClosePool()

	ctx := testUtils.GetContext()
	defer ctx.Done()

	password.JwtInit(pgPool, ctx)

	os.Exit(m.Run())
}
