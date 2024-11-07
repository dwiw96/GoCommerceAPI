package repository

import (
	"context"

	db "github.com/dwiw96/vocagame-technical-test-backend/internal/db"
	wallets "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets"
)

type walletsRepository struct {
	db  db.DBTX
	ctx context.Context
}

func NewWalletsRepository(db db.DBTX, ctx context.Context) wallets.IRepository {
	return &walletsRepository{
		db:  db,
		ctx: ctx,
	}
}

const createWallet = `-- name: CreateWallet :one
INSERT INTO wallets(
    user_id,
    balance
) VALUES (
    $1, $2
) RETURNING id, user_id, balance, created_at, updated_at
`

func (r *walletsRepository) CreateWallet(arg wallets.CreateWalletParams) (*wallets.Wallet, error) {
	row := r.db.QueryRow(r.ctx, createWallet, arg.UserID, arg.Balance)
	var i wallets.Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const getWalletByID = `-- name: GetWalletByUserID :one
SELECT id, user_id, balance, created_at, updated_at FROM wallets WHERE user_id = $1
`

func (r *walletsRepository) GetWalletByUserID(userID int32) (*wallets.Wallet, error) {
	row := r.db.QueryRow(r.ctx, getWalletByID, userID)
	var i wallets.Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const updateWallet = `-- name: UpdateWallet :one
UPDATE
    wallets
SET
    balance = balance + ($1), 
    updated_at = NOW()
WHERE user_id = $2
RETURNING id, user_id, balance, created_at, updated_at
`

func (r *walletsRepository) UpdateWallet(arg wallets.UpdateWalletParams) (*wallets.Wallet, error) {
	row := r.db.QueryRow(r.ctx, updateWallet, arg.Amount, arg.UserID)
	var i wallets.Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
