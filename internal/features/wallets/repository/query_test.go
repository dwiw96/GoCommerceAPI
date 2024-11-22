package repository

import (
	"context"
	"os"
	"testing"

	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	authRepo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/repository"
	wallets "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"
	testUtils "github.com/dwiw96/vocagame-technical-test-backend/testutils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	repoTest     wallets.IRepository
	ctx          context.Context
	pool         *pgxpool.Pool
	authRepoTest auth.IRepository
)

func TestMain(m *testing.M) {
	pool = testUtils.GetPool()
	defer pool.Close()
	ctx = testUtils.GetContext()
	defer ctx.Done()

	schemaCleanup := testUtils.SetupDB("repo_wallet")

	repoTest = NewWalletsRepository(pool, ctx)
	authRepoTest = authRepo.NewAuthRepository(pool, pool)

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
		Balance: int32(generator.RandomInt(0, 5000)),
	}

	res, err := repoTest.CreateWallet(arg)
	require.NoError(t, err)
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
		desc string
		args wallets.CreateWalletParams
		err  bool
	}{
		{
			desc: "success_all_params",
			args: wallets.CreateWalletParams{
				UserID:  user.ID,
				Balance: int32(generator.RandomInt(0, 50)),
			},
			err: false,
		}, {
			desc: "failed_duplicate_user_id",
			args: wallets.CreateWalletParams{
				UserID:  user.ID,
				Balance: int32(generator.RandomInt(0, 50)),
			},
			err: true,
		}, {
			desc: "failed_without_user_id",
			args: wallets.CreateWalletParams{
				Balance: int32(generator.RandomInt(0, 50)),
			},
			err: true,
		}, {
			desc: "failed_wrong_user_id",
			args: wallets.CreateWalletParams{
				UserID:  0,
				Balance: int32(generator.RandomInt(0, 50)),
			},
			err: true,
		}, {
			desc: "failed_minus_balance",
			args: wallets.CreateWalletParams{
				UserID:  user2.ID,
				Balance: int32(generator.RandomInt(-5000, -1)),
			},
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(*testing.T) {
			res, err := repoTest.CreateWallet(tC.args)
			if !tC.err {
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
		err    bool
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
			err: false,
		}, {
			desc:   "failed_wrong_user_id",
			userID: walletArg.UserID + 5,
			err:    true,
		}, {
			desc: "failed_no_user_id",
			err:  true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.GetWalletByUserID(tC.userID)
			if !tC.err {
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

func TestUpdateWalletByUserID(t *testing.T) {
	walletArg, wallet := createWalletTest(t)
	deposit := int32(generator.RandomInt(500, 5000))
	withdraw := int32(generator.RandomInt(-500, -1))
	minusBalance := -(wallet.Balance + deposit + withdraw) * 2

	testCases := []struct {
		desc string
		args wallets.UpdateWalletParams
		ans  wallets.Wallet
		err  bool
	}{
		{
			desc: "success_deposit",
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
			err: false,
		}, {
			desc: "success_withdraw",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: withdraw,
			},
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance + deposit + withdraw,
				CreatedAt: wallet.CreatedAt,
			},
			err: false,
		}, {
			desc: "failed_minus_balance",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID,
				Amount: minusBalance,
			},
			err: true,
		}, {
			desc: "failed_not_found_user_id",
			args: wallets.UpdateWalletParams{
				UserID: walletArg.UserID + 5,
				Amount: minusBalance,
			},
			err: true,
		}, {
			desc: "failed_without_user_id",
			args: wallets.UpdateWalletParams{
				Amount: minusBalance,
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateWalletByUserID(tC.args)
			if !tC.err {
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

func TestGetWalletByID(t *testing.T) {
	_, wallet := createWalletTest(t)
	testCases := []struct {
		desc     string
		walletID int32
		ans      wallets.Wallet
		err      bool
	}{
		{
			desc:     "success",
			walletID: wallet.ID,
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance,
				CreatedAt: wallet.CreatedAt,
				UpdatedAt: wallet.UpdatedAt,
			},
			err: false,
		}, {
			desc:     "failed_wrong_user_id",
			walletID: wallet.ID + 5,
			err:      true,
		}, {
			desc: "failed_no_user_id",
			err:  true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.GetWalletByID(tC.walletID)
			if !tC.err {
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

func TestUpdateWalletByID(t *testing.T) {
	_, wallet := createWalletTest(t)
	deposit := int32(generator.RandomInt(500, 5000))
	withdraw := int32(generator.RandomInt(-500, -1))
	minusBalance := -(wallet.Balance + deposit + withdraw) * 2

	testCases := []struct {
		desc string
		args wallets.UpdateWalletParams
		ans  wallets.Wallet
		err  bool
	}{
		{
			desc: "success_deposit",
			args: wallets.UpdateWalletParams{
				WalletID: wallet.ID,
				Amount:   deposit,
			},
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance + deposit,
				CreatedAt: wallet.CreatedAt,
			},
			err: false,
		}, {
			desc: "success_withdraw",
			args: wallets.UpdateWalletParams{
				WalletID: wallet.ID,
				Amount:   withdraw,
			},
			ans: wallets.Wallet{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance + deposit + withdraw,
				CreatedAt: wallet.CreatedAt,
			},
			err: false,
		}, {
			desc: "failed_minus_balance",
			args: wallets.UpdateWalletParams{
				WalletID: wallet.ID,
				Amount:   minusBalance,
			},
			err: true,
		}, {
			desc: "failed_not_found_wallet_id",
			args: wallets.UpdateWalletParams{
				WalletID: wallet.ID + 5,
				Amount:   minusBalance,
			},
			err: true,
		}, {
			desc: "failed_without_wallet_id",
			args: wallets.UpdateWalletParams{
				Amount: minusBalance,
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateWalletByID(tC.args)
			if !tC.err {
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
