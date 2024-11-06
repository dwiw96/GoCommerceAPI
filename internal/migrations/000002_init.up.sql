BEGIN;
CREATE TABLE users(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_users_id PRIMARY KEY,
    username VARCHAR(255) NOT NULL
        CONSTRAINT ck_users_username_length CHECK (LENGTH(TRIM(username)) > 0),
    email VARCHAR(255) NOT NULL
        CONSTRAINT uq_users_email UNIQUE,
        CONSTRAINT ck_users_email_length CHECK (LENGTH(TRIM(email)) >= 5),
    hashed_password TEXT NOT NULL
        CONSTRAINT ck_users_hashed_password_length CHECK (LENGTH(TRIM(hashed_password)) > 0),
        CONSTRAINT uq_users_hashed_password UNIQUE(hashed_password),
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_users_username ON users(username);
CREATE INDEX ix_users_email ON users(email);
CREATE INDEX ix_users_created_at ON users(created_at);

CREATE TABLE refresh_token_whitelist(
    id INT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT pk_refresh_token_whitelist_id PRIMARY KEY,
    user_id INT NOT NULL
        CONSTRAINT uq_refresh_token_whitelist_user_id UNIQUE,
        CONSTRAINT fk_refresh_token_whitelist_user_id FOREIGN KEY (user_id)
            REFERENCES users(id),
    refresh_token UUID NOT NULL
        CONSTRAINT uq_refresh_token_whitelist_refresh_token UNIQUE,
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '10 second',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_refresh_token_whitelist_user_id ON refresh_token_whitelist(id);
CREATE INDEX ix_refresh_token_whitelist_refresh_token ON refresh_token_whitelist(refresh_token);
CREATE INDEX ix_refresh_token_whitelist_created_at ON refresh_token_whitelist(created_at);
COMMIT;
