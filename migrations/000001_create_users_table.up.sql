CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    first_name    VARCHAR(100) NOT NULL,
    last_name     VARCHAR(100) NOT NULL,
    username      VARCHAR(100) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by    BIGINT REFERENCES users(id),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by    BIGINT REFERENCES users(id),
    deleted_at    TIMESTAMPTZ,
    deleted_by    BIGINT REFERENCES users(id)
);

CREATE UNIQUE INDEX uq_users_username
    ON users (username)
    WHERE deleted_at IS NULL;
