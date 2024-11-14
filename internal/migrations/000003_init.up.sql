BEGIN;
CREATE TABLE products(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_products_id PRIMARY KEY,
    name VARCHAR(255) NOT NULL
        CONSTRAINT uq_products_name UNIQUE,
        CONSTRAINT ck_products_name_length CHECK (LENGTH(TRIM(name)) > 0),
    description TEXT NULL DEFAULT ''
        CONSTRAINT ck_products_description_length CHECK (LENGTH(TRIM(description)) >= 0),
    price INT NOT NULL DEFAULT 0
        CONSTRAINT ck_products_price CHECK (price >= 0),
    availability INT NOT NULL DEFAULT 0
        CONSTRAINT ck_products_availability CHECK (availability >= 0)
);

CREATE INDEX ix_products_name ON products(name);
CREATE INDEX ix_products_price ON products(price);
CREATE INDEX ix_products_availability ON products(availability);

CREATE TABLE wallets(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_wallets_id PRIMARY KEY,
    user_id INT NOT NULL
        CONSTRAINT uq_wallets_user_id UNIQUE,
        CONSTRAINT fk_wallets_user_id FOREIGN KEY (user_id)
            REFERENCES users(id),
    balance INT NOT NULL DEFAULT 0
        CONSTRAINT ck_wallets_balance CHECK (balance >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
COMMIT;
