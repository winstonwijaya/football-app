package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrConflict is returned by Create/Update when the write violates a unique
// constraint (e.g. squad number already taken within a team). Services
// translate this into apperror.Conflict.
var ErrConflict = errors.New("conflict")

const pgUniqueViolation = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgUniqueViolation
	}
	return false
}
