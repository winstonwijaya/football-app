package dto

// GoalInput — one goal within a result report. TeamID is the side credited
// with the goal (not necessarily the scoring player's own team — for an
// own goal, TeamID is the *benefiting* side while PlayerID belongs to the
// opposing team).
type GoalInput struct {
	TeamID    int64 `json:"team_id" binding:"required"`
	PlayerID  int64 `json:"player_id" binding:"required"`
	Minute    int16 `json:"minute" binding:"required,gte=1,lte=130"`
	IsOwnGoal bool  `json:"is_own_goal"`
}

// ReportResultRequest — POST /api/v1/matches/:id/result. Scores aren't
// "required" (0 is a valid score, and required treats zero as empty) —
// gte=0 is the actual constraint.
type ReportResultRequest struct {
	HomeScore int16       `json:"home_score" binding:"gte=0"`
	AwayScore int16       `json:"away_score" binding:"gte=0"`
	Goals     []GoalInput `json:"goals" binding:"dive"`
}
