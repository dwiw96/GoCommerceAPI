package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	cfg "github.com/dwiw96/vocagame-technical-test-backend/config"
	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	authRepo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/repository"
	wallets "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets"
	walletsRepo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets/repository"
	pg "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/postgresql"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"
	errs "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/responses"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	serviceTest  wallets.IService
	ctx          context.Context
	pool         *pgxpool.Pool
	authRepoTest auth.IRepository
)

func TestMain(m *testing.M) {
	env := cfg.GetEnvConfig()
	pool = pg.ConnectToPg(env)
	defer pool.Close()
	ctx = context.Background()
	defer ctx.Done()

	repo := walletsRepo.NewWalletsRepository(pool, ctx)
	authRepoTest = authRepo.NewAuthRepository(pool, pool)
	serviceTest = NewWalletsService(ctx, repo)

	os.Exit(m.Run())
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

	res, err := authRepoTest.CreateUser(ctx, arg)
	require.NoError(t, err)
	assert.NotZero(t, res.ID)
	assert.Equal(t, username, res.Username)
	assert.Equal(t, arg.Email, res.Email)
	assert.Equal(t, arg.HashedPassword, res.HashedPassword)

	return res
}

func createWalletTest(t *testing.T) (input wallets.CreateWalletParams, res *wallets.Wallet) {
	user := createRandomUser(t)

	arg := wallets.CreateWalletParams{
		UserID:  user.ID,
		Balance: int32(generator.RandomInt(1000, 5000)),
	}

	res, code, err := serviceTest.CreateWallet(arg)
	require.NoError(t, err)
	assert.Equal(t, errs.CodeSuccessCreate, code)
	assert.NotZero(t, res.ID)
	assert.Equal(t, user.ID, res.UserID)
	assert.Equal(t, arg.Balance, res.Balance)
	assert.False(t, res.CreatedAt.IsZero())
	assert.False(t, res.UpdatedAt.IsZero())

	return arg, res
}

func TestCreateWallet(t *testing.T) {
	user := createRandomUser(t)
	user2 := createRandomUser(t)

	testCases := []struct {
		desc  string
		args  wallets.CreateWalletParams
		code  int
		err   error
		isErr bool
	}{
		{
			desc: "success_all_params",
			args: wallets.CreateWalletParams{
				UserID:  user.ID,
				Balance: int32(generator.RandomInt(0, 50)),
			},
			code:  errs.CodeSuccessCreate,
			err:   nil,
			isErr: false,
		}, {
			desc: "failed_duplicate_user_id",
			args: wallets.CreateWalletParams{
				UserID:  user.ID,
				Balance: int32(generator.RandomInt(0, 50)),
			},
			code:  errs.CodeFailedDuplicated,
			err:   errs.ErrDuplicate,
			isErr: true,
		}, {
			desc: "failed_without_user_id",
			args: wallets.CreateWalletParams{
				Balance: int32(generator.RandomInt(0, 50)),
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrViolation,
			isErr: true,
		}, {
			desc: "failed_not_found_user_id",
			args: wallets.CreateWalletParams{
				UserID:  0,
				Balance: int32(generator.RandomInt(0, 50)),
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrViolation,
			isErr: true,
		}, {
			desc: "failed_minus_balance",
			args: wallets.CreateWalletParams{
				UserID:  user2.ID,
				Balance: int32(generator.RandomInt(-5000, -1)),
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrBalanceLessThanZero,
			isErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(*testing.T) {
			res, code, err := serviceTest.CreateWallet(tC.args)
			assert.Equal(t, tC.code, code)
			assert.Equal(t, tC.err, err)
			if !tC.isErr {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.args.UserID, res.UserID)
				assert.Equal(t, tC.args.Balance, res.Balance)
				assert.False(t, res.CreatedAt.IsZero())
				assert.False(t, res.UpdatedAt.IsZero())
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetWalletByUserID(t *testing.T) {
	walletArg, wallet := createWalletTest(t)
	testCases := []struct {
		desc   string
		userID int32
		ans    wallets.Wallet
		code   int
		err    error
		isErr  bool
	}{
		{
			desc:   "success",
			userID: walletArg.UserID,
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance,
				CreatedAt: wallet.CreatedAt,
				UpdatedAt: wallet.UpdatedAt,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc:   "failed_not_found_user_id",
			userID: walletArg.UserID + 5,
			code:   errs.CodeFailedUser,
			err:    errs.ErrNoData,
			isErr:  true,
		}, {
			desc:  "failed_without_user_id",
			code:  errs.CodeFailedUser,
			err:   errs.ErrNoData,
			isErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.GetWalletByUserID(tC.userID)
			assert.Equal(t, tC.code, code)
			assert.Equal(t, tC.err, err)
			if !tC.isErr {
				require.NoError(t, err)
				assert.Equal(t, wallet.ID, res.ID)
				assert.Equal(t, wallet.UserID, res.UserID)
				assert.Equal(t, wallet.Balance, res.Balance)
				assert.Equal(t, wallet.CreatedAt, res.CreatedAt)
				assert.Equal(t, wallet.UpdatedAt, res.UpdatedAt)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDepositToWallet(t *testing.T) {
	walletArg, wallet := createWalletTest(t)
	deposit := int32(generator.RandomInt(500, 5000))
	withdraw := int32(generator.RandomInt(-500, -1))

	testCases := []struct {
		desc  string
		args  wallets.UpdateWalletParams
		ans   wallets.Wallet
		code  int
		err   error
		isErr bool
	}{
		{
			desc: "success",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: deposit,
			},
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance + deposit,
				CreatedAt: wallet.CreatedAt,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc: "failed_zero_amount",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: 0,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrInvalidInput,
			isErr: true,
		}, {
			desc: "failed_minus_amount",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: withdraw,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrInvalidInput,
			isErr: true,
		}, {
			desc: "failed_not_found_user_id",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID + 5,
				Amount: deposit,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrNoData,
			isErr: true,
		}, {
			desc: "failed_without_user_id",
			args: wallets.UpdateWalletParams{
				Amount: deposit,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrNoData,
			isErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.DepositToWallet(tC.args)
			assert.Equal(t, tC.code, code)
			assert.Equal(t, tC.err, err)
			if !tC.isErr {
				require.NoError(t, err)
				assert.Equal(t, tC.ans.ID, res.ID)
				assert.Equal(t, tC.ans.UserID, res.UserID)
				assert.Equal(t, tC.ans.Balance, res.Balance)
				assert.Equal(t, tC.ans.CreatedAt, res.CreatedAt)
				assert.True(t, res.UpdatedAt.After(wallet.UpdatedAt))
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestWithdrawFromWallet(t *testing.T) {
	walletArg, wallet := createWalletTest(t)
	withdraw := int32(generator.RandomInt(1, 400))
	minusBalance := (wallet.Balance + withdraw) * 2

	testCases := []struct {
		desc  string
		args  wallets.UpdateWalletParams
		ans   wallets.Wallet
		code  int
		err   error
		isErr bool
	}{
		{
			desc: "success_1",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: withdraw,
			},
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance - withdraw,
				CreatedAt: wallet.CreatedAt,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc: "success_2",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: withdraw,
			},
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance - (withdraw * 2),
				CreatedAt: wallet.CreatedAt,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc: "failed_zero_amount",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: 0,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrInvalidInput,
			isErr: true,
		}, {
			desc: "failed_minus_balance",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: minusBalance,
			},
			code:  errs.CodeFailedUser,
			err:   fmt.Errorf("balance minimum is 0"),
			isErr: true,
		}, {
			desc: "failed_not_found_user_id",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID + 5,
				Amount: withdraw,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrNoData,
			isErr: true,
		}, {
			desc: "failed_without_user_id",
			args: wallets.UpdateWalletParams{
				Amount: withdraw,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrNoData,
			isErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.WithdrawFromWallet(tC.args)
			assert.Equal(t, tC.code, code)
			assert.Equal(t, tC.err, err)
			if !tC.isErr {
				require.NoError(t, err)
				assert.Equal(t, tC.ans.ID, res.ID)
				assert.Equal(t, tC.ans.UserID, res.UserID)
				assert.Equal(t, tC.ans.Balance, res.Balance)
				assert.Equal(t, tC.ans.CreatedAt, res.CreatedAt)
				assert.True(t, res.UpdatedAt.After(wallet.UpdatedAt))
			} else {
				require.Error(t, err)
			}
		})
	}
}
