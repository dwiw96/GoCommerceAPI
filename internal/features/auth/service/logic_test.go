package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	cfg "github.com/dwiw96/vocagame-technical-test-backend/config"
	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	cache "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/cache"
	repo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/repository"
	pg "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/postgresql"
	rd "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/redis"
	middleware "github.com/dwiw96/vocagame-technical-test-backend/pkg/middleware"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	serviceTest auth.IService
	pool        *pgxpool.Pool
	ctx         context.Context
	repoTest    auth.IRepository
)

func TestMain(m *testing.M) {
	os.Setenv("DB_USERNAME", "dwiw")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "technical_test")

	envConfig := &cfg.EnvConfig{
		DB_USERNAME: os.Getenv("DB_USERNAME"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_NAME:     os.Getenv("DB_NAME"),
	}

	pool = pg.ConnectToPg(envConfig)
	defer pool.Close()

	client := rd.ConnectToRedis(envConfig)
	defer client.Close()

	ctx = context.Background()
	defer ctx.Done()

	repoTest = repo.NewAuthRepository(pool, pool)
	cacheTest := cache.NewAuthCache(client, ctx)
	serviceTest = NewAuthService(repoTest, cacheTest, ctx)

	os.Exit(m.Run())
}

func createUser(t *testing.T) (user *auth.User, token string, signupReq auth.SignupRequest) {
	email := generator.CreateRandomEmail(generator.CreateRandomString(5))

	input := auth.SignupRequest{
		Username: generator.CreateRandomString(5),
		Email:    email,
		Password: generator.CreateRandomString(10),
	}

	res, token, code, err := serviceTest.SignUp(input)

	require.NoError(t, err)
	require.Equal(t, 200, code)
	assert.Equal(t, input.Username, res.Username)
	assert.Equal(t, input.Email, res.Email)
	assert.NotEqual(t, input.Password, res.HashedPassword)
	assert.False(t, res.IsVerified)
	assert.NotEmpty(t, token)

	return res, token, input
}

func TestSignUp(t *testing.T) {
	email := generator.CreateRandomEmail(generator.CreateRandomString(5))
	tests := []struct {
		desc  string
		input auth.SignupRequest
		err   bool
	}{
		{
			desc: "success",
			input: auth.SignupRequest{
				Username: generator.CreateRandomString(5),
				Email:    email,
				Password: generator.CreateRandomString(10),
			},
			err: false,
		}, {
			desc: "failed__empty_username",
			input: auth.SignupRequest{
				Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Password: generator.CreateRandomString(10),
			},
			err: true,
		}, {
			desc: "failed__empty_password",
			input: auth.SignupRequest{
				Username: generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
			},
			err: true,
		}, {
			desc: "failed__empty_email",
			input: auth.SignupRequest{
				Username: generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Password: generator.CreateRandomString(10),
			},
			err: true,
		}, {
			desc: "failed__duplicate_email",
			input: auth.SignupRequest{
				Username: generator.CreateRandomString(5),
				Email:    email,
				Password: generator.CreateRandomString(10),
			},
			err: true,
		},
	}

	for _, tC := range tests {
		t.Run(tC.desc, func(t *testing.T) {
			res, token, code, err := serviceTest.SignUp(tC.input)

			if !tC.err {
				// t.Log("res id:", res.ID)
				require.NoError(t, err)
				require.Equal(t, 200, code)
				assert.Equal(t, tC.input.Username, res.Username)
				assert.Equal(t, tC.input.Email, res.Email)
				assert.NotEqual(t, tC.input.Password, res.HashedPassword)
				assert.False(t, res.IsVerified)
				assert.NotEmpty(t, token)
			} else {
				require.Error(t, err)
				require.NotZero(t, code)
			}
		})
	}
}

func TestLogIn(t *testing.T) {
	user, _, signUpReq := createUser(t)

	tests := []struct {
		name  string
		input auth.LoginRequest
		err   bool
		code  int
	}{
		{
			name: "success",
			input: auth.LoginRequest{
				Email:    signUpReq.Email,
				Password: signUpReq.Password,
			},
			err:  false,
			code: 1,
		}, {
			name: "failed_email_wrong",
			input: auth.LoginRequest{
				Email:    "err" + signUpReq.Email,
				Password: signUpReq.Password,
			},
			err:  true,
			code: 2,
		}, {
			name: "failed_password_wrong",
			input: auth.LoginRequest{
				Email:    signUpReq.Email,
				Password: "err" + signUpReq.Password,
			},
			err:  true,
			code: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, accessToken, refreshToken, code, err := serviceTest.LogIn(test.input)
			if !test.err {
				require.NoError(t, err)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
				assert.Equal(t, 200, code)
				user.Username = res.Username

				assert.Equal(t, user.Username, res.Username)
				assert.Equal(t, user.Email, res.Email)
				assert.Equal(t, user.HashedPassword, res.HashedPassword)
				assert.NotZero(t, res.CreatedAt)
				assert.False(t, res.IsVerified)
			} else {
				require.Error(t, err)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
				assert.Equal(t, 401, code)
				assert.Nil(t, res)
			}

			if test.code == 2 {
				assert.Equal(t, err, fmt.Errorf("no user found with this email %s", test.input.Email))
			} else if test.code == 3 {
				assert.Equal(t, err, errors.New("password is wrong"))
			}
		})
	}
}

func TestLogOut(t *testing.T) {
	_, _, signUpReq := createUser(t)

	argLogin := auth.LoginRequest{
		Email:    signUpReq.Email,
		Password: signUpReq.Password,
	}

	_, accessToken, refreshToken, code, err := serviceTest.LogIn(argLogin)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, 200, code)

	key, err := repoTest.LoadKey(ctx)
	require.NoError(t, err)
	require.NotNil(t, key)

	payload, err := middleware.ReadToken(accessToken, key)
	require.NoError(t, err)

	err = serviceTest.LogOut(*payload)
	require.NoError(t, err)
}

func TestDeleteUser(t *testing.T) {
	var users []auth.User
	for i := 0; i < 5; i++ {
		user, _, _ := createUser(t)
		users = append(users, *user)
	}

	testCases := []struct {
		desc string
		arg  auth.DeleteUserParams
		err  bool
	}{
		{
			desc: "success",
			arg: auth.DeleteUserParams{
				ID:    users[0].ID,
				Email: users[0].Email,
			},
			err: false,
		}, {
			desc: "success",
			arg: auth.DeleteUserParams{
				ID:    users[1].ID,
				Email: users[1].Email,
			},
			err: false,
		}, {
			desc: "success",
			arg: auth.DeleteUserParams{
				ID:    users[2].ID,
				Email: users[2].Email,
			},
			err: false,
		}, {
			desc: "failed_wrong_id",
			arg: auth.DeleteUserParams{
				ID:    users[3].ID + 5,
				Email: users[3].Email,
			},
			err: true,
		}, {
			desc: "failed_wrong_email",
			arg: auth.DeleteUserParams{
				ID:    users[4].ID,
				Email: "a" + users[4].Email,
			},
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			code, err := serviceTest.DeleteUser(tC.arg)
			if !tC.err {
				require.Equal(t, 200, code)
				require.NoError(t, err)
			} else {
				require.NotEqual(t, 200, code)
				require.Error(t, err)
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	_, _, signUpReq := createUser(t)

	argLogin := auth.LoginRequest{
		Email:    signUpReq.Email,
		Password: signUpReq.Password,
	}

	_, accessToken, refreshToken, code, err := serviceTest.LogIn(argLogin)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, 200, code)
	accessTokenNoBearer := strings.Split(accessToken, " ")

	testCases := []struct {
		desc         string
		refreshToken string
		accessToken  string
		err          bool
	}{
		{
			desc:         "success",
			refreshToken: refreshToken,
			accessToken:  accessTokenNoBearer[1],
			err:          false,
		}, {
			desc:         "failed_invalid_access_token",
			refreshToken: refreshToken,
			accessToken:  accessTokenNoBearer[1] + "a",
			err:          true,
		}, {
			desc:         "failed_invalid_refresh_token",
			refreshToken: refreshToken + "a",
			accessToken:  accessTokenNoBearer[1],
			err:          true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			newRefreshToken, newAccessToken, code, err := serviceTest.RefreshToken(tC.refreshToken, tC.accessToken)
			if !tC.err {
				require.NoError(t, err)
				require.Equal(t, 200, code)
				assert.NotEmpty(t, newRefreshToken)
				assert.NotEmpty(t, newAccessToken)
			} else {
				require.Error(t, err)
			}
		})
	}
}
