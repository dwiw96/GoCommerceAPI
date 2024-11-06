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
    user_id INT NOT NULL,
        CONSTRAINT fk_wallets_user_id FOREIGN KEY (user_id)
            REFERENCES users(id),
    balance INT NOT NULL DEFAULT 0
        CONSTRAINT ck_wallets_balance CHECK (balance >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TYPE transaction_type AS ENUM ('purchase', 'deposit', 'withdrawal');
CREATE TYPE transaction_status AS ENUM ('completed', 'pending', 'failed');

CREATE TABLE transactions_history(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_transactions_history_id PRIMARY KEY,
    from_wallet_id INT NOT NULL,
        CONSTRAINT fk_transactions_history_from_wallet_id FOREIGN KEY (from_wallet_id)
            REFERENCES wallets(id),
    to_wallet_id INT NULL,
        CONSTRAINT fk_transactions_history_to_wallet_id FOREIGN KEY (to_wallet_id)
            REFERENCES wallets(id),
    product_id INT NOT NULL,
        CONSTRAINT fk_transactions_history_product_id FOREIGN KEY (product_id)
            REFERENCES products(id),
    amount INT NOT NULL DEFAULT 0
        CONSTRAINT ck_transactions_history_amount CHECK (amount >= 0),
    t_type transaction_type NOT NULL,
    t_status transaction_status NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_transactions_history_from_wallet_id ON transactions_history(from_wallet_id);
CREATE INDEX ix_transactions_history_to_wallet_id ON transactions_history(to_wallet_id);
CREATE INDEX ix_transactions_history_product_id ON transactions_history(product_id);
CREATE INDEX ix_transactions_history_t_type ON transactions_history(t_type);
CREATE INDEX ix_transactions_history_t_status ON transactions_history(t_status);
CREATE INDEX ix_transactions_history_created_at ON transactions_history(created_at);
COMMIT;
