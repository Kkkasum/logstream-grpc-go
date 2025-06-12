package database

import (
	"context"
	"database/sql"
	"fmt"
)

type contextKey string

const dbKey = contextKey("db")

func WithDB(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

func FromContext(ctx context.Context, db *sql.DB) *sql.DB {
	if ctx == nil {
		return db
	}
	if stored, ok := ctx.Value(dbKey).(*sql.DB); ok {
		return stored
	}
	return db
}

func RunInTx(ctx context.Context, db *sql.DB, f func(ctx context.Context) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start tx: %v", err)
	}

	ctx = WithDB(ctx, db)
	if err := f(ctx); err != nil {
		if err1 := tx.Rollback().Error; err1 != nil {
			return fmt.Errorf("failed to rollback tx: %v", err1)
		}
		return fmt.Errorf("failed to invoke func: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit tx: %v", err)
	}
	return nil
}
