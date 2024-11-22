package repository

import (
	"context"
	"fmt"

	db "github.com/dwiw96/GoCommerceAPI/internal/db"
	products "github.com/dwiw96/GoCommerceAPI/internal/features/products"
	productsRepo "github.com/dwiw96/GoCommerceAPI/internal/features/products/repository"
	transactions "github.com/dwiw96/GoCommerceAPI/internal/features/transactions"
	wallets "github.com/dwiw96/GoCommerceAPI/internal/features/wallets"
	walletsRepo "github.com/dwiw96/GoCommerceAPI/internal/features/wallets/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionsRepository struct {
	db           db.DBTX
	dbTx         *pgxpool.Pool
	ctx          context.Context
	walletsRepo  wallets.IRepository
	productsRepo products.IRepository
}

func NewTransactionsRepository(db db.DBTX, dbTx *pgxpool.Pool, ctx context.Context) transactions.IRepository {
	return &transactionsRepository{
		db:   db,
		ctx:  ctx,
		dbTx: dbTx,
	}
}

type transactionTx struct {
	db *pgxpool.Pool
}

func NewTransactionTx(db *pgxpool.Pool) *transactionTx {
	return &transactionTx{
		db: db,
	}
}

func (r *transactionsRepository) ExecDbTx(fn func(*transactionsRepository) error) error {
	tx, err := r.dbTx.Begin(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to start db transaction, err: %w", err)
	}

	productRepo := productsRepo.NewProductRepository(tx)
	walletRepo := walletsRepo.NewWalletsRepository(tx, r.ctx)

	q := &transactionsRepository{db: tx, ctx: r.ctx, walletsRepo: walletRepo, productsRepo: productRepo}
	err = fn(q)

	defer func() {
		if err != nil {
			tx.Rollback(r.ctx)
		} else {
			tx.Commit(r.ctx)
		}
	}()

	return err
}

const createTransaction = `-- name: CreateTransaction :one
INSERT INTO 
    transaction_histories(
        from_wallet_id,
        to_wallet_id,
        product_id,
        amount,
        quantity,
        t_type,
        t_status
    )
VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, from_wallet_id, to_wallet_id, product_id, amount, quantity, t_type, t_status, created_at
`

func (r *transactionsRepository) CreateTransaction(arg transactions.CreateTransactionParams) (*transactions.TransactionHistory, error) {
	row := r.db.QueryRow(r.ctx, createTransaction,
		arg.FromWalletID,
		arg.ToWalletID,
		arg.ProductID,
		arg.Amount,
		arg.Quantity,
		arg.TType,
		arg.TStatus,
	)
	var i transactions.TransactionHistory
	err := row.Scan(
		&i.ID,
		&i.FromWalletID,
		&i.ToWalletID,
		&i.ProductID,
		&i.Amount,
		&i.Quantity,
		&i.TType,
		&i.TStatus,
		&i.CreatedAt,
	)
	return &i, err
}

const updateTransactionStatus = `-- name: UpdateTransactionStatus :one
UPDATE
    transaction_histories
SET
	amount = $1,
    t_status = $2
WHERE 
    id = $3
RETURNING id, from_wallet_id, to_wallet_id, product_id, amount, quantity, t_type, t_status, created_at
`

func (r *transactionsRepository) UpdateTransactionStatus(arg transactions.UpdateTransactionStatusParams) (*transactions.TransactionHistory, error) {
	row := r.db.QueryRow(r.ctx, updateTransactionStatus, arg.Amount, arg.TStatus, arg.ID)
	var i transactions.TransactionHistory
	err := row.Scan(
		&i.ID,
		&i.FromWalletID,
		&i.ToWalletID,
		&i.ProductID,
		&i.Amount,
		&i.Quantity,
		&i.TType,
		&i.TStatus,
		&i.CreatedAt,
	)
	return &i, err
}

func (t *transactionsRepository) TransactionPurchaseProduct(arg transactions.TransactionParams) (*transactions.TransactionHistory, error) {
	var (
		res    *transactions.TransactionHistory
		amount int32
		err    error
	)

	errCreateTransaction := t.ExecDbTx(func(tr *transactionsRepository) error {
		resGetProduct, err := tr.productsRepo.GetProductByID(tr.ctx, arg.ProductID.Int32)
		if err != nil {
			return fmt.Errorf("failed to get product, err: %w", err)
		}
		amount = resGetProduct.Price * arg.Quantity.Int32

		createTransactionArg := transactions.CreateTransactionParams{
			FromWalletID: arg.FromWalletID,
			ProductID:    arg.ProductID,
			Amount:       0,
			Quantity:     arg.Quantity,
			TType:        transactions.TransactionTypesPurchase,
			TStatus:      transactions.TransactionStatusPending,
		}
		res, err = tr.CreateTransaction(createTransactionArg)
		if err != nil {
			return fmt.Errorf("failed to create transaction, err: %w", err)
		}

		return err
	})
	if errCreateTransaction != nil {
		return nil, errCreateTransaction
	}

	errUpdate := t.ExecDbTx(func(tr *transactionsRepository) error {
		updateProductArg := products.UpdateProductAvailabilityParams{
			ID:           arg.ProductID.Int32,
			Availability: -arg.Quantity.Int32,
		}
		_, err = tr.productsRepo.UpdateProductAvailability(tr.ctx, updateProductArg)
		if err != nil {
			return fmt.Errorf("failed to update product, err: %w", err)
		}

		updateWalletArg := wallets.UpdateWalletParams{
			Amount: -amount,
			UserID: arg.UserID.Int32,
		}
		_, err = tr.walletsRepo.UpdateWalletByUserID(updateWalletArg)
		if err != nil {
			return fmt.Errorf("failed to update wallet, err: %w", err)
		}

		return err
	})

	// change transaction status
	var errUpdateTransaction error
	errUpdateStatus := t.ExecDbTx(func(tr *transactionsRepository) error {
		argUpdateStatus := transactions.UpdateTransactionStatusParams{
			Amount: amount,
			ID:     res.ID,
		}

		if errUpdate == nil {
			argUpdateStatus.TStatus = transactions.TransactionStatusCompleted
		} else {
			argUpdateStatus.TStatus = transactions.TransactionStatusFailed
		}

		res, errUpdateTransaction = tr.UpdateTransactionStatus(argUpdateStatus)
		if errUpdateTransaction != nil {
			return fmt.Errorf("failed to update transaction, err: %w", errUpdateTransaction)
		}

		return errUpdateTransaction
	})

	if errUpdateStatus == nil {
		return res, errUpdate
	}

	return res, errUpdateStatus
}

func (r *transactionsRepository) TransactionDepositOrWithdraw(arg transactions.TransactionParams) (*transactions.TransactionHistory, error) {
	var (
		res *transactions.TransactionHistory
		err error
	)

	// create transaction history with 'pending' status
	createTransactionArg := transactions.CreateTransactionParams{
		FromWalletID: arg.FromWalletID,
		ToWalletID:   arg.ToWalletID,
		Amount:       arg.Amount,
		TType:        arg.TType,
		TStatus:      transactions.TransactionStatusPending,
	}
	res, err = r.CreateTransaction(createTransactionArg)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction, err: %w", err)
	}

	// update wallet balance
	errTransaction := r.ExecDbTx(func(tr *transactionsRepository) error {
		updateWalletArg := wallets.UpdateWalletParams{
			Amount: arg.Amount,
			UserID: arg.UserID.Int32,
		}
		_, err = tr.walletsRepo.UpdateWalletByUserID(updateWalletArg)
		if err != nil {
			return fmt.Errorf("failed to update wallet, err: %w", err)
		}

		return err
	})

	// update transaction status
	var errUpdateTransaction error
	errUpdateStatus := r.ExecDbTx(func(tr *transactionsRepository) error {
		argUpdateStatus := transactions.UpdateTransactionStatusParams{
			Amount: arg.Amount,
			ID:     res.ID,
		}

		if errTransaction == nil {
			argUpdateStatus.TStatus = transactions.TransactionStatusCompleted
		} else {
			argUpdateStatus.TStatus = transactions.TransactionStatusFailed
		}

		res, errUpdateTransaction = tr.UpdateTransactionStatus(argUpdateStatus)
		if errUpdateTransaction != nil {
			return fmt.Errorf("failed to update transaction, err: %w", errUpdateTransaction)
		}

		return errUpdateTransaction
	})

	if errTransaction != nil {
		return res, errTransaction
	}

	return res, errUpdateStatus
}

func (r *transactionsRepository) TransactionTransfer(arg transactions.TransactionParams) (*transactions.TransactionHistory, error) {
	var (
		res *transactions.TransactionHistory
		err error
	)

	// create transaction history with 'pending' status
	createTransactionArg := transactions.CreateTransactionParams{
		FromWalletID: arg.FromWalletID,
		ToWalletID:   arg.ToWalletID,
		Amount:       arg.Amount,
		TType:        arg.TType,
		TStatus:      transactions.TransactionStatusPending,
	}
	res, err = r.CreateTransaction(createTransactionArg)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction, err: %w", err)
	}

	// update wallet balance
	errTransaction := r.ExecDbTx(func(tr *transactionsRepository) error {
		// update 'from_wallet' balance
		updateWalletArg := wallets.UpdateWalletParams{
			Amount:   -arg.Amount,
			WalletID: arg.FromWalletID.Int32,
		}
		_, err = tr.walletsRepo.UpdateWalletByID(updateWalletArg)
		if err != nil {
			return fmt.Errorf("failed to update 'from_wallet', err: %w", err)
		}

		// update 'to_wallet' balance
		updateWalletArg = wallets.UpdateWalletParams{
			Amount:   arg.Amount,
			WalletID: arg.ToWalletID.Int32,
		}
		_, err = tr.walletsRepo.UpdateWalletByID(updateWalletArg)
		if err != nil {
			return fmt.Errorf("failed to update 'to_wallet', err: %w", err)
		}

		return err
	})

	// update transaction status
	var errUpdateTransaction error
	errUpdateStatus := r.ExecDbTx(func(tr *transactionsRepository) error {
		argUpdateStatus := transactions.UpdateTransactionStatusParams{
			Amount: arg.Amount,
			ID:     res.ID,
		}

		if errTransaction == nil {
			argUpdateStatus.TStatus = transactions.TransactionStatusCompleted
		} else {
			argUpdateStatus.TStatus = transactions.TransactionStatusFailed
		}

		res, errUpdateTransaction = tr.UpdateTransactionStatus(argUpdateStatus)
		if errUpdateTransaction != nil {
			return fmt.Errorf("failed to update transaction, err: %w", errUpdateTransaction)
		}

		return errUpdateTransaction
	})

	if errTransaction != nil {
		return res, errTransaction
	}

	return res, errUpdateStatus
}
