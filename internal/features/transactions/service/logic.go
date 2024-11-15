package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	transactions "github.com/dwiw96/vocagame-technical-test-backend/internal/features/transactions"
	errs "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/responses"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type transactionsService struct {
	ctx  context.Context
	repo transactions.IRepository
}

func NewTransactionsService(ctx context.Context, repo transactions.IRepository) transactions.IService {
	return &transactionsService{
		ctx:  ctx,
		repo: repo,
	}
}

func handleError(arg error) (code int, err error) {
	log.Println(arg)
	if errors.Is(arg, pgx.ErrNoRows) {
		return errs.CodeFailedUser, errs.ErrNoData
	}
	var pgErr *pgconn.PgError
	if errors.As(arg, &pgErr) {
		if pgErr.ConstraintName == "ck_transactions_balance" {
			return errs.CodeFailedUser, fmt.Errorf("balance minimum is 0")
		}
		switch pgErr.Code {
		case "23505": // UNIQUE violation
			return errs.CodeFailedDuplicated, errs.ErrDuplicate
		case "23514": // CHECK violation
			if pgErr.ConstraintName == "ck_wallets_balance" {
				return errs.CodeFailedUser, fmt.Errorf("balance is insufficient")
			}
			if pgErr.ConstraintName == "ck_products_availability" {
				return errs.CodeFailedUser, fmt.Errorf("product stock is insufficient")
			}
			return errs.CodeFailedUser, errs.ErrCheckConstraint
		case "23502": // NOT NULL violation
			return errs.CodeFailedUser, errs.ErrNotNull
		case "23503": // Foreign Key violation
			return errs.CodeFailedUser, errs.ErrViolation
		default:
			err = fmt.Errorf("database error occurred")
		}
	}

	return errs.CodeFailedServer, err
}

func (s *transactionsService) PurchaseProduct(arg transactions.TransactionParams) (res *transactions.TransactionHistory, code int, err error) {
	if arg.Quantity.Int32 <= int32(0) {
		return nil, 400, fmt.Errorf("quantity must be more than 0")
	}
	code = 200
	res, err = s.repo.TransactionPurchaseProduct(arg)
	if err != nil {
		code, newErr := handleError(err)
		log.Println("handlerError code:", code)
		log.Println("handlerError err:", newErr)
		return res, code, newErr
	}

	return
}

func (s *transactionsService) DepositOrWithdraw(arg transactions.TransactionParams) (res *transactions.TransactionHistory, code int, err error) {
	if arg.TType == transactions.TransactionTypesDeposit {
		if arg.Amount <= int32(0) {
			return nil, errs.CodeFailedUser, errs.ErrLessOrEqualToZero
		}
		arg.FromWalletID.Valid = false
	}
	if arg.TType == transactions.TransactionTypesWithdrawal {
		if arg.Amount >= int32(0) {
			return nil, errs.CodeFailedUser, errs.ErrGreaterOrEqualToZero
		}
		arg.ToWalletID.Valid = false
	}

	arg.ProductID.Valid = false
	arg.Quantity.Valid = false

	code = errs.CodeSuccess
	res, err = s.repo.TransactionDepositOrWithdraw(arg)
	if err != nil {
		code, err = handleError(err)
		return res, code, err
	}

	return
}
