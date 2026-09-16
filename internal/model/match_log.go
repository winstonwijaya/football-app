package model

// MatchLogAction identifies the kind of event a match_logs row records.
// Only GOAL is implemented today; the column is plain text (no CHECK/enum)
// so cards, substitutions, etc. can be added later without a migration.
type MatchLogAction string

const (
	MatchLogActionGoal MatchLogAction = "GOAL"
)

type MatchLog struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	MatchID   int64          `gorm:"column:match_id;not null"`
	TeamID    int64          `gorm:"column:team_id;not null"`
	PlayerID  int64          `gorm:"column:player_id;not null"`
	Action    MatchLogAction `gorm:"column:action;not null"`
	Minute    int16          `gorm:"column:minute;not null"`
	IsOwnGoal bool           `gorm:"column:is_own_goal;not null"`

	CreateAudit

	Match  *Match  `gorm:"foreignKey:MatchID"`
	Team   *Team   `gorm:"foreignKey:TeamID"`
	Player *Player `gorm:"foreignKey:PlayerID"`
}

func (MatchLog) TableName() string {
	return "match_logs"
}
