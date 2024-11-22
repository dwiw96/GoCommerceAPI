package factory

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	authCache "github.com/dwiw96/GoCommerceAPI/internal/features/auth/cache"
	authHandler "github.com/dwiw96/GoCommerceAPI/internal/features/auth/handler"
	authRepository "github.com/dwiw96/GoCommerceAPI/internal/features/auth/repository"
	authService "github.com/dwiw96/GoCommerceAPI/internal/features/auth/service"

	productsHandler "github.com/dwiw96/GoCommerceAPI/internal/features/products/handler"
	productsRepository "github.com/dwiw96/GoCommerceAPI/internal/features/products/repository"
	productsService "github.com/dwiw96/GoCommerceAPI/internal/features/products/service"

	walletsHandler "github.com/dwiw96/GoCommerceAPI/internal/features/wallets/handler"
	walletsRepository "github.com/dwiw96/GoCommerceAPI/internal/features/wallets/repository"
	walletsService "github.com/dwiw96/GoCommerceAPI/internal/features/wallets/service"

	transactionsHandler "github.com/dwiw96/GoCommerceAPI/internal/features/transactions/handler"
	transactionsRepository "github.com/dwiw96/GoCommerceAPI/internal/features/transactions/repository"
	transactionsService "github.com/dwiw96/GoCommerceAPI/internal/features/transactions/service"
)

func InitFactory(router *gin.Engine, pool *pgxpool.Pool, rdClient *redis.Client, ctx context.Context) {
	iAuthRepo := authRepository.NewAuthRepository(pool, pool)
	iAuthCache := authCache.NewAuthCache(rdClient, ctx)
	iAuthService := authService.NewAuthService(iAuthRepo, iAuthCache, ctx)
	authHandler.NewAuthHandler(router, iAuthService, pool, rdClient, ctx)

	iProductRep := productsRepository.NewProductRepository(pool)
	iProductService := productsService.NewProductService(ctx, iProductRep)
	productsHandler.NewProductHandler(router, iProductService, pool, rdClient, ctx)

	iWalletsRep := walletsRepository.NewWalletsRepository(pool, ctx)
	iWalletsService := walletsService.NewWalletsService(ctx, iWalletsRep)
	walletsHandler.NewWalletsHandler(router, iWalletsService, pool, rdClient, ctx)

	iTransactionsRep := transactionsRepository.NewTransactionsRepository(pool, pool, ctx)
	iTransactionsService := transactionsService.NewTransactionsService(ctx, iTransactionsRep)
	transactionsHandler.NewTransactionsHandler(router, iTransactionsService, pool, rdClient, ctx)
}
