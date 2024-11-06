-- name: CreateUser :one
INSERT INTO users(
    username,
    email,
    hashed_password
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserVerification :exec
UPDATE users SET is_verified = TRUE WHERE id = $1 AND email = $2;

-- name: UpdateUser :one
UPDATE
    users
SET
    username = coalesce($1, username),
    hashed_password = coalesce($2, hashed_password),
WHERE
    id = $3
AND (
    $1::VARCHAR IS NOT NULL AND $1 IS DISTINCT FROM username OR
    $2::VARCHAR IS NOT NULL AND $2 IS DISTINCT FROM hashed_password
) RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1 AND email = $2;

-- name: CreateProduct :one
INSERT INTO products(
    name, 
    description, 
    price, 
    availability
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
SELECT * FROM products
ORDER BY name ASC LIMIT $1 OFFSET $2;

-- name: UpdateProduct :one
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
)  RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: CreateWallet :one
INSERT INTO wallets(
    user_id,
    balance
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetWalletByID :one
SELECT * FROM wallets WHERE id = $1;

-- name: ListWallets :many
SELECT * FROM wallets 
ORDER BY id ASC LIMIT $1 OFFSET $2;

-- name: UpdateWallet :one
UPDATE
    wallets
SET
    balance = balance + (-$1), 
    updated = NOW()
WHERE id = $2
RETURNING *;

-- name: DeleteWallet :exec
DELETE FROM wallets WHERE id = $1;
