package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	product "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products"
	converter "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/converter"
	errorHandler "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/responses"
)

type productService struct {
	ctx  context.Context
	repo product.IRepository
}

func NewProductService(ctx context.Context, repo product.IRepository) product.IService {
	return &productService{
		ctx:  ctx,
		repo: repo,
	}
}

func (s *productService) CreateProduct(params product.CreateProductParams) (res *product.Product, code int, err error) {
	res, err = s.repo.CreateProduct(s.ctx, params)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errorHandler.CodeFailedDuplicated, errorHandler.ErrDuplicate
		}
		if strings.Contains(err.Error(), "violates") {
			return nil, errorHandler.CodeFailedUser, errorHandler.ErrViolation
		}
		return nil, errorHandler.CodeFailedServer, fmt.Errorf("failed to create product, err: %v", err)
	}

	return res, errorHandler.CodeSuccessCreate, err
}

func (s *productService) GetProductByID(id string) (res *product.Product, code int, err error) {
	newId, err := converter.ConvertStrToInt(id)
	if err != nil {
		return nil, errorHandler.CodeFailedServer, err
	}
	res, err = s.repo.GetProductByID(s.ctx, int32(newId))

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errorHandler.CodeSuccess, errorHandler.ErrNoData
		}
		return nil, errorHandler.CodeFailedServer, fmt.Errorf("failed to get product, err: %v", err)
	}

	return res, errorHandler.CodeSuccess, nil
}

func (s *productService) ListProducts(pageInput, limitInput string) (res *[]product.Product, currentPage, totalPages int, code int, err error) {
	pageConverted, err := converter.ConvertStrToInt(pageInput)
	if err != nil {
		return nil, 0, 0, errorHandler.CodeFailedServer, err
	}
	limitConverted, err := converter.ConvertStrToInt(limitInput)
	if err != nil {
		return nil, 0, 0, errorHandler.CodeFailedServer, err
	}

	page := int32(pageConverted)
	limit := int32(limitConverted)

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	arg := product.ListProductsParams{
		Limit:  limit,
		Offset: offset,
	}

	totalData, err := s.repo.GetTotalProducts(s.ctx)
	if err != nil {
		return nil, 0, 0, errorHandler.CodeFailedServer, fmt.Errorf("failed to get total data of products, err: %v", err)
	}
	totalPages = int(math.Ceil(float64(totalData) / float64(limit)))

	res, err = s.repo.ListProducts(s.ctx, arg)
	if err != nil {
		return nil, 0, 0, errorHandler.CodeFailedServer, fmt.Errorf("failed to list products, err: %v", err)
	}

	if *res == nil {
		return res, int(page), totalPages, errorHandler.CodeSuccess, nil
	}

	if len(*res) < int(limit) {
		return res, int(page), totalPages, errorHandler.CodeSuccess, nil
	}

	return res, int(page), totalPages, errorHandler.CodeSuccess, nil
}
func (s *productService) UpdateProduct(arg product.UpdateProductParams) (res *product.Product, code int, err error) {
	res, err = s.repo.UpdateProduct(s.ctx, arg)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errorHandler.CodeFailedDuplicated, errorHandler.ErrDuplicate
		}
		if strings.Contains(err.Error(), "violates") {
			return nil, errorHandler.CodeFailedUser, errorHandler.ErrViolation
		}
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errorHandler.CodeFailedUser, errorHandler.ErrNoData
		}
		return nil, errorHandler.CodeFailedServer, fmt.Errorf("failed to update product, err: %v", err)
	}

	return res, errorHandler.CodeSuccess, err
}
func (s *productService) DeleteProduct(idInput string) error {
	id, err := converter.ConvertStrToInt(idInput)
	if err != nil {
		return err
	}
	err = s.repo.DeleteProduct(s.ctx, int32(id))

	if err != nil {
		return fmt.Errorf("failed to delete product, err: %v", err)
	}

	return nil
}
