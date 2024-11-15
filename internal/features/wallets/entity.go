package wallets

import (
	"time"
)

type Wallet struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Balance   int32     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateWalletParams struct {
	UserID  int32
	Balance int32
}

type UpdateWalletParams struct {
	Amount   int32
	UserID   int32
	WalletID int32
}

type IRepository interface {
	CreateWallet(arg CreateWalletParams) (*Wallet, error)
	GetWalletByUserID(UserID int32) (*Wallet, error)
	UpdateWalletByUserID(arg UpdateWalletParams) (*Wallet, error)
	GetWalletByID(walletID int32) (*Wallet, error)
	UpdateWalletByID(arg UpdateWalletParams) (*Wallet, error)
}

type IService interface {
	CreateWallet(arg CreateWalletParams) (res *Wallet, code int, err error)
	GetWalletByUserID(UserID int32) (res *Wallet, code int, err error)
	DepositToWallet(arg UpdateWalletParams) (res *Wallet, code int, err error)
	WithdrawFromWallet(arg UpdateWalletParams) (res *Wallet, code int, err error)
}
