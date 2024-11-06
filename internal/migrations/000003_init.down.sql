BEGIN;
DROP TABLE IF EXISTS transactions_history;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS transaction_status;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS products;
COMMIT;
