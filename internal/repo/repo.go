package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"logstream/internal/database"
	pb "logstream/pkg/api/logstream"
)

type repo struct {
	db *sql.DB
}

type Repo interface {
	// RunInTx - run in transaction
	RunInTx(ctx context.Context, f func(ctx context.Context) error) error

	// GetLog - get log
	GetLog(ctx context.Context, id int) (*pb.Log, error)

	// GetLogs - get logs by filter
	GetLogs(ctx context.Context, level int32, keyword string, startTime int64, endTime int64) ([]*pb.Log, error)

	// AddLog - add log
	AddLog(ctx context.Context, log *pb.Log) (int64, error)

	// AddLogs - add logs
	AddLogs(ctx context.Context, logs []*pb.Log) (int64, error)
}

func NewRepo(db *sql.DB) Repo {
	return &repo{
		db: db,
	}
}

func (r *repo) RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	return database.RunInTx(ctx, r.db, f)
}

func (r *repo) GetLog(ctx context.Context, id int) (*pb.Log, error) {
	db := database.FromContext(ctx, r.db)

	var log pb.Log
	query := "SELECT * FROM logs WHERE id = $1 LIMIT 1"
	err := db.QueryRowContext(ctx, query, id).Scan(
		&log.Id,
		&log.Source,
		&log.Level,
		&log.Message,
		&log.Timestamp,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get log: %v", err)
	}

	return nil, nil
}

func (r *repo) GetLogs(ctx context.Context, level int32, keyword string, startTime int64, endTime int64) ([]*pb.Log, error) {
	db := database.FromContext(ctx, r.db)

	query := "SELECT id, source, lvl, message, created_at FROM logs WHERE created_at >= $1 AND created_at <= $2"
	rows, err := db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %v", err)
	}
	defer rows.Close()

	var logs []*pb.Log
	for rows.Next() {
		var log pb.Log
		if err := rows.Scan(&log.Id, &log.Source, &log.Level, &log.Message, &log.Timestamp); err != nil {
			//return nil, fmt.Errorf("failed to scan log: %v", err)
			continue
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %v", err)
	}

	return logs, nil
}

func (r *repo) AddLog(ctx context.Context, log *pb.Log) (int64, error) {
	db := database.FromContext(ctx, r.db)

	var id int64
	query := "INSERT INTO logs (source, source_id, lvl, message, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := db.QueryRowContext(ctx, query, log.Source, log.Id, log.Level, log.Message, log.Timestamp).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to add log: %v", err)
	}

	return id, nil
}

func (r *repo) AddLogs(ctx context.Context, logs []*pb.Log) (int64, error) {
	return 0, nil
}
