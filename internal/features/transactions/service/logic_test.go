package service

import (
	"context"
	"os"
	"testing"

	auth "github.com/dwiw96/GoCommerceAPI/internal/features/auth"
	authRepo "github.com/dwiw96/GoCommerceAPI/internal/features/auth/repository"
	products "github.com/dwiw96/GoCommerceAPI/internal/features/products"
	productsRepo "github.com/dwiw96/GoCommerceAPI/internal/features/products/repository"
	transactions "github.com/dwiw96/GoCommerceAPI/internal/features/transactions"
	transactionsRepo "github.com/dwiw96/GoCommerceAPI/internal/features/transactions/repository"
	wallets "github.com/dwiw96/GoCommerceAPI/internal/features/wallets"
	walletsRepo "github.com/dwiw96/GoCommerceAPI/internal/features/wallets/repository"
	testUtils "github.com/dwiw96/GoCommerceAPI/testutils"

	generator "github.com/dwiw96/GoCommerceAPI/pkg/utils/generator"
	errs "github.com/dwiw96/GoCommerceAPI/pkg/utils/responses"

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
	pool = testUtils.GetPool()
	defer pool.Close()
	ctx = testUtils.GetContext()
	defer ctx.Done()

	schemaCleanup := testUtils.SetupDB("test_service_transaction")

	authRepoTest = authRepo.NewAuthRepository(pool, pool)
	productRepoTest = productsRepo.NewProductRepository(pool)
	walletRepoTest = walletsRepo.NewWalletsRepository(pool, ctx)
	repoTest = transactionsRepo.NewTransactionsRepository(pool, pool, ctx)
	serviceTest = NewTransactionsService(ctx, repoTest)

	exitTest := m.Run()

	schemaCleanup()

	os.Exit(exitTest)
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
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	user1, wallet1, product1 := createPreparationTest(t)
	_, product2 := createProductTest(t)

	userID := pgtype.Int4{Int32: user1.ID, Valid: true}
	fromWalletID := pgtype.Int4{Int32: wallet1.ID, Valid: true}
	productID := pgtype.Int4{Int32: product1.ID, Valid: true}
	randQuantity := generator.RandomInt32(1, product1.Availability)
	quantity := pgtype.Int4{Int32: randQuantity, Valid: true}
	tTypePurchase := transactions.TransactionTypesPurchase

	testCases := []struct {
		desc      string
		arg       transactions.TransactionParams
		ans       transactions.TransactionHistory
		code      int
		err       error
		isErr     bool
		isSuccess bool
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
			code:      errs.CodeSuccess,
			err:       nil,
			isErr:     false,
			isSuccess: true,
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
			code:      errs.CodeFailedUser,
			err:       errs.ErrInsufficientStock,
			isErr:     true,
			isSuccess: true,
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
			code:      errs.CodeFailedUser,
			err:       errs.ErrInsufficientBalance,
			isErr:     true,
			isSuccess: true,
		}, {
			desc: "failed_wrong_from_wallet_id",
			arg: transactions.TransactionParams{
				FromWalletID: pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				ProductID:    pgtype.Int4{Int32: product1.ID, Valid: true},
				Quantity:     quantity,
				TType:        tTypePurchase,
			},
			code:      errs.CodeFailedUser,
			err:       errs.ErrCheckConstraint,
			isErr:     true,
			isSuccess: false,
		}, {
			desc: "failed_wrong_product_id",
			arg: transactions.TransactionParams{
				FromWalletID: fromWalletID,
				ProductID:    pgtype.Int4{Int32: product1.ID + 5, Valid: true},
				Quantity:     quantity,
				TType:        tTypePurchase,
			},
			code:      errs.CodeFailedUser,
			err:       errs.ErrCheckConstraint,
			isErr:     true,
			isSuccess: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.PurchaseProduct(tC.arg)
			assert.Equal(t, tC.code, code)
			if !tC.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			if tC.isSuccess {
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.ans.FromWalletID, res.FromWalletID)
				assert.Equal(t, tC.ans.ToWalletID, res.ToWalletID)
				assert.Equal(t, tC.ans.ProductID, res.ProductID)
				assert.Equal(t, tC.ans.Amount, res.Amount)
				assert.Equal(t, tC.ans.Quantity, res.Quantity)
				assert.Equal(t, tC.ans.TType, res.TType)
				assert.Equal(t, tC.ans.TStatus, res.TStatus)
				assert.False(t, res.CreatedAt.Time.IsZero())
			}
		})
	}
}

func TestDepositAndWithdraw(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	user1, wallet1, _ := createPreparationTest(t)

	userID := pgtype.Int4{Int32: user1.ID, Valid: true}
	toWalletID := pgtype.Int4{Int32: wallet1.ID, Valid: true}
	amount := generator.RandomInt32(100, 1000)
	tTypeDeposit := transactions.TransactionTypesDeposit

	testCases := []struct {
		desc      string
		arg       transactions.TransactionParams
		ans       transactions.TransactionHistory
		code      int
		err       error
		isErr     bool
		isSuccess bool
	}{
		{
			desc: "deposit_success_completed",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   toWalletID,
				Amount:       amount,
				TType:        tTypeDeposit,
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
			code:      errs.CodeSuccess,
			err:       nil,
			isErr:     false,
			isSuccess: true,
		}, {
			desc: "deposit_failed_negative_amount",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   toWalletID,
				Amount:       -amount,
				TType:        tTypeDeposit,
			},
			code:      errs.CodeFailedUser,
			err:       errs.ErrLessOrEqualToZero,
			isErr:     true,
			isSuccess: false,
		}, {
			desc: "deposit_failed_invalid_to_wallet_id",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				Amount:       amount,
				TType:        tTypeDeposit,
			},
			code:      errs.CodeFailedUser,
			err:       errs.ErrViolation,
			isErr:     true,
			isSuccess: false,
		}, {
			desc: "deposit_failed_invalid_user_id",
			arg: transactions.TransactionParams{
				UserID:       pgtype.Int4{Int32: user1.ID + 5, Valid: true},
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   toWalletID,
				Amount:       amount,
				TType:        tTypeDeposit,
			},
			code:      errs.CodeFailedUser,
			err:       errs.ErrViolation,
			isErr:     true,
			isSuccess: false,
		}, {
			desc: "withdraw_success_completed",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Int32: wallet1.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Valid: false},
				Amount:       amount,
				TType:        transactions.TransactionTypesWithdrawal,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Int32: wallet1.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Valid: false},
				ProductID:    pgtype.Int4{Valid: false},
				Amount:       -amount,
				Quantity:     pgtype.Int4{Valid: false},
				TType:        transactions.TransactionTypesWithdrawal,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			code:      errs.CodeSuccess,
			err:       nil,
			isErr:     false,
			isSuccess: true,
		}, {
			desc: "withdraw_success_failed_insufficient_balance",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Int32: wallet1.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Valid: false},
				Amount:       wallet1.Balance + amount,
				TType:        transactions.TransactionTypesWithdrawal,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Int32: wallet1.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Valid: false},
				ProductID:    pgtype.Int4{Valid: false},
				Amount:       -(wallet1.Balance + amount),
				Quantity:     pgtype.Int4{Valid: false},
				TType:        transactions.TransactionTypesWithdrawal,
				TStatus:      transactions.TransactionStatusFailed,
			},
			code:      errs.CodeFailedUser,
			err:       errs.ErrInsufficientBalance,
			isErr:     true,
			isSuccess: true,
		}, {
			desc: "withdraw_failed_negative_amount",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Int32: wallet1.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Valid: false},
				Amount:       -amount,
				TType:        transactions.TransactionTypesWithdrawal,
			},
			code:      errs.CodeFailedUser,
			err:       errs.ErrLessOrEqualToZero,
			isErr:     true,
			isSuccess: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.DepositOrWithdraw(tC.arg)
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

func TestTransfer(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	user1, wallet1, _ := createPreparationTest(t)
	_, wallet2, _ := createPreparationTest(t)

	userID := pgtype.Int4{Int32: user1.ID, Valid: true}
	wallet1ID := pgtype.Int4{Int32: wallet1.ID, Valid: true}
	wallet2ID := pgtype.Int4{Int32: wallet2.ID, Valid: true}
	amount := generator.RandomInt32(100, 1000)
	transferType := transactions.TransactionTypesTransfer
	failedStatus := transactions.TransactionStatusFailed

	testCases := []struct {
		desc      string
		arg       transactions.TransactionParams
		ans       transactions.TransactionHistory
		code      int
		isSuccess bool
		isErr     bool
	}{
		{
			desc: "success_completed",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: wallet1ID,
				ToWalletID:   wallet2ID,
				Amount:       amount,
				TType:        transferType,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: wallet1ID,
				ToWalletID:   wallet2ID,
				ProductID:    pgtype.Int4{Valid: false},
				Amount:       amount,
				Quantity:     pgtype.Int4{Valid: false},
				TType:        transferType,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			code:      errs.CodeSuccess,
			isSuccess: true,
			isErr:     false,
		}, {
			desc: "success_failed_insufficient_balance",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: wallet2ID,
				ToWalletID:   wallet1ID,
				Amount:       wallet2.Balance + amount + 1,
				TType:        transferType,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: wallet2ID,
				ToWalletID:   wallet1ID,
				ProductID:    pgtype.Int4{Valid: false},
				Amount:       wallet1.Balance + amount + 1,
				Quantity:     pgtype.Int4{Valid: false},
				TType:        transferType,
				TStatus:      failedStatus,
			},
			code:      errs.CodeFailedUser,
			isSuccess: true,
			isErr:     true,
		}, {
			desc: "error_not_found_from_wallet",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				ToWalletID:   wallet2ID,
				Amount:       wallet1.Balance + amount,
				TType:        transferType,
			},
			code:      errs.CodeFailedUser,
			isSuccess: false,
			isErr:     true,
		}, {
			desc: "error_not_found_to_wallet",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: wallet1ID,
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID + 5, Valid: true},
				Amount:       wallet1.Balance + amount,
				TType:        transferType,
			},
			code:      errs.CodeFailedUser,
			isSuccess: false,
			isErr:     true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.Transfer(tC.arg)
			assert.Equal(t, tC.code, code)
			if !tC.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			if tC.isSuccess {
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.ans.FromWalletID, res.FromWalletID)
				assert.Equal(t, tC.ans.ToWalletID, res.ToWalletID)
				assert.Equal(t, tC.ans.ProductID, res.ProductID)
				assert.Equal(t, tC.ans.Amount, res.Amount)
				assert.Equal(t, tC.ans.Quantity, res.Quantity)
				assert.Equal(t, tC.ans.TType, res.TType)
				assert.Equal(t, tC.ans.TStatus, res.TStatus)
				assert.False(t, res.CreatedAt.Time.IsZero())
			}
		})
	}
}
