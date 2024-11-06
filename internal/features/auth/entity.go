package auth

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// database model for users table
type User struct {
	ID             int32
	Username       string
	Email          string
	HashedPassword string
	IsVerified     bool
	CreatedAt      time.Time
}

// params for repository method
type CreateUserParams struct {
	Username       string
	Email          string
	HashedPassword string
}

type UpdateUserParams struct {
	Username       string
	HashedPassword string
	ID             int32
}

type DeleteUserParams struct {
	ID    int32
	Email string
}

// params for service method
type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// param for jwt
type JwtPayload struct {
	jwt.RegisteredClaims
	ID      uuid.UUID `json:"id"`
	UserID  int32     `json:"user_id"`
	Iss     string    `json:"iss"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Address string    `json:"address,omitempty"`
	Iat     int64     `json:"iat"`
	Exp     int64     `json:"exp"`
}

// param for refresh token
type RefreshTokenWhitelist struct {
	ID           int32
	UserID       int32
	RefreshToken uuid.UUID
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type IRepository interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error)
	DeleteUser(ctx context.Context, arg DeleteUserParams) error

	LoadKey(ctx context.Context) (key *rsa.PrivateKey, err error)
	ReadRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) (res *RefreshTokenWhitelist, err error)
	InsertRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) (err error)
	DeleteRefreshToken(ctx context.Context, userID int32) (err error)
	DeleteAllUserInformation(ctx context.Context, arg DeleteUserParams) (err error)
	UpdateRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) (err error)
}

type IService interface {
	SignUp(input SignupRequest) (user *User, token string, code int, err error)
	LogIn(input LoginRequest) (user *User, accessToken, refreshToken string, code int, err error)
	LogOut(payload JwtPayload) error
	DeleteUser(arg DeleteUserParams) (code int, err error)
	RefreshToken(refreshToken, accessToken string) (newRefreshToken, newAccessToken string, code int, err error)
}

type ICache interface {
	CachingBlockedToken(payload JwtPayload) error
	CheckBlockedToken(payload JwtPayload) error
}
