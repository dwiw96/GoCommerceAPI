package transactions

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type TransactionStatus string

const (
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type NullTransactionStatus struct {
	TransactionStatus TransactionStatus
	Valid             bool // Valid is true if TransactionStatus is not NULL
}

type TransactionTypes string

const (
	TransactionTypesPurchase   TransactionTypes = "purchase"
	TransactionTypesDeposit    TransactionTypes = "deposit"
	TransactionTypesWithdrawal TransactionTypes = "withdrawal"
	TransactionTypesTransfer   TransactionTypes = "transfer"
)

type NullTransactionTypes struct {
	TransactionTypes TransactionTypes
	Valid            bool // Valid is true if TransactionTypes is not NULL
}

type TransactionHistory struct {
	ID           int32
	FromWalletID pgtype.Int4
	ToWalletID   pgtype.Int4
	ProductID    pgtype.Int4
	Amount       int32
	Quantity     pgtype.Int4
	TType        TransactionTypes
	TStatus      TransactionStatus
	CreatedAt    pgtype.Timestamp
}

type CreateTransactionParams struct {
	FromWalletID pgtype.Int4
	ToWalletID   pgtype.Int4
	ProductID    pgtype.Int4
	Amount       int32
	Quantity     pgtype.Int4
	TType        TransactionTypes
	TStatus      TransactionStatus
}

type UpdateTransactionStatusParams struct {
	Amount  int32
	TStatus TransactionStatus
	ID      int32
}

type TransactionParams struct {
	UserID       pgtype.Int4
	FromWalletID pgtype.Int4
	ToWalletID   pgtype.Int4
	ProductID    pgtype.Int4
	Amount       int32
	Quantity     pgtype.Int4
	TType        TransactionTypes
}

type IRepository interface {
	CreateTransaction(arg CreateTransactionParams) (*TransactionHistory, error)
	UpdateTransactionStatus(arg UpdateTransactionStatusParams) (*TransactionHistory, error)
	TransactionPurchaseProduct(arg TransactionParams) (*TransactionHistory, error)
}

type IService interface {
	PurchaseProduct(arg TransactionParams) (res *TransactionHistory, code int, err error)
}
