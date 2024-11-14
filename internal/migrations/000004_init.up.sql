BEGIN;
CREATE TYPE transaction_types AS ENUM ('purchase', 'deposit', 'withdrawal', 'transfer');

CREATE TYPE transaction_status AS ENUM ('completed', 'pending', 'failed');

CREATE TABLE transaction_histories(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_transaction_histories_id PRIMARY KEY,
    from_wallet_id INT NULL,
        CONSTRAINT fk_transaction_histories_from_wallet_id FOREIGN KEY (from_wallet_id)
            REFERENCES wallets(id),
    to_wallet_id INT NULL,
        CONSTRAINT fk_transaction_histories_to_wallet_id FOREIGN KEY (to_wallet_id)
            REFERENCES wallets(id),
    product_id INT NULL,
        CONSTRAINT fk_transaction_histories_product_id FOREIGN KEY (product_id)
            REFERENCES products(id),
    amount INT NOT NULL DEFAULT 0,
    quantity INT NULL DEFAULT 0
        CONSTRAINT ck_transaction_histories_quantity CHECK (quantity >= 0),
    t_type transaction_types NOT NULL,
    t_status transaction_status NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_transaction_histories_from_wallet_id ON transaction_histories(from_wallet_id);
CREATE INDEX ix_transaction_histories_to_wallet_id ON transaction_histories(to_wallet_id);
CREATE INDEX ix_transaction_histories_product_id ON transaction_histories(product_id);
CREATE INDEX ix_transaction_histories_t_type ON transaction_histories(t_type);
CREATE INDEX ix_transaction_histories_t_status ON transaction_histories(t_status);
CREATE INDEX ix_transaction_histories_created_at ON transaction_histories(created_at);
COMMIT;
