package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	GetLog(ctx context.Context, id int32) (*pb.Log, error)

	// GetLogs - get logs by filter
	GetLogs(ctx context.Context, source string, level int32, startTime, endTime int64) ([]*pb.Log, error)

	// AddLog - add log
	AddLog(ctx context.Context, log *pb.Log) (int32, error)

	// AddLogs - add logs
	AddLogs(ctx context.Context, logs []*pb.Log) ([]int32, error)
}

func NewRepo(db *sql.DB) Repo {
	return &repo{
		db: db,
	}
}

func (r *repo) RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	return database.RunInTx(ctx, r.db, f)
}

func (r *repo) GetLog(ctx context.Context, id int32) (*pb.Log, error) {
	db := database.FromContext(ctx, r.db)

	var log pb.Log
	query := "SELECT id, source, lvl, message, created_at FROM logs WHERE id = $1"
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

	return &log, nil
}

func (r *repo) GetLogs(ctx context.Context, source string, level int32, startTime, endTime int64) ([]*pb.Log, error) {
	if level > 2 {
		return nil, fmt.Errorf("invalid log level: should be 0 (INFO), 1 (WARN), 2 (ERROR)")
	}

	db := database.FromContext(ctx, r.db)

	query := "SELECT id, source, lvl, message, created_at FROM logs WHERE source = $1 AND level = $2 AND created_at >= $3 AND created_at <= $4"
	rows, err := db.QueryContext(ctx, query, source, level, startTime, endTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
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

func (r *repo) AddLog(ctx context.Context, log *pb.Log) (int32, error) {
	if log.Level > 2 {
		return 0, fmt.Errorf("invalid log level: should be 0 (INFO), 1 (WARN), 2 (ERROR)")
	}

	db := database.FromContext(ctx, r.db)

	var id int32
	query := "INSERT INTO logs (source, lvl, message, created_at) VALUES ($1, $2, $3, $4) RETURNING id"
	err := db.QueryRowContext(ctx, query, log.Source, log.Level, log.Message, log.Timestamp).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to add log: %v", err)
	}

	return id, nil
}

func (r *repo) AddLogs(ctx context.Context, logs []*pb.Log) ([]int32, error) {
	if len(logs) == 0 {
		return nil, fmt.Errorf("no logs to add")
	}

	db := database.FromContext(ctx, r.db)

	query := "INSERT INTO logs (source, lvl, message, created_at) VALUES "
	values := make([]interface{}, 0, len(logs)*4)
	placeholders := make([]string, len(logs))
	for i, log := range logs {
		base := i * 4
		placeholders[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
		values = append(values, log.Source, log.Level, log.Message, log.Timestamp)
	}
	query += strings.Join(placeholders, ", ") + " RETURNING id"

	rows, err := db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to add logs: %v", err)
	}
	defer rows.Close()

	ids := make([]int32, 0, len(logs))
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan id: %v", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %v", err)
	}

	if len(ids) != len(logs) {
		return nil, fmt.Errorf("inserted %d logs, expected %d", len(ids), len(logs))
	}

	return ids, nil
}
