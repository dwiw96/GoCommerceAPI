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
    hashed_password = coalesce($2, hashed_password)
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
ORDER BY id ASC LIMIT $1 OFFSET $2;

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

-- name: UpdateProductAvailability :one
UPDATE 
    products
SET 
    availability = coalesce(availability + ($1), availability)
WHERE 
    id = $2
AND 
    $1::INT IS NOT NULL AND (availability + $1) >= 0
RETURNING id, name, description, price, availability;

-- name: CreateWallet :one
INSERT INTO wallets(
    user_id,
    balance
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetWalletByUserID :one
SELECT * FROM wallets WHERE user_id = $1;

-- name: ListWallets :many
SELECT * FROM wallets 
ORDER BY id ASC LIMIT $1 OFFSET $2;

-- name: UpdateWallet :one
UPDATE
    wallets
SET
    balance = balance + $1, 
    updated_at = NOW()
WHERE 
    user_id = $2
RETURNING *;

-- name: DeleteWallet :exec
DELETE FROM wallets WHERE id = $1;


-- name: CreateTransaction :one
INSERT INTO 
    transaction_histories(
        from_wallet_id,
        to_wallet_id,
        product_id,
        amount,
        quantity,
        t_type,
        t_status
    )
VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateTransactionStatus :one
UPDATE
    transaction_histories
SET
    t_status = $1
WHERE 
    id = $2
RETURNING *;

-- name: TransactionToWallet :exec
BEGIN;
UPDATE 
    wallets
SET 
    balance = balance + $1,
    updated_at = NOW()
WHERE
    user_id = $2;

UPDATE 
    wallets
SET 
    balance = balance - $1,
    updated_at = NOW()
WHERE
    user_id = $3;

INSERT INTO  transaction_histories(
    from_user_id,
    to_wallet_id,
    product_id,
    amount,
    quantity
) VALUES (
    $2, $3, $4, $1, $5
);
COMMIT;

-- name: PurchaseProduct :exec
BEGIN;
UPDATE 
    wallets
SET 
    balance = balance - $1,
    updated_at = NOW()
WHERE
    user_id = $2;

UPDATE 
    products
SET 
    availability = availability - $3
WHERE 
    id = $4;

INSERT INTO 
    transaction_histories(
        from_user_id,
        to_wallet_id,
        product_id,
        amount,
        quantity
    )
VALUES (
    $2, $3, $4, $1, $5
);
COMMIT;
