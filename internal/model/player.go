package model

// PlayerPosition mirrors the player_positions Postgres enum
// (migrations/000003_create_players_table.up.sql). Values must match
// exactly, including spacing and casing.
type PlayerPosition string

const (
	PositionGoalkeeper          PlayerPosition = "Goalkeeper"
	PositionFullback            PlayerPosition = "Fullback"
	PositionCentreBack          PlayerPosition = "Centre Back"
	PositionDefensiveMidfielder PlayerPosition = "Defensive Midfielder"
	PositionCentralMidfielder   PlayerPosition = "Central Midfielder"
	PositionAttackingMidfielder PlayerPosition = "Attacking Midfielder"
	PositionWinger              PlayerPosition = "Winger"
	PositionStriker             PlayerPosition = "Striker"
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
