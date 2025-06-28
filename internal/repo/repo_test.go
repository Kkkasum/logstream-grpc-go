package repo_test

import (
	"context"
	"database/sql"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"logstream/internal/repo"
	pb "logstream/pkg/api/logstream"
)

type Suite struct {
	suite.Suite

	db   *sql.DB
	mock sqlmock.Sqlmock
	r    repo.Repo
	ctx  context.Context
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}

func (s *Suite) SetupSuite() {
	var err error
	s.db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.r = repo.NewRepo(s.db)
	s.ctx = context.Background()
}

func (s *Suite) AfterTest(suiteName, testName string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *Suite) TearDownSuite() {
	s.db.Close()
}

func (s *Suite) TestGetLog() {
	testCases := []struct {
		name        string
		inputLogId  int32
		mockSetup   func(mock sqlmock.Sqlmock)
		expectedLog *repo.Log
		expectedErr string
	}{
		{
			name:       "get log",
			inputLogId: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE id = $1`)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "source", "lvl", "message", "created_at"}).
						AddRow(1, "test-source", pb.Level_LEVEL_INFO, "test message", time.Now().Unix()))
			},
			expectedLog: &repo.Log{
				Id:        func() *int32 { id := int32(1); return &id }(),
				Source:    "test-source",
				Level:     int32(pb.Level_LEVEL_INFO),
				Message:   "test message",
				CreatedAt: time.Now().Unix(),
			},
		},
		{
			name:       "log not found",
			inputLogId: 2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE id = $1`)).
					WithArgs(2).
					WillReturnError(sql.ErrNoRows)
			},
			expectedLog: nil,
			expectedErr: "record not found",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			actualLog, err := s.r.GetLog(s.ctx, tc.inputLogId)

			if tc.expectedErr == "" {
				require.NoError(t, err)
				assert.NotNil(t, actualLog)
				assert.True(t, reflect.DeepEqual(tc.expectedLog, actualLog))
			} else {
				assert.Nil(t, actualLog)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

func (s *Suite) TestGetLogs() {
	testCases := []struct {
		name           string
		inputSource    string
		inputLevel     int32
		inputStartTime int64
		inputEndTime   int64
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedLogs   []*repo.Log
		expectedErr    string
	}{
		{
			name:           "get logs",
			inputSource:    "test-source",
			inputLevel:     1,
			inputStartTime: 10000,
			inputEndTime:   1000000,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE source = $1 AND lvl = $2 AND created_at >= $3 AND created_at <= $4`)).
					WithArgs("test-source", 1, 10000, 1000000).
					WillReturnRows(sqlmock.NewRows([]string{"id", "source", "lvl", "message", "created_at"}).
						AddRow(1, "test-source", 1, "test message 1", 10000).
						AddRow(2, "test-source", 1, "test message 2", 10001))
			},
			expectedLogs: []*repo.Log{
				{
					Id:        func() *int32 { id := int32(1); return &id }(),
					Source:    "test-source",
					Level:     int32(pb.Level_LEVEL_WARN),
					Message:   "test message 1",
					CreatedAt: 10000,
				},
				{
					Id:        func() *int32 { id := int32(2); return &id }(),
					Source:    "test-source",
					Level:     int32(pb.Level_LEVEL_WARN),
					Message:   "test message 2",
					CreatedAt: 10001,
				},
			},
		},
		{
			name:        "invalid log level",
			inputLevel:  1000,
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedErr: "invalid log level",
		},
		{
			name: "logs not found",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE source = $1 AND lvl = $2 AND created_at >= $3 AND created_at <= $4`)).
					WithArgs("", 0, 0, 0).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: "record not found",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			actualLogs, err := s.r.GetLogs(s.ctx, tc.inputSource, tc.inputLevel, tc.inputStartTime, tc.inputEndTime)

			if tc.expectedErr == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, tc.expectedLogs, actualLogs)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, actualLogs)
			}
		})
	}
}

func (s *Suite) TestAddLog() {
	testCases := []struct {
		name        string
		inputLog    *repo.Log
		mockSetup   func(mock sqlmock.Sqlmock)
		expectedId  int32
		expectedErr string
	}{
		{
			name: "add log",
			inputLog: &repo.Log{
				Source:    "test-source",
				Level:     int32(pb.Level_LEVEL_INFO),
				Message:   "test message",
				CreatedAt: time.Now().Unix(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO logs (source, lvl, message, created_at) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs("test-source", pb.Level_LEVEL_INFO, "test message", time.Now().Unix()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedId: 1,
		},
		{
			name: "invalid level",
			inputLog: &repo.Log{
				Source:    "test-source",
				Level:     1000,
				Message:   "test message",
				CreatedAt: time.Now().Unix(),
			},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedId:  0,
			expectedErr: "invalid log level",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			actualId, err := s.r.AddLog(s.ctx, tc.inputLog)

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
			assert.Equal(t, tc.expectedId, actualId)
		})
	}
}

func (s *Suite) TestAddLogs() {
	testCases := []struct {
		name        string
		inputLogs   []*repo.Log
		mockSetup   func(mock sqlmock.Sqlmock)
		expectedIds []int32
		expectedErr string
	}{
		{
			name: "add logs",
			inputLogs: []*repo.Log{
				{
					Source:    "test-source-1",
					Level:     int32(pb.Level_LEVEL_INFO),
					Message:   "test message 1",
					CreatedAt: time.Now().Unix(),
				},
				{
					Source:    "test-source-2",
					Level:     int32(pb.Level_LEVEL_WARN),
					Message:   "test message 2",
					CreatedAt: time.Now().Unix(),
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO logs (source, lvl, message, created_at) VALUES ($1, $2, $3, $4), ($5, $6, $7, $8) RETURNING id`)).
					WithArgs(
						"test-source-1", pb.Level_LEVEL_INFO, "test message 1", time.Now().Unix(),
						"test-source-2", pb.Level_LEVEL_WARN, "test message 2", time.Now().Unix(),
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			},
			expectedIds: []int32{1, 2},
		},
		{
			name:        "add zero logs",
			inputLogs:   []*repo.Log{},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedIds: nil,
			expectedErr: "no logs to add",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			actualIds, err := s.r.AddLogs(s.ctx, tc.inputLogs)

			if tc.expectedErr == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, tc.expectedIds, actualIds)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, actualIds)
			}
		})
	}
}
