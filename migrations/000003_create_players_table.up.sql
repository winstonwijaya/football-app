-- Add enum for the player positions, can be adjusted as needed
CREATE TYPE player_positions AS ENUM ('Goalkeeper', 'Fullback', 'Centre Back', 'Defensive Midfielder', 'Central Midfielder', 'Attacking Midfielder', 'Winger', 'Striker');

CREATE TABLE players (
    id           BIGSERIAL PRIMARY KEY,
    team_id      BIGINT NOT NULL REFERENCES teams(id),
    name         VARCHAR(150) NOT NULL,
    height    NUMERIC(5,2),
    weight    NUMERIC(5,2),
    position     player_positions NOT NULL,
    squad_number SMALLINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by   BIGINT REFERENCES users(id),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by   BIGINT REFERENCES users(id),
    deleted_at   TIMESTAMPTZ,
    deleted_by   BIGINT REFERENCES users(id)
);

CREATE INDEX idx_players_team_id ON players (team_id) WHERE deleted_at IS NULL;
-- Add notes for height and weight columns to clarify the units
COMMENT ON COLUMN players.height IS 'Height in centimeters';
COMMENT ON COLUMN players.weight IS 'Weight in kilograms';

-- Add unique index for team_id and squad_number for active records
CREATE UNIQUE INDEX uq_players_team_squad
    ON players (team_id, squad_number)
    WHERE deleted_at IS NULL;
