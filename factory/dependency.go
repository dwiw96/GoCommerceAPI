package factory

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	authCache "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/cache"
	authHandler "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/handler"
	authRepository "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/repository"
	authService "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth/service"

	productsHandler "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products/handler"
	productsRepository "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products/repository"
	productsService "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products/service"
)

func InitFactory(router *gin.Engine, pool *pgxpool.Pool, rdClient *redis.Client, ctx context.Context) {
	iAuthRepo := authRepository.NewAuthRepository(pool, pool)
	iAuthCache := authCache.NewAuthCache(rdClient, ctx)
	iAuthService := authService.NewAuthService(iAuthRepo, iAuthCache, ctx)
	authHandler.NewAuthHandler(router, iAuthService, pool, rdClient, ctx)

	iProductRep := productsRepository.NewProductRepository(pool)
	iProductService := productsService.NewProductService(ctx, iProductRep)
	productsHandler.NewProductHandler(router, iProductService, pool, rdClient, ctx)
}
