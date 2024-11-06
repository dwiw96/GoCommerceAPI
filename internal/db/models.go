// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type TransactionStatus string

const (
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusFailed    TransactionStatus = "failed"
)

func (e *TransactionStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TransactionStatus(s)
	case string:
		*e = TransactionStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for TransactionStatus: %T", src)
	}
	return nil
}

type NullTransactionStatus struct {
	TransactionStatus TransactionStatus
	Valid             bool // Valid is true if TransactionStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTransactionStatus) Scan(value interface{}) error {
	if value == nil {
		ns.TransactionStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TransactionStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTransactionStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TransactionStatus), nil
}

type TransactionType string

const (
	TransactionTypePurchase   TransactionType = "purchase"
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
)

func (e *TransactionType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TransactionType(s)
	case string:
		*e = TransactionType(s)
	default:
		return fmt.Errorf("unsupported scan type for TransactionType: %T", src)
	}
	return nil
}

type NullTransactionType struct {
	TransactionType TransactionType
	Valid           bool // Valid is true if TransactionType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTransactionType) Scan(value interface{}) error {
	if value == nil {
		ns.TransactionType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TransactionType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTransactionType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TransactionType), nil
}

type Product struct {
	ID           int32
	Name         string
	Description  pgtype.Text
	Price        int32
	Availability int32
}

type RefreshTokenWhitelist struct {
	ID           int32
	UserID       int32
	RefreshToken pgtype.UUID
	ExpiresAt    pgtype.Timestamp
	CreatedAt    pgtype.Timestamp
}

type SecM struct {
	PrivateKey []byte
}

type TransactionsHistory struct {
	ID           int32
	FromWalletID int32
	ToWalletID   pgtype.Int4
	ProductID    int32
	Amount       int32
	TType        TransactionType
	TStatus      TransactionStatus
	CreatedAt    pgtype.Timestamp
}

type User struct {
	ID             int32
	Username       string
	Email          string
	HashedPassword string
	IsVerified     bool
	CreatedAt      pgtype.Timestamp
}

type Wallet struct {
	ID        int32
	UserID    int32
	Balance   int32
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}
