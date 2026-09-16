package model

import "time"

// MatchStatus mirrors the match_status Postgres enum
// (migrations/000004_create_matches_table.up.sql). This is lifecycle only —
// home win / away win / draw is derived from the scores at query time, never
// stored.
type MatchStatus string

const (
	MatchStatusScheduled MatchStatus = "Scheduled"
	MatchStatusPlayed    MatchStatus = "Played"
	MatchStatusCancelled MatchStatus = "Cancelled"
)

type Match struct {
	ID            int64       `gorm:"column:id;primaryKey"`
	HomeTeamID    int64       `gorm:"column:home_team_id;not null"`
	AwayTeamID    int64       `gorm:"column:away_team_id;not null"`
	MatchDatetime time.Time   `gorm:"column:match_datetime;not null"`
	HomeScore     *int16      `gorm:"column:home_score"`
	AwayScore     *int16      `gorm:"column:away_score"`
	Status        MatchStatus `gorm:"column:status;type:match_status;not null"`

	Audit

	HomeTeam *Team `gorm:"foreignKey:HomeTeamID"`
	AwayTeam *Team `gorm:"foreignKey:AwayTeamID"`
}

func (Match) TableName() string {
	return "matches"
}
