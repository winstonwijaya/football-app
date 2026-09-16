-- Add enum for status of the match
DROP TYPE IF EXISTS match_status;
CREATE TYPE match_status AS ENUM ('Scheduled', 'Played', 'Cancelled');

CREATE TABLE matches (
    id             BIGSERIAL PRIMARY KEY,
    home_team_id   BIGINT NOT NULL REFERENCES teams(id),
    away_team_id   BIGINT NOT NULL REFERENCES teams(id),
    match_datetime TIMESTAMPTZ NOT NULL,
    home_score     SMALLINT,
    away_score     SMALLINT,
    status         match_status NOT NULL DEFAULT 'Scheduled',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by     BIGINT REFERENCES users(id),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by     BIGINT REFERENCES users(id),
    deleted_at     TIMESTAMPTZ,
    deleted_by     BIGINT REFERENCES users(id),

    CHECK (home_team_id <> away_team_id)
);

CREATE INDEX idx_matches_home_team_datetime
    ON matches (home_team_id, match_datetime) WHERE deleted_at IS NULL;

CREATE INDEX idx_matches_away_team_datetime
    ON matches (away_team_id, match_datetime) WHERE deleted_at IS NULL;