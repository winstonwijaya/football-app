package model

import "time"

// Audit holds the created/updated/deleted tracking columns shared by every mutable table.
type Audit struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	CreatedBy *int64     `gorm:"column:created_by"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;autoUpdateTime"`
	UpdatedBy *int64     `gorm:"column:updated_by"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	DeletedBy *int64     `gorm:"column:deleted_by"`
}

// CreateAudit is used by append-only tables (match_logs) without update/delete tracking.
type CreateAudit struct {
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	CreatedBy *int64    `gorm:"column:created_by"`
}
