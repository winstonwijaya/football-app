package model

// PlayerPosition mirrors the player_positions Postgres enum
// (migrations/000003_create_players_table.up.sql). Values must match
// exactly, including spacing and casing.
type PlayerPosition string

const (
	PositionForward    PlayerPosition = "Forward"
	PositionMidfielder PlayerPosition = "Midfielder"
	PositionDefender   PlayerPosition = "Defender"
	PositionGoalkeeper PlayerPosition = "Goalkeeper"
)

type Player struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	TeamID      int64          `gorm:"column:team_id;not null"`
	Name        string         `gorm:"column:name;not null"`
	HeightCM    *float64       `gorm:"column:height"`
	WeightKG    *float64       `gorm:"column:weight"`
	Position    PlayerPosition `gorm:"column:position;type:player_positions;not null"`
	SquadNumber int16          `gorm:"column:squad_number;not null"`

	Audit

	Team *Team `gorm:"foreignKey:TeamID"`
}

func (Player) TableName() string {
	return "players"
}
