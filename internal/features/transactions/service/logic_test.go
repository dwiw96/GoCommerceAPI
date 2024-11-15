package service

import (
	"context"
	"os"
	"testing"

	cfg "github.com/dwiw96/vocagame-technical-test-backend/config"
	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	authRepo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/repository"
	products "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products"
	productsRepo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products/repository"
	transactions "github.com/dwiw96/vocagame-technical-test-backend/internal/features/transactions"
	transactionsRepo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/transactions/repository"
	wallets "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets"
	walletsRepo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets/repository"
	pg "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/postgresql"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"
	errs "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/responses"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	serviceTest     transactions.IService
	repoTest        transactions.IRepository
	ctx             context.Context
	pool            *pgxpool.Pool
	productRepoTest products.IRepository
	walletRepoTest  wallets.IRepository
	authRepoTest    auth.IRepository
)

func TestMain(m *testing.M) {
	env := cfg.GetEnvConfig()
	pool = pg.ConnectToPg(env)
	defer pool.Close()
	ctx = context.Background()
	defer ctx.Done()

	authRepoTest = authRepo.NewAuthRepository(pool, pool)
	productRepoTest = productsRepo.NewProductRepository(pool)
	walletRepoTest = walletsRepo.NewWalletsRepository(pool, ctx)
	repoTest = transactionsRepo.NewTransactionsRepository(pool, pool, ctx)
	serviceTest = NewTransactionsService(ctx, repoTest)

	os.Exit(m.Run())
}

func createRandomUser(t *testing.T) (res *auth.User) {
	username := generator.CreateRandomString(generator.RandomInt(3, 13))
	arg := auth.CreateUserParams{
		Username:       username,
		Email:          generator.CreateRandomEmail(username),
		HashedPassword: generator.CreateRandomString(generator.RandomInt(20, 20)),
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

func createWalletTest(t *testing.T, user *auth.User) (input wallets.CreateWalletParams, res *wallets.Wallet) {
	arg := wallets.CreateWalletParams{
		UserID:  user.ID,
		Balance: 1000,
	}

	res, err := walletRepoTest.CreateWallet(arg)
	require.NoError(t, err)
	assert.NotZero(t, res.ID)
	assert.Equal(t, user.ID, res.UserID)
	assert.Equal(t, arg.Balance, res.Balance)
	assert.False(t, res.CreatedAt.IsZero())
	assert.False(t, res.UpdatedAt.IsZero())

	return arg, res
}

func createProductTest(t *testing.T) (input products.CreateProductParams, res *products.Product) {
	arg := products.CreateProductParams{
		Name:         generator.CreateRandomString(7),
		Description:  generator.CreateRandomString(50),
		Price:        int32(20),
		Availability: int32(50),
	}

	res, err := productRepoTest.CreateProduct(ctx, arg)
	require.NoError(t, err)
	assert.Equal(t, arg.Name, res.Name)
	assert.Equal(t, arg.Description, res.Description)
	assert.Equal(t, arg.Price, res.Price)
	assert.Equal(t, arg.Availability, res.Availability)

	return arg, res
}

func createPreparationTest(t *testing.T) (user *auth.User, wallet *wallets.Wallet, product *products.Product) {
	user = createRandomUser(t)
	_, wallet = createWalletTest(t, user)
	_, product = createProductTest(t)

	return
}

func TestTransactionPurchaseProduct(t *testing.T) {
	user1, wallet1, product1 := createPreparationTest(t)
	t.Log("user1:", user1)
	t.Log("wallet1:", wallet1)
	t.Log("product1:", product1)
	_, product2 := createProductTest(t)
	t.Log("product2:", product2)

	userID := pgtype.Int4{Int32: user1.ID, Valid: true}
	fromWalletID := pgtype.Int4{Int32: wallet1.ID, Valid: true}
	productID := pgtype.Int4{Int32: product1.ID, Valid: true}
	randQuantity := generator.RandomInt32(1, product1.Availability)
	quantity := pgtype.Int4{Int32: randQuantity, Valid: true}
	tTypePurchase := transactions.TransactionTypesPurchase
	t.Log("QUANTITIY:", quantity)

	testCases := []struct {
		desc  string
		arg   transactions.TransactionParams
		ans   transactions.TransactionHistory
		code  int
		err   error
		isErr bool
	}{
		{
			desc: "success_completed",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: fromWalletID,
				ProductID:    productID,
				Quantity:     quantity,
				TType:        tTypePurchase,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: fromWalletID,
				ToWalletID:   pgtype.Int4{Valid: false},
				ProductID:    productID,
				Amount:       product1.Price * randQuantity,
				Quantity:     quantity,
				TType:        transactions.TransactionTypesPurchase,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc: "success_failed_insufficient_stock",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: fromWalletID,
				ProductID:    pgtype.Int4{Int32: product2.ID, Valid: true},
				Quantity:     pgtype.Int4{Int32: product2.Availability + 1, Valid: true},
				TType:        tTypePurchase,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: fromWalletID,
				ToWalletID:   pgtype.Int4{Valid: false},
				ProductID:    pgtype.Int4{Int32: product2.ID, Valid: true},
				Amount:       product1.Price * (product1.Availability + 1),
				Quantity:     pgtype.Int4{Int32: product2.Availability + 1, Valid: true},
				TType:        transactions.TransactionTypesPurchase,
				TStatus:      transactions.TransactionStatusFailed,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc: "success_failed_insufficient_balance",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: fromWalletID,
				ProductID:    pgtype.Int4{Int32: product2.ID, Valid: true},
				Quantity:     pgtype.Int4{Int32: product2.Availability, Valid: true},
				TType:        tTypePurchase,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: fromWalletID,
				ToWalletID:   pgtype.Int4{Valid: false},
				ProductID:    pgtype.Int4{Int32: product2.ID, Valid: true},
				Amount:       product2.Price * product2.Availability,
				Quantity:     pgtype.Int4{Int32: product2.Availability, Valid: true},
				TType:        transactions.TransactionTypesPurchase,
				TStatus:      transactions.TransactionStatusFailed,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc: "failed_wrong_from_wallet_id",
			arg: transactions.TransactionParams{
				FromWalletID: pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				ProductID:    pgtype.Int4{Int32: product1.ID, Valid: true},
				Quantity:     quantity,
				TType:        tTypePurchase,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrCheckConstraint,
			isErr: true,
		}, {
			desc: "failed_wrong_product_id",
			arg: transactions.TransactionParams{
				FromWalletID: fromWalletID,
				ProductID:    pgtype.Int4{Int32: product1.ID + 5, Valid: true},
				Quantity:     quantity,
				TType:        tTypePurchase,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrCheckConstraint,
			isErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.PurchaseProduct(tC.arg)
			assert.Equal(t, tC.code, code)
			if !tC.isErr {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.ans.FromWalletID, res.FromWalletID)
				assert.Equal(t, tC.ans.ToWalletID, res.ToWalletID)
				assert.Equal(t, tC.ans.ProductID, res.ProductID)
				assert.Equal(t, tC.ans.Amount, res.Amount)
				assert.Equal(t, tC.ans.Quantity, res.Quantity)
				assert.Equal(t, tC.ans.TType, res.TType)
				assert.Equal(t, tC.ans.TStatus, res.TStatus)
				assert.False(t, res.CreatedAt.Time.IsZero())
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDeposit(t *testing.T) {
	user1, wallet1, _ := createPreparationTest(t)
	t.Log("wallet1:", wallet1)

	userID := pgtype.Int4{Int32: user1.ID, Valid: true}
	toWalletID := pgtype.Int4{Int32: wallet1.ID, Valid: true}
	amount := generator.RandomInt32(100, 1000)
	tTypeDeposit := transactions.TransactionTypesDeposit

	testCases := []struct {
		desc  string
		arg   transactions.TransactionParams
		ans   transactions.TransactionHistory
		code  int
		err   error
		isErr bool
	}{
		{
			desc: "success_completed",
			arg: transactions.TransactionParams{
				UserID:     userID,
				ToWalletID: toWalletID,
				Amount:     amount,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   toWalletID,
				ProductID:    pgtype.Int4{Valid: false},
				Amount:       amount,
				Quantity:     pgtype.Int4{Valid: false},
				TType:        tTypeDeposit,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			code:  errs.CodeSuccess,
			err:   nil,
			isErr: false,
		}, {
			desc: "failed_negative_amount",
			arg: transactions.TransactionParams{
				UserID:     userID,
				ToWalletID: toWalletID,
				Amount:     -amount,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrLessOrEqualToZero,
			isErr: true,
		}, {
			desc: "failed_invalid_to_wallet_id",
			arg: transactions.TransactionParams{
				UserID:     userID,
				ToWalletID: pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				Amount:     amount,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrViolation,
			isErr: true,
		}, {
			desc: "failed_invalid_user_id",
			arg: transactions.TransactionParams{
				UserID:     pgtype.Int4{Int32: user1.ID + 5, Valid: true},
				ToWalletID: toWalletID,
				Amount:     amount,
			},
			code:  errs.CodeFailedUser,
			err:   errs.ErrViolation,
			isErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.Deposit(tC.arg)
			assert.Equal(t, tC.code, code)
			if !tC.isErr {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.ans.FromWalletID, res.FromWalletID)
				assert.Equal(t, tC.ans.ToWalletID, res.ToWalletID)
				assert.Equal(t, tC.ans.ProductID, res.ProductID)
				assert.Equal(t, tC.ans.Amount, res.Amount)
				assert.Equal(t, tC.ans.Quantity, res.Quantity)
				assert.Equal(t, tC.ans.TType, res.TType)
				assert.Equal(t, tC.ans.TStatus, res.TStatus)
				assert.False(t, res.CreatedAt.Time.IsZero())
			} else {
				require.Error(t, err)
			}
		})
	}
}
