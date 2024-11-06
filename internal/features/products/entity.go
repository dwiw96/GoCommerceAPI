package products

import (
	"context"
)

type Product struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int32  `json:"price"`
	Availability int32  `json:"availability"`
}

type CreateProductParams struct {
	Name         string
	Description  string
	Price        int32
	Availability int32
}

type UpdateProductParams struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int32  `json:"price"`
	Availability int32  `json:"availability"`
}

type ListProductsParams struct {
	Limit  int32
	Offset int32
}

type IService interface {
	CreateProduct(params CreateProductParams) (res *Product, code int, err error)
	GetProductByID(id string) (res *Product, code int, err error)
	ListProducts(page, limit string) (res *[]Product, currentPage, totalPages int, code int, err error)
	UpdateProduct(arg UpdateProductParams) (res *Product, code int, err error)
	DeleteProduct(id string) error
}

type IRepository interface {
	CreateProduct(ctx context.Context, arg CreateProductParams) (*Product, error)
	GetProductByID(ctx context.Context, id int32) (*Product, error)
	ListProducts(ctx context.Context, arg ListProductsParams) (*[]Product, error)
	GetTotalProducts(ctx context.Context) (int, error)
	UpdateProduct(ctx context.Context, arg UpdateProductParams) (*Product, error)
	DeleteProduct(ctx context.Context, id int32) error
}
