package database

import "errors"

var (
	ErrNotFound     = errors.New("record not found")
	ErrPKeyConflict = errors.New("primary key conflict")
)
