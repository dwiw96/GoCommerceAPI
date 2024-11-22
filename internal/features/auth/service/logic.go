package service

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	auth "github.com/dwiw96/GoCommerceAPI/internal/features/auth"
	middleware "github.com/dwiw96/GoCommerceAPI/pkg/middleware"
	password "github.com/dwiw96/GoCommerceAPI/pkg/utils/password"
	errorHandler "github.com/dwiw96/GoCommerceAPI/pkg/utils/responses"
)

type authService struct {
	repo  auth.IRepository
	cache auth.ICache
	ctx   context.Context
}

func NewAuthService(repo auth.IRepository, cache auth.ICache, ctx context.Context) auth.IService {
	return &authService{
		repo:  repo,
		cache: cache,
		ctx:   ctx,
	}
}

func (s *authService) SignUp(input auth.SignupRequest) (user *auth.User, token string, code int, err error) {
	// check if the email have registered
	resGetUser, err := s.repo.GetUserByEmail(s.ctx, input.Email)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, "", errorHandler.CodeFailedServer, err
	}
	if resGetUser.Email == input.Email {
		return nil, "", errorHandler.CodeFailedDuplicated, fmt.Errorf("this email address is already in use")
	}

	// create arg for create user repository
	arg := auth.CreateUserParams{
		Username:       input.Username,
		Email:          input.Email,
		HashedPassword: input.Password,
	}

	arg.HashedPassword, err = password.HashingPassword(input.Password)
	if err != nil {
		return nil, "", errorHandler.CodeFailedServer, err
	}

	user, err = s.repo.CreateUser(s.ctx, arg)
	if err != nil {
		return nil, "", errorHandler.CodeFailedUser, err
	}

	// load private key
	key, err := s.repo.LoadKey(s.ctx)
	if err != nil {
		return nil, "", errorHandler.CodeFailedServer, fmt.Errorf("load key error: %w", err)
	}

	token, err = middleware.CreateToken(*user, 10, key)
	if err != nil {
		return nil, "", errorHandler.CodeFailedServer, err
	}

	return user, token, errorHandler.CodeSuccess, nil
}

func (s *authService) LogIn(input auth.LoginRequest) (user *auth.User, accessToken, refreshToken string, code int, err error) {
	user, err = s.repo.GetUserByEmail(s.ctx, input.Email)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			errMsg := fmt.Errorf("no user found with this email %s", input.Email)
			return nil, "", "", errorHandler.CodeFailedUnauthorized, errMsg
		}
		return nil, "", "", errorHandler.CodeFailedServer, err
	}

	err = password.VerifyHashPassword(input.Password, user.HashedPassword)
	if err != nil {
		errMsg := errors.New("password is wrong")
		return nil, "", "", errorHandler.CodeFailedUnauthorized, errMsg
	}

	key, err := s.repo.LoadKey(s.ctx)
	if err != nil {
		return nil, "", "", errorHandler.CodeFailedServer, fmt.Errorf("load key error: %w", err)
	}

	accessToken, err = middleware.CreateToken(*user, 60, key)
	if err != nil {
		errMsg := errors.New("failed generate access token")
		return nil, "", "", errorHandler.CodeFailedServer, errMsg
	}
	refreshTokenUUID, err := uuid.NewRandom()
	if err != nil {
		errMsg := errors.New("failed generate refresh token")
		return nil, "", "", errorHandler.CodeFailedServer, errMsg
	}

	err = s.repo.InsertRefreshToken(s.ctx, user.ID, refreshTokenUUID)
	if err != nil {
		return nil, "", "", errorHandler.CodeFailedServer, err
	}

	refreshToken = refreshTokenUUID.String()

	return user, accessToken, refreshToken, errorHandler.CodeSuccess, nil
}

func (s *authService) LogOut(payload auth.JwtPayload) error {
	err := s.repo.DeleteRefreshToken(s.ctx, payload.UserID)
	if err != nil {
		return err
	}

	err = s.cache.CachingBlockedToken(payload)

	return err
}

func (s *authService) DeleteUser(arg auth.DeleteUserParams) (code int, err error) {
	err = s.repo.DeleteAllUserInformation(s.ctx, arg)
	if err != nil {
		return errorHandler.CodeFailedUser, err
	}

	return errorHandler.CodeSuccess, nil
}

func (s *authService) RefreshToken(refreshToken, accessToken string) (newRefreshToken, newAccessToken string, code int, err error) {
	code = errorHandler.CodeSuccess
	key, err := s.repo.LoadKey(s.ctx)
	if err != nil {
		return "", "", errorHandler.CodeFailedServer, err
	}

	authHeader := "Bearer " + accessToken
	payload, err := middleware.ReadToken(authHeader, key)
	if err != nil {
		return "", "", errorHandler.CodeFailedServer, err
	}

	err = s.cache.CachingBlockedToken(*payload)
	if err != nil {
		return "", "", errorHandler.CodeFailedServer, fmt.Errorf("failed to caching access token, msg: %v", err)
	}

	refreshTokenUUID, err := uuid.Parse(refreshToken)
	if err != nil {
		return "", "", errorHandler.CodeFailedServer, fmt.Errorf("failed to convert refresh token from string to uuid, msg: %v", err)
	}

	// Read and validate refresh token from database
	res, errReadRefreshToken := s.repo.ReadRefreshToken(s.ctx, payload.UserID, refreshTokenUUID)
	err = s.validateRefreshToken(res, errReadRefreshToken)
	if err != nil {
		return "", "", errorHandler.CodeFailedUnauthorized, err
	}

	newAccessToken, newRefreshToken, err = s.createNewToken(key, payload)
	if err != nil {
		return "", "", errorHandler.CodeFailedServer, err
	}

	return
}

// ValidateRefreshToken return error.
//
// ValidateRefreshToken check the refresh token from database, what to check:
//   - check error when read from database.
//   - check if the refresh token is nil
//   - check if refresh token is expired
//
// When the refresh token is expired:
//
//	delete refresh token from database
func (s *authService) validateRefreshToken(arg *auth.RefreshTokenWhitelist, errIn error) (err error) {
	if errIn != nil {
		return fmt.Errorf("invalid refresh token, msg: %v", errIn)
	}
	if arg.RefreshToken == uuid.Nil {
		return fmt.Errorf("invalid refresh token")
	}

	if time.Now().UTC().After(arg.ExpiresAt) {
		err = s.repo.DeleteRefreshToken(s.ctx, arg.UserID)
		if err != nil {
			return fmt.Errorf("failed to process expired refresh token, msg: %v", err)
		}
		return fmt.Errorf("refresh token is expire")
	}

	return nil
}

// createNewToken return new access token, new refresh token and error
func (s *authService) createNewToken(key *rsa.PrivateKey, payload *auth.JwtPayload) (newAccessToken, newRefreshToken string, err error) {
	user := auth.User{
		ID:       payload.UserID,
		Username: payload.Name,
		Email:    payload.Email,
	}

	newAccessToken, err = middleware.CreateToken(user, 60, key)
	if err != nil {
		return
	}
	newRefreshTokenUUID, err := uuid.NewRandom()
	if err != nil {
		return
	}
	err = s.repo.UpdateRefreshToken(s.ctx, payload.UserID, newRefreshTokenUUID)
	if err != nil {
		return
	}

	newRefreshToken = newRefreshTokenUUID.String()

	return
}
