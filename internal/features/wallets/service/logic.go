package service

import (
	"context"
	"errors"
	"fmt"

	wallets "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets"
	errs "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/responses"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type walletsService struct {
	ctx  context.Context
	repo wallets.IRepository
}

func NewWalletsService(ctx context.Context, repo wallets.IRepository) wallets.IService {
	return &walletsService{
		ctx:  ctx,
		repo: repo,
	}
}

func handleError(err error) (code int, errRes error) {
	if err == pgx.ErrNoRows {
		return errs.CodeFailedUser, errs.ErrNoData
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.ConstraintName == "ck_wallets_balance" {
			return errs.CodeFailedUser, fmt.Errorf("balance minimum is 0")
		}
		switch pgErr.Code {
		case "23505": // UNIQUE violation
			return errs.CodeFailedDuplicated, errs.ErrDuplicate
		case "23514": // CHECK violation
			return errs.CodeFailedUser, errs.ErrCheckConstraint
		case "23502": // NOT NULL violation
			return errs.CodeFailedUser, errs.ErrNotNull
		case "23503": // Foreign Key violation
			return errs.CodeFailedUser, errs.ErrViolation
		default:
			errRes = fmt.Errorf("database error occurred")
		}
	}

	return errs.CodeFailedServer, errRes
}

func (s *walletsService) CreateWallet(arg wallets.CreateWalletParams) (res *wallets.Wallet, code int, err error) {
	res, err = s.repo.CreateWallet(arg)
	if err != nil {
		code, err = handleError(err)
		return nil, code, err
	}

	return res, errs.CodeSuccess, err
}

func (s *walletsService) GetWalletByUserID(userID int32) (res *wallets.Wallet, code int, err error) {
	res, err = s.repo.GetWalletByUserID(userID)
	if err != nil {
		code, err = handleError(err)
		return nil, code, err
	}
	return res, errs.CodeSuccess, nil
}

func (s *walletsService) DepositToWallet(arg wallets.UpdateWalletParams) (res *wallets.Wallet, code int, err error) {
	if arg.Amount <= 0 {
		return nil, errs.CodeFailedUser, errs.ErrInvalidInput
	}

	res, err = s.repo.UpdateWallet(arg)
	if err != nil {
		code, err = handleError(err)
		return nil, code, err
	}

	return res, errs.CodeSuccess, nil
}

func (s *walletsService) WithdrawFromWallet(arg wallets.UpdateWalletParams) (res *wallets.Wallet, code int, err error) {
	arg.Amount *= -1
	if arg.Amount >= 0 {
		return nil, errs.CodeFailedUser, errs.ErrInvalidInput
	}

	res, err = s.repo.UpdateWallet(arg)
	if err != nil {
		code, err = handleError(err)
		return nil, code, err
	}

	return res, errs.CodeSuccess, nil
}
