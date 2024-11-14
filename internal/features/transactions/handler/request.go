package handler

import (
	"github.com/dwiw96/vocagame-technical-test-backend/internal/features/transactions"
	"github.com/jackc/pgx/v5/pgtype"
)

type transactionReq struct {
	TransactionType  string `json:"transaction_type" validate:"required,oneof=purchase transfer deposit withdraw"`
	FromWalletUserID int32  `json:"from_wallet_id" validate:"number"`
	ToWalletUserID   int32  `json:"to_wallet_id" validate:"number"`
	ProductID        int32  `json:"product_id"`
	Amount           int32  `json:"amount"`
	Quantity         int32  `json:"quantity" validate:"number"`
}

func toPurchaseProductArg(userID int32, input transactionReq) transactions.TransactionParams {
	return transactions.TransactionParams{
		UserID:       pgtype.Int4{Int32: userID, Valid: true},
		FromWalletID: pgtype.Int4{Int32: input.FromWalletUserID, Valid: true},
		ToWalletID:   pgtype.Int4{Int32: input.ToWalletUserID, Valid: true},
		Amount:       input.Amount,
		ProductID:    pgtype.Int4{Int32: input.ProductID, Valid: true},
		Quantity:     pgtype.Int4{Int32: input.Quantity, Valid: true},
	}
}
