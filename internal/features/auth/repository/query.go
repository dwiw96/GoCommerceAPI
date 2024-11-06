package repository

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	db "github.com/dwiw96/vocagame-technical-test-backend/internal/db"
	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authRepository struct {
	db   db.DBTX
	txDb *pgxpool.Pool
}

func NewAuthRepository(db db.DBTX, txDb *pgxpool.Pool) auth.IRepository {
	return &authRepository{
		db:   db,
		txDb: txDb,
	}
}

const createUser = `-- name: CreateUser :one
INSERT INTO users(
    username,
    email,
    hashed_password
) VALUES (
    $1, $2, $3
) RETURNING id, username, email, hashed_password, is_verified, created_at
`

func (r *authRepository) CreateUser(ctx context.Context, arg auth.CreateUserParams) (*auth.User, error) {
	row := r.db.QueryRow(ctx, createUser, arg.Username, arg.Email, arg.HashedPassword)
	var i auth.User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.HashedPassword,
		&i.IsVerified,
		&i.CreatedAt,
	)
	return &i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, hashed_password, is_verified, created_at FROM users WHERE email = $1
`

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	row := r.db.QueryRow(ctx, getUserByEmail, email)
	var i auth.User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.HashedPassword,
		&i.IsVerified,
		&i.CreatedAt,
	)
	return &i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE
    users
SET
    username = coalesce($1, username),
    hashed_password = coalesce($2, hashed_password)
WHERE
    id = $3
AND (
    $1::VARCHAR IS NOT NULL AND $1 IS DISTINCT FROM username OR
    $2::VARCHAR IS NOT NULL AND $2 IS DISTINCT FROM hashed_password
) RETURNING id, username, email, hashed_password, is_verified, created_at
`

func (r *authRepository) UpdateUser(ctx context.Context, arg auth.UpdateUserParams) (*auth.User, error) {
	row := r.db.QueryRow(ctx, updateUser, arg.Username, arg.HashedPassword, arg.ID)
	var i auth.User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.HashedPassword,
		&i.IsVerified,
		&i.CreatedAt,
	)
	return &i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1 AND email = $2
`

func (r *authRepository) DeleteUser(ctx context.Context, arg auth.DeleteUserParams) error {
	res, err := r.db.Exec(ctx, deleteUser, arg.ID, arg.Email)

	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete user, err: no user found")
	}
	return err
}

func (r *authRepository) LoadKey(ctx context.Context) (key *rsa.PrivateKey, err error) {
	query := "select private_key from sec_m"
	var keyBytes []byte
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		errMsg := fmt.Errorf("failed to load private key, err: %v", err)
		return nil, errMsg
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&keyBytes)
		if err != nil {
			errMsg := fmt.Errorf("failed to scan private key, err: %v", err)
			return nil, errMsg
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
		if err != nil {
			errMsg := fmt.Errorf("failed to parse private key, err: %v", err)
			return nil, errMsg
		}

		return privateKey, nil
	}

	return nil, errors.New("no private key found in database")
}

func (r *authRepository) ReadRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) (res *auth.RefreshTokenWhitelist, err error) {
	query := "SELECT * FROM refresh_token_whitelist WHERE user_id = $1 AND refresh_token = $2;"

	var result auth.RefreshTokenWhitelist
	err = r.db.QueryRow(ctx, query, userID, refreshToken).Scan(&result.ID, &result.UserID, &result.RefreshToken, &result.ExpiresAt, &result.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh token, err: %v", err)
	}

	return &result, nil
}

func (r *authRepository) InsertRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) (err error) {
	query := "INSERT INTO refresh_token_whitelist(user_id, refresh_token, expires_at) VALUES($1, $2, NOW() + INTERVAL '5 minute')"

	res, err := r.db.Exec(ctx, query, userID, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token, err: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("there are no rows affrected when insert refresh token")
	}

	return nil
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, userID int32) (err error) {
	query := "DELETE FROM refresh_token_whitelist WHERE user_id = $1;"

	res, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token, err: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("there are no rows affected when delete refresh token")
	}

	return nil
}

func (r *authRepository) ExecDbTx(ctx context.Context, fn func(*authRepository) error) error {
	tx, err := r.txDb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start db transaction, err: %v", err)
	}

	err = fn(r)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback, err = %v", rbErr)
		}
		return fmt.Errorf("failed db transaction, err: %v", err)
	}

	return tx.Commit(ctx)
}

func (r *authRepository) DeleteAllUserInformation(ctx context.Context, arg auth.DeleteUserParams) (err error) {
	r.ExecDbTx(ctx, func(ar *authRepository) error {
		err = ar.DeleteUser(ctx, arg)
		if err != nil {
			return err
		}

		err = ar.DeleteRefreshToken(ctx, arg.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}

func (r *authRepository) UpdateRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) (err error) {
	r.ExecDbTx(ctx, func(ar *authRepository) error {
		err = ar.DeleteRefreshToken(ctx, userID)
		if err != nil {
			return err
		}

		err = ar.InsertRefreshToken(ctx, userID, refreshToken)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}
