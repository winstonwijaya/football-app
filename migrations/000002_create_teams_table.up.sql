CREATE TABLE teams (
    id               BIGSERIAL PRIMARY KEY,
    name             VARCHAR(100) NOT NULL,
    logo_url         TEXT,
    established_year SMALLINT,
    address          VARCHAR(255),
    city             VARCHAR(100),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by       BIGINT REFERENCES users(id),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by       BIGINT REFERENCES users(id),
    deleted_at       TIMESTAMPTZ,
    deleted_by       BIGINT REFERENCES users(id)
);

CREATE INDEX idx_teams_city ON teams (city) WHERE deleted_at IS NULL;
