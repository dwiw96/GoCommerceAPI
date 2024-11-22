package repository

import (
	"context"
	"fmt"

	db "github.com/dwiw96/GoCommerceAPI/internal/db"
	product "github.com/dwiw96/GoCommerceAPI/internal/features/products"
)

type productRepository struct {
	db db.DBTX
}

func NewProductRepository(db db.DBTX) product.IRepository {
	return &productRepository{
		db: db,
	}
}

const createProduct = `-- name: CreateProduct :one
INSERT INTO products(
    name, 
    description, 
    price, 
    availability
) VALUES (
    $1, $2, $3, $4
) RETURNING id, name, description, price, availability
`

func (q *productRepository) CreateProduct(ctx context.Context, arg product.CreateProductParams) (*product.Product, error) {
	row := q.db.QueryRow(ctx, createProduct,
		arg.Name,
		arg.Description,
		arg.Price,
		arg.Availability,
	)
	var i product.Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Price,
		&i.Availability,
	)
	return &i, err
}

const getProductByID = `-- name: GetProductByID :one
SELECT id, name, description, price, availability FROM products
WHERE id = $1 LIMIT 1
`

func (q *productRepository) GetProductByID(ctx context.Context, id int32) (*product.Product, error) {
	row := q.db.QueryRow(ctx, getProductByID, id)
	var i product.Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Price,
		&i.Availability,
	)
	return &i, err
}

const listProducts = `-- name: ListProducts :many
SELECT id, name, description, price, availability FROM products
ORDER BY id ASC LIMIT $1 OFFSET $2
`

type ListProductsParams struct {
	Limit  int32
	Offset int32
}

func (q *productRepository) ListProducts(ctx context.Context, arg product.ListProductsParams) (*[]product.Product, error) {
	rows, err := q.db.Query(ctx, listProducts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []product.Product
	for rows.Next() {
		var i product.Product
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Price,
			&i.Availability,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &items, nil
}

const getTotalProduct = `-- name: GetTotalProduct :one
SELECT
	COUNT(*)
FROM products;
`

func (q *productRepository) GetTotalProducts(ctx context.Context) (int, error) {
	row := q.db.QueryRow(ctx, getTotalProduct)

	var res int
	err := row.Scan(
		&res,
	)

	return res, err
}

const updateProduct = `-- name: UpdateProduct :one
UPDATE 
    products
SET 
    name = coalesce($1, name),
    description = coalesce($2, description),
    price = coalesce($3, price),
    availability = coalesce($4, availability)
WHERE 
    id = $5
AND (
    $1::VARCHAR IS NOT NULL AND $1 IS DISTINCT FROM name OR
    $2::TEXT IS NOT NULL AND $2 IS DISTINCT FROM description OR
    $3::INT IS NOT NULL AND $3 IS DISTINCT FROM price OR
    $4::INT IS NOT NULL AND $4 IS DISTINCT FROM availability
)  RETURNING id, name, description, price, availability
`

func (q *productRepository) UpdateProduct(ctx context.Context, arg product.UpdateProductParams) (*product.Product, error) {
	row := q.db.QueryRow(ctx, updateProduct,
		arg.Name,
		arg.Description,
		arg.Price,
		arg.Availability,
		arg.ID,
	)
	var i product.Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Price,
		&i.Availability,
	)
	return &i, err
}

const deleteProduct = `-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1
`

func (q *productRepository) DeleteProduct(ctx context.Context, id int32) error {
	res, err := q.db.Exec(ctx, deleteProduct, id)

	if res.RowsAffected() == 0 {
		return fmt.Errorf("no row deleted")
	}
	return err
}

const updateProductAvailability = `-- name: UpdateProductAvailability :one
UPDATE 
    products
SET 
    availability = coalesce(availability + ($1), availability)
WHERE 
    id = $2
-- AND 
    -- $1::INT IS NOT NULL AND (availability + $1) >= 0
RETURNING id, name, description, price, availability
`

func (q *productRepository) UpdateProductAvailability(ctx context.Context, arg product.UpdateProductAvailabilityParams) (*product.Product, error) {
	row := q.db.QueryRow(ctx, updateProductAvailability,
		arg.Availability,
		arg.ID,
	)
	var i product.Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Price,
		&i.Availability,
	)
	return &i, err
}
