package repository

import "gorm.io/gorm"

// paginate applies OFFSET/LIMIT for the given page/limit. Callers are
// expected to have already normalized page/limit to sane values (see
// dto.PaginationQuery.Normalize).
func paginate(db *gorm.DB, page, limit int) *gorm.DB {
	offset := (page - 1) * limit
	return db.Offset(offset).Limit(limit)
}
