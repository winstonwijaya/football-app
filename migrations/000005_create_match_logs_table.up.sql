-- Append-only match-event log. Only Goal is used right now as the action, but can be adjusted later to include other events, e.g. yellow cards
CREATE TABLE match_logs (
    id          BIGSERIAL PRIMARY KEY,
    match_id    BIGINT NOT NULL REFERENCES matches(id),
    team_id     BIGINT NOT NULL REFERENCES teams(id),
    player_id   BIGINT NOT NULL REFERENCES players(id),
    action      VARCHAR(20) NOT NULL,
    minute      SMALLINT NOT NULL,
    is_own_goal BOOLEAN NOT NULL DEFAULT false,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  BIGINT REFERENCES users(id)
);

CREATE INDEX idx_match_logs_match_id ON match_logs (match_id);
