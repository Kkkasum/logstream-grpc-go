package server

import (
	"database/sql"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"

	pb "logstream/pkg/api/logstream"
)

type Suite struct {
	suite.Suite

	db     *sql.DB
	mock   sqlmock.Sqlmock
	server *Server
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}

func (s *Suite) SetupSuite() {
	var err error
	s.db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.server = NewServer(s.db)
}

func (s *Suite) AfterTest(suiteName, testName string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *Suite) TearDownSuite() {
	s.db.Close()
}

func (s *Suite) TestSaveLog() {
	testCases := []struct {
		name         string
		req          *pb.SaveLogRequest
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedResp *pb.SaveLogResponse
		expectedErr  string
	}{
		{
			name: "save log",
			req: &pb.SaveLogRequest{
				Log: &pb.Log{
					Source:    "test-source",
					Level:     pb.Level_LEVEL_INFO,
					Message:   "test message",
					Timestamp: time.Now().Unix(),
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO logs (source, lvl, message, created_at) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs("test-source", pb.Level_LEVEL_INFO, "test message", time.Now().Unix()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedResp: &pb.SaveLogResponse{
				Id: 1,
			},
		},
		{
			name:        "invalid request log",
			req:         &pb.SaveLogRequest{},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedErr: codes.InvalidArgument.String(),
		},
		{
			name: "invalid request log source",
			req: &pb.SaveLogRequest{
				Log: &pb.Log{
					Source:    "",
					Level:     pb.Level_LEVEL_INFO,
					Message:   "test message",
					Timestamp: time.Now().Unix(),
				},
			},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedErr: codes.InvalidArgument.String(),
		},
		{
			name: "invalid request log level",
			req: &pb.SaveLogRequest{
				Log: &pb.Log{
					Source:    "test-source",
					Level:     5,
					Message:   "test message",
					Timestamp: time.Now().Unix(),
				},
			},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedErr: codes.InvalidArgument.String(),
		},
		{
			name: "invalid request log message",
			req: &pb.SaveLogRequest{
				Log: &pb.Log{
					Source:    "test-source",
					Level:     pb.Level_LEVEL_INFO,
					Message:   "",
					Timestamp: time.Now().Unix(),
				},
			},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedErr: codes.InvalidArgument.String(),
		},
		{
			name: "invalid request log timestamp",
			req: &pb.SaveLogRequest{
				Log: &pb.Log{
					Source:    "test-source",
					Level:     pb.Level_LEVEL_INFO,
					Message:   "test message",
					Timestamp: 0,
				},
			},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedErr: codes.InvalidArgument.String(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			resp, err := s.server.SaveLog(t.Context(), tc.req)

			if tc.expectedErr == "" {
				require.NoError(t, err)
				assert.True(t, reflect.DeepEqual(tc.expectedResp, resp))
			} else {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

func (s *Suite) TestSaveLogStream() {}

func (s *Suite) TestListLog() {
	testCases := []struct {
		name         string
		req          *pb.ListLogRequest
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedResp *pb.ListLogResponse
		expectedErr  string
	}{
		{
			name: "list log",
			req: &pb.ListLogRequest{
				Id: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE id = $1`)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "source", "lvl", "message", "created_at"}).
						AddRow(1, "test-source", pb.Level_LEVEL_INFO, "test message", time.Now().Unix()))
			},
			expectedResp: &pb.ListLogResponse{
				Log: &pb.Log{
					Id:        func() *int32 { id := int32(1); return &id }(),
					Source:    "test-source",
					Level:     pb.Level_LEVEL_INFO,
					Message:   "test message",
					Timestamp: time.Now().Unix(),
				},
			},
		},
		{
			name: "log not found",
			req: &pb.ListLogRequest{
				Id: 42,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE id = $1`)).
					WithArgs(42).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: codes.NotFound.String(),
		},
		{
			name: "invalid request id",
			req: &pb.ListLogRequest{
				Id: 0,
			},
			mockSetup:   func(mock sqlmock.Sqlmock) {},
			expectedErr: codes.InvalidArgument.String(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			resp, err := s.server.ListLog(t.Context(), tc.req)

			if tc.expectedErr == "" {
				require.NoError(t, err)
				assert.True(t, reflect.DeepEqual(tc.expectedResp, resp))
			} else {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

func (s *Suite) TestListLogStream() {}

func (s *Suite) TestListLogs() {
	testCases := []struct {
		name         string
		req          *pb.ListLogsRequest
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedResp *pb.ListLogsResponse
		expectedErr  string
	}{
		{
			name: "list logs",
			req: &pb.ListLogsRequest{
				Source:    "test-source",
				Level:     pb.Level_LEVEL_WARN,
				StartTime: 10000,
				EndTime:   1000000,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE source = $1 AND lvl = $2 AND created_at >= $3 AND created_at <= $4`)).
					WithArgs("test-source", 1, 10000, 1000000).
					WillReturnRows(sqlmock.NewRows([]string{"id", "source", "lvl", "message", "created_at"}).
						AddRow(1, "test-source", pb.Level_LEVEL_WARN, "test message 1", 10000).
						AddRow(2, "test-source", pb.Level_LEVEL_WARN, "test message 2", 10001))
			},
			expectedResp: &pb.ListLogsResponse{
				Logs: []*pb.Log{
					{
						Id:        func() *int32 { id := int32(1); return &id }(),
						Source:    "test-source",
						Level:     pb.Level_LEVEL_WARN,
						Message:   "test message 1",
						Timestamp: 10000,
					},
					{
						Id:        func() *int32 { id := int32(2); return &id }(),
						Source:    "test-source",
						Level:     pb.Level_LEVEL_WARN,
						Message:   "test message 2",
						Timestamp: 10001,
					},
				},
			},
		},
		{
			name: "logs not found",
			req: &pb.ListLogsRequest{
				Source:    "test-source",
				Level:     pb.Level_LEVEL_WARN,
				StartTime: 10000,
				EndTime:   1000000,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, source, lvl, message, created_at FROM logs WHERE source = $1 AND lvl = $2 AND created_at >= $3 AND created_at <= $4`)).
					WithArgs("test-source", 1, 10000, 1000000).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: codes.NotFound.String(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			resp, err := s.server.ListLogs(t.Context(), tc.req)

			if tc.expectedErr == "" {
				require.NoError(t, err)
				assert.True(t, reflect.DeepEqual(tc.expectedResp, resp))
			} else {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

func (s *Suite) TestListLogsStream() {}
