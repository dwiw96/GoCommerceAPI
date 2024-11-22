package repository

import (
	"context"
	"os"

	"testing"

	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"
	password "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/password"
	testUtils "github.com/dwiw96/vocagame-technical-test-backend/testutils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	repoTest auth.IRepository
	pool     *pgxpool.Pool
	ctx      context.Context
)

func TestMain(m *testing.M) {
	pool = testUtils.GetPool()
	defer pool.Close()
	ctx = testUtils.GetContext()
	defer ctx.Done()

	schemaCleanup := testUtils.SetupDB("test_repo_auth")

	password.JwtInit(pool, ctx)

	repoTest = NewAuthRepository(pool, pool)

	exitTest := m.Run()

	schemaCleanup()

	os.Exit(exitTest)
}

func createRandomUser(t *testing.T) (res *auth.User) {
	username := generator.CreateRandomString(int(generator.RandomInt(3, 13)))
	arg := auth.CreateUserParams{
		Username:       username,
		Email:          generator.CreateRandomEmail(username),
		HashedPassword: generator.CreateRandomString(int(generator.RandomInt(20, 20))),
	}

	assert.NotEmpty(t, arg.Username)
	assert.NotEmpty(t, arg.Email)
	assert.NotEmpty(t, arg.HashedPassword)

	res, err := repoTest.CreateUser(ctx, arg)
	require.NoError(t, err)
	assert.NotZero(t, res.ID)
	assert.Equal(t, username, res.Username)
	assert.Equal(t, arg.Email, res.Email)
	assert.Equal(t, arg.HashedPassword, res.HashedPassword)

	return res
}

func TestCreateUser(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	email := generator.CreateRandomEmail(generator.CreateRandomString(5))
	testCases := []struct {
		desc string
		arg  auth.CreateUserParams
		err  bool
	}{
		{
			desc: "success",
			arg: auth.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          email,
				HashedPassword: generator.CreateRandomString(60),
			},
			err: false,
		}, {
			desc: "failed_empty_username",
			arg: auth.CreateUserParams{
				Email:          generator.CreateRandomEmail(generator.CreateRandomString(5)),
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		}, {
			desc: "failed_empty_email",
			arg: auth.CreateUserParams{
				Username:       generator.CreateRandomEmail(generator.CreateRandomString(5)),
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		}, {
			desc: "failed_empty_hashed_password",
			arg: auth.CreateUserParams{
				Username: generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
			},
			err: true,
		}, {
			desc: "failed_duplicate_email",
			arg: auth.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          email,
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.CreateUser(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.arg.Username, res.Username)
				assert.Equal(t, tC.arg.Email, res.Email)
				assert.Equal(t, tC.arg.HashedPassword, res.HashedPassword)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	user := createRandomUser(t)

	testCases := []struct {
		desc  string
		email string
		err   bool
	}{
		{
			desc:  "success",
			email: user.Email,
			err:   false,
		},
		{
			desc:  "failed_empty_email",
			email: "",
			err:   true,
		}, {
			desc:  "failed_invalid_email",
			email: "av088@mail.com",
			err:   true,
		}, {
			desc:  "failed_typo_email",
			email: "a" + user.Email,
			err:   true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.GetUserByEmail(ctx, tC.email)
			if !tC.err {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, user.Email, res.Email)
				assert.Equal(t, user.Username, res.Username)
				assert.Equal(t, user.HashedPassword, res.HashedPassword)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	user := createRandomUser(t)

	usernameSuccess := generator.CreateRandomString(7)
	hashedPasswordSuccess := generator.CreateRandomString(60)
	hashedPasswordEmptyUsername := generator.CreateRandomString(60)
	usernameEmptyHashedPassword := generator.CreateRandomString(60)
	testCases := []struct {
		desc string
		arg  auth.UpdateUserParams
		ans  auth.User
		err  bool
	}{
		{
			desc: "success",
			arg: auth.UpdateUserParams{
				ID:             user.ID,
				Username:       usernameSuccess,
				HashedPassword: hashedPasswordSuccess,
			},
			ans: auth.User{
				ID:             user.ID,
				Username:       usernameSuccess,
				Email:          user.Email,
				HashedPassword: hashedPasswordSuccess,
				IsVerified:     user.IsVerified,
				CreatedAt:      user.CreatedAt,
			},
			err: false,
		}, {
			desc: "failed_empty_username",
			arg: auth.UpdateUserParams{
				ID:             user.ID,
				HashedPassword: hashedPasswordEmptyUsername,
			},
			ans: auth.User{
				ID:             0,
				Username:       "",
				HashedPassword: hashedPasswordEmptyUsername,
			},
			err: true,
		}, {
			desc: "failed_empty_hashed_password",
			arg: auth.UpdateUserParams{
				ID:       user.ID,
				Username: usernameEmptyHashedPassword,
			},
			ans: auth.User{
				ID:             0,
				Username:       usernameEmptyHashedPassword,
				HashedPassword: "",
			},
			err: true,
		}, {
			desc: "failed_empty_arg",
			arg: auth.UpdateUserParams{
				ID: user.ID,
			},
			ans: auth.User{
				ID:             0,
				Username:       "",
				HashedPassword: "",
			},
			err: true,
		}, {
			desc: "failed_wrong_id",
			arg: auth.UpdateUserParams{
				ID:             0,
				Username:       generator.CreateRandomString(7),
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateUser(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, &tC.ans, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	var users []auth.User
	for i := 0; i < 5; i++ {
		user := createRandomUser(t)
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
			err := repoTest.DeleteUser(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestLoadKey(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	res, err := repoTest.LoadKey(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestInsertRefreshToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	testCases := []struct {
		name  string
		user  *auth.User
		token uuid.UUID
		err   bool
	}{
		{
			name:  "succes_1",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "succes_2",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "succes_3",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "error_wrong_id",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   true,
		}, {
			name:  "error_duplicate_uuid",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			if !tC.err {
				err := repoTest.InsertRefreshToken(ctx, tC.user.ID, tC.token)
				require.NoError(t, err)
			}
			if tC.name == "error_wrong_id" {
				err := repoTest.InsertRefreshToken(ctx, 0, tC.token)
				require.Error(t, err)
			}
			if tC.name == "error_duplicate_uuid" {
				err := repoTest.InsertRefreshToken(ctx, tC.user.ID, tC.token)
				require.NoError(t, err)
				err = repoTest.InsertRefreshToken(ctx, tC.user.ID, tC.token)
				require.Error(t, err)
			}
		})
	}
}

func TestReadRefreshToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	tests := []struct {
		name  string
		user  *auth.User
		token uuid.UUID
		err   bool
	}{
		{
			name:  "succes_1",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "succes_2",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "succes_3",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "error_wrong_id",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   true,
		}, {
			name:  "error_duplicate_uuid",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := repoTest.InsertRefreshToken(ctx, test.user.ID, test.token)
			require.NoError(t, err)
			if !test.err {
				res, err := repoTest.ReadRefreshToken(ctx, test.user.ID, test.token)
				require.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, test.user.ID, res.UserID)
				assert.Equal(t, test.token, res.RefreshToken)
			}
			if test.name == "error_wrong_id" {
				res, err := repoTest.ReadRefreshToken(ctx, 0, test.token)
				require.Error(t, err)
				require.Nil(t, res)
			}
			if test.name == "error_wrong_uuid" {
				res, err := repoTest.ReadRefreshToken(ctx, test.user.ID, uuid.New())
				require.Error(t, err)
				require.Nil(t, res)
			}
		})
	}
}

func TestDeleteRefreshToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	testCases := []struct {
		name  string
		user  *auth.User
		token uuid.UUID
		err   bool
	}{
		{
			name:  "succes_1",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "succes_2",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "succes_3",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   false,
		}, {
			name:  "error_wrong_id",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   true,
		}, {
			name:  "error_empty_uuid",
			user:  createRandomUser(t),
			token: uuid.New(),
			err:   true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := repoTest.InsertRefreshToken(ctx, tC.user.ID, tC.token)
			require.NoError(t, err)
			if !tC.err {
				err := repoTest.DeleteRefreshToken(ctx, tC.user.ID)
				require.NoError(t, err)
			}
			if tC.name == "error_wrong_id" {
				err := repoTest.DeleteRefreshToken(ctx, 0)
				require.Error(t, err)
			}
			if tC.name == "error_empty_uuid" {
				err := repoTest.DeleteRefreshToken(ctx, tC.user.ID)
				require.NoError(t, err)
				err = repoTest.DeleteRefreshToken(ctx, tC.user.ID)
				require.Error(t, err)
			}
		})
	}
}
