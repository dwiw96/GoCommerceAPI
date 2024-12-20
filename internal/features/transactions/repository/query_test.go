package repository

import (
	"context"
	"os"
	"testing"

	testUtils "github.com/dwiw96/GoCommerceAPI/testutils"
	// cfg "github.com/dwiw96/GoCommerceAPI/config"
	auth "github.com/dwiw96/GoCommerceAPI/internal/features/auth"
	authRepo "github.com/dwiw96/GoCommerceAPI/internal/features/auth/repository"
	products "github.com/dwiw96/GoCommerceAPI/internal/features/products"
	productsRepo "github.com/dwiw96/GoCommerceAPI/internal/features/products/repository"
	transactions "github.com/dwiw96/GoCommerceAPI/internal/features/transactions"
	wallets "github.com/dwiw96/GoCommerceAPI/internal/features/wallets"
	walletsRepo "github.com/dwiw96/GoCommerceAPI/internal/features/wallets/repository"

	// pg "github.com/dwiw96/GoCommerceAPI/pkg/driver/postgresql"
	generator "github.com/dwiw96/GoCommerceAPI/pkg/utils/generator"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
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

	schemaCleanup := testUtils.SetupDB("test_repo_transaction")

	authRepoTest = authRepo.NewAuthRepository(pool, pool)
	productRepoTest = productsRepo.NewProductRepository(pool)
	walletRepoTest = walletsRepo.NewWalletsRepository(pool, ctx)
	repoTest = NewTransactionsRepository(pool, pool, ctx)

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

func TestCreateTransaction(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	_, wallet1, product1 := createPreparationTest(t)
	_, wallet2, _ := createPreparationTest(t)

	quantity := generator.RandomInt32(1, 10)
	fromWalletID := pgtype.Int4{Int32: wallet1.ID, Valid: true}
	toWalletID := pgtype.Int4{Int32: wallet2.ID, Valid: true}
	amount := generator.RandomInt32(100, wallet1.Balance)
	Quantity := pgtype.Int4{Int32: quantity, Valid: true}
	tTypePurchase := transactions.TransactionTypesPurchase
	tTypeTransfer := transactions.TransactionTypesTransfer
	tStatusCompleted := transactions.TransactionStatusCompleted

	testCases := []struct {
		desc  string
		arg   transactions.CreateTransactionParams
		ans   transactions.TransactionHistory
		isErr bool
	}{
		{
			desc: "success_purchase",
			arg: transactions.CreateTransactionParams{
				Amount:       product1.Price * quantity,
				FromWalletID: fromWalletID,
				ProductID:    pgtype.Int4{Int32: product1.ID, Valid: true},
				Quantity:     pgtype.Int4{Int32: quantity, Valid: true},
				TType:        tTypePurchase,
				TStatus:      tStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: fromWalletID,
				ToWalletID:   pgtype.Int4{Int32: 0, Valid: false},
				ProductID:    pgtype.Int4{Int32: product1.ID, Valid: true},
				Amount:       product1.Price * quantity,
				Quantity:     Quantity,
				TType:        tTypePurchase,
				TStatus:      tStatusCompleted,
			},
			isErr: false,
		}, {
			desc: "success_transfer",
			arg: transactions.CreateTransactionParams{
				Amount:       amount,
				FromWalletID: fromWalletID,
				ToWalletID:   toWalletID,
				TType:        tTypeTransfer,
				TStatus:      tStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: fromWalletID,
				ToWalletID:   toWalletID,
				ProductID:    pgtype.Int4{Int32: 0, Valid: false},
				Amount:       amount,
				Quantity:     pgtype.Int4{Int32: 0, Valid: false},
				TType:        tTypeTransfer,
				TStatus:      tStatusCompleted,
			},
			isErr: false,
		}, {
			desc: "success_deposit",
			arg: transactions.CreateTransactionParams{
				Amount:       amount,
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				TType:        transactions.TransactionTypesDeposit,
				TStatus:      tStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ProductID:    pgtype.Int4{Int32: 0, Valid: false},
				Amount:       amount,
				Quantity:     pgtype.Int4{Int32: 0, Valid: false},
				TType:        transactions.TransactionTypesDeposit,
				TStatus:      tStatusCompleted,
			},
			isErr: false,
		}, {
			desc: "success_withdraw",
			arg: transactions.CreateTransactionParams{
				Amount:       -amount,
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				TType:        transactions.TransactionTypesWithdrawal,
				TStatus:      tStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ProductID:    pgtype.Int4{Int32: 0, Valid: false},
				Amount:       -amount,
				Quantity:     pgtype.Int4{Int32: 0, Valid: false},
				TType:        transactions.TransactionTypesWithdrawal,
				TStatus:      tStatusCompleted,
			},
			isErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.CreateTransaction(tC.arg)
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
				require.NoError(t, err)
			}
		})
	}
}

func createTransactionTest(t *testing.T) (purchase, transfer, deposit, withdrawal *transactions.TransactionHistory) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	_, wallet1, product1 := createPreparationTest(t)
	_, wallet2, _ := createPreparationTest(t)

	quantity := generator.RandomInt32(1, 10)
	fromWalletID := pgtype.Int4{Int32: wallet1.ID, Valid: true}
	toWalletID := pgtype.Int4{Int32: wallet2.ID, Valid: true}
	amount := generator.RandomInt32(100, wallet1.Balance)
	Quantity := pgtype.Int4{Int32: quantity, Valid: true}
	tTypePurchase := transactions.TransactionTypesPurchase
	tTypeTransfer := transactions.TransactionTypesTransfer
	tStatusPending := transactions.TransactionStatusPending

	testCases := []struct {
		desc  string
		arg   transactions.CreateTransactionParams
		ans   transactions.TransactionHistory
		isErr bool
	}{
		{
			desc: "success_purchase",
			arg: transactions.CreateTransactionParams{
				Amount:       product1.Price * quantity,
				FromWalletID: fromWalletID,
				ProductID:    pgtype.Int4{Int32: product1.ID, Valid: true},
				Quantity:     pgtype.Int4{Int32: quantity, Valid: true},
				TType:        tTypePurchase,
				TStatus:      tStatusPending,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: fromWalletID,
				ToWalletID:   pgtype.Int4{Int32: 0, Valid: false},
				ProductID:    pgtype.Int4{Int32: product1.ID, Valid: true},
				Amount:       product1.Price * quantity,
				Quantity:     Quantity,
				TType:        tTypePurchase,
				TStatus:      tStatusPending,
			},
			isErr: false,
		}, {
			desc: "success_transfer",
			arg: transactions.CreateTransactionParams{
				Amount:       amount,
				FromWalletID: fromWalletID,
				ToWalletID:   toWalletID,
				TType:        tTypeTransfer,
				TStatus:      tStatusPending,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: fromWalletID,
				ToWalletID:   toWalletID,
				ProductID:    pgtype.Int4{Int32: 0, Valid: false},
				Amount:       amount,
				Quantity:     pgtype.Int4{Int32: 0, Valid: false},
				TType:        tTypeTransfer,
				TStatus:      tStatusPending,
			},
			isErr: false,
		}, {
			desc: "success_deposit",
			arg: transactions.CreateTransactionParams{
				Amount:       amount,
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				TType:        transactions.TransactionTypesDeposit,
				TStatus:      tStatusPending,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ProductID:    pgtype.Int4{Int32: 0, Valid: false},
				Amount:       amount,
				Quantity:     pgtype.Int4{Int32: 0, Valid: false},
				TType:        transactions.TransactionTypesDeposit,
				TStatus:      tStatusPending,
			},
			isErr: false,
		}, {
			desc: "success_withdraw",
			arg: transactions.CreateTransactionParams{
				Amount:       -amount,
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				TType:        transactions.TransactionTypesWithdrawal,
				TStatus:      tStatusPending,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Int32: wallet2.ID, Valid: true},
				ProductID:    pgtype.Int4{Int32: 0, Valid: false},
				Amount:       -amount,
				Quantity:     pgtype.Int4{Int32: 0, Valid: false},
				TType:        transactions.TransactionTypesWithdrawal,
				TStatus:      tStatusPending,
			},
			isErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.CreateTransaction(tC.arg)
			switch tC.arg.TType {
			case transactions.TransactionTypesPurchase:
				purchase = res
			case transactions.TransactionTypesTransfer:
				transfer = res
			case transactions.TransactionTypesDeposit:
				deposit = res
			case transactions.TransactionTypesWithdrawal:
				withdrawal = res
			}
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
				require.NoError(t, err)
			}
		})
	}

	return
}

func TestUpdateTransaction(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	purchase, transfer, deposit, withdrawal := createTransactionTest(t)
	require.Equal(t, transactions.TransactionStatusPending, purchase.TStatus)
	require.Equal(t, transactions.TransactionStatusPending, transfer.TStatus)
	require.Equal(t, transactions.TransactionStatusPending, deposit.TStatus)
	require.Equal(t, transactions.TransactionStatusPending, withdrawal.TStatus)

	amount := generator.RandomInt32(10, 100)

	testCases := []struct {
		desc  string
		arg   transactions.UpdateTransactionStatusParams
		ans   transactions.TransactionHistory
		isErr bool
	}{
		{
			desc: "success_purchase",
			arg: transactions.UpdateTransactionStatusParams{
				ID:      purchase.ID,
				Amount:  amount,
				TStatus: transactions.TransactionStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: purchase.FromWalletID,
				ToWalletID:   purchase.ToWalletID,
				ProductID:    purchase.ProductID,
				Amount:       amount,
				Quantity:     purchase.Quantity,
				TType:        transactions.TransactionTypesPurchase,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			isErr: false,
		}, {
			desc: "success_transfer",
			arg: transactions.UpdateTransactionStatusParams{
				ID:      transfer.ID,
				TStatus: transactions.TransactionStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: transfer.FromWalletID,
				ToWalletID:   transfer.ToWalletID,
				ProductID:    transfer.ProductID,
				Amount:       0,
				Quantity:     transfer.Quantity,
				TType:        transactions.TransactionTypesTransfer,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			isErr: false,
		}, {
			desc: "success_deposit",
			arg: transactions.UpdateTransactionStatusParams{
				ID:      deposit.ID,
				TStatus: transactions.TransactionStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: deposit.FromWalletID,
				ToWalletID:   deposit.ToWalletID,
				ProductID:    deposit.ProductID,
				Amount:       0,
				Quantity:     deposit.Quantity,
				TType:        transactions.TransactionTypesDeposit,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			isErr: false,
		}, {
			desc: "success_withdraw",
			arg: transactions.UpdateTransactionStatusParams{
				ID:      withdrawal.ID,
				TStatus: transactions.TransactionStatusCompleted,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: withdrawal.FromWalletID,
				ToWalletID:   withdrawal.ToWalletID,
				ProductID:    withdrawal.ProductID,
				Amount:       0,
				Quantity:     withdrawal.Quantity,
				TType:        transactions.TransactionTypesWithdrawal,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			isErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateTransactionStatus(tC.arg)
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
				require.NoError(t, err)
			}
		})
	}
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
		isSuccess bool
		isErr     bool
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
			isSuccess: true,
			isErr:     false,
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
			isSuccess: true,
			isErr:     true,
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
			isSuccess: true,
			isErr:     true,
		}, {
			desc: "failed_wrong_from_wallet_id",
			arg: transactions.TransactionParams{
				FromWalletID: pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				ProductID:    pgtype.Int4{Int32: product1.ID, Valid: true},
				Quantity:     quantity,
				TType:        tTypePurchase,
			},
			isSuccess: false,
			isErr:     true,
		}, {
			desc: "failed_wrong_product_id",
			arg: transactions.TransactionParams{
				FromWalletID: fromWalletID,
				ProductID:    pgtype.Int4{Int32: product1.ID + 5, Valid: true},
				Quantity:     quantity,
				TType:        tTypePurchase,
			},
			isSuccess: false,
			isErr:     true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.TransactionPurchaseProduct(tC.arg)
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

func TestTransactionDepositOrWithdraw(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	user1, wallet1, _ := createPreparationTest(t)

	userID := pgtype.Int4{Int32: user1.ID, Valid: true}
	amount := generator.RandomInt32(100, 1000)

	testCases := []struct {
		desc  string
		arg   transactions.TransactionParams
		ans   transactions.TransactionHistory
		isErr bool
	}{
		{
			desc: "deposit_success_completed",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   pgtype.Int4{Int32: wallet1.ID, Valid: true},
				Amount:       amount,
				TType:        transactions.TransactionTypesDeposit,
			},
			ans: transactions.TransactionHistory{
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   pgtype.Int4{Int32: wallet1.ID, Valid: true},
				ProductID:    pgtype.Int4{Valid: false},
				Amount:       amount,
				Quantity:     pgtype.Int4{Valid: false},
				TType:        transactions.TransactionTypesDeposit,
				TStatus:      transactions.TransactionStatusCompleted,
			},
			isErr: false,
		}, {
			desc: "deposit_failed_invalid_to_wallet_id",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Valid: false},
				ToWalletID:   pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				Amount:       amount,
				TType:        transactions.TransactionTypesDeposit,
			},
			isErr: true,
		}, {
			desc: "withdraw_success_completed",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Int32: wallet1.ID, Valid: true},
				ToWalletID:   pgtype.Int4{Valid: false},
				Amount:       -amount,
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
			isErr: false,
		}, {
			desc: "withdraw_failed_invalid_to_wallet_id",
			arg: transactions.TransactionParams{
				UserID:       userID,
				FromWalletID: pgtype.Int4{Int32: wallet1.ID + 5, Valid: true},
				ToWalletID:   pgtype.Int4{Valid: false},
				Amount:       amount,
			},
			isErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.TransactionDepositOrWithdraw(tC.arg)
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

func TestTransactionTransfer(t *testing.T) {
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
			isSuccess: false,
			isErr:     true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.TransactionTransfer(tC.arg)
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
