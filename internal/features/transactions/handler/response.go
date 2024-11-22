package handler

import (
	"time"

	transactions "github.com/dwiw96/GoCommerceAPI/internal/features/transactions"
)

type transactionResp struct {
	FromWalletUserID int32                          `json:"from_wallet_id" validate:"number"`
	ToWalletUserID   int32                          `json:"to_wallet_id" validate:"number"`
	ProductID        int32                          `json:"product_id"`
	Quantity         int32                          `json:"quantity" validate:"number"`
	Amount           int32                          `json:"amount"`
	TType            transactions.TransactionTypes  `json:"transaction_type"`
	TStatus          transactions.TransactionStatus `json:"transaction_status"`
	CreatedAt        time.Time                      `json:"created_at"`
}

func toTransactionResp(input *transactions.TransactionHistory) transactionResp {
	return transactionResp{
		FromWalletUserID: input.FromWalletID.Int32,
		ToWalletUserID:   input.ToWalletID.Int32,
		ProductID:        input.ProductID.Int32,
		Quantity:         input.Quantity.Int32,
		Amount:           input.Amount,
		TType:            input.TType,
		TStatus:          input.TStatus,
		CreatedAt:        input.CreatedAt.Time,
	}
}
