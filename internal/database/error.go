package database

import (
	"database/sql"
	"errors"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrPKeyConflict = errors.New("primary key conflict")
)

func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, sql.ErrNoRows)
}
