package server

import (
	"context"
	"database/sql"
	"io"
	"net"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "logstream/pkg/api/logstream"
)

const bufSize = 1024 * 1024

type ServerStreamSuite struct {
	suite.Suite

	db     *sql.DB
	mock   sqlmock.Sqlmock
	server *grpc.Server
	conn   *grpc.ClientConn
	client pb.LogsServiceClient
}

func TestServerStreamSuite(t *testing.T) {
	suite.Run(t, &ServerStreamSuite{})
}

func (s *ServerStreamSuite) SetupSuite() {
	var err error
	s.db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	lis := bufconn.Listen(bufSize)
	s.server = grpc.NewServer()
	pb.RegisterLogsServiceServer(s.server, NewServer(s.db))

	go func() {
		if err := s.server.Serve(lis); err != nil {
			require.NoError(s.T(), err)
		}
	}()

	s.conn, err = grpc.NewClient(
		"passthrough:bufnet",
		grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(s.T(), err)

	s.client = pb.NewLogsServiceClient(s.conn)
}

func (s *ServerStreamSuite) AfterTest(suiteName, testName string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ServerStreamSuite) TearDownSuite() {
	s.server.Stop()
	s.conn.Close()
	s.db.Close()
}

func (s *ServerStreamSuite) TestSaveLogStream() {
	testCases := []struct {
		name         string
		reqs         []*pb.SaveLogRequest
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedResp []*pb.SaveLogResponse
	}{
		{
			name: "save logs",
			reqs: []*pb.SaveLogRequest{
				{
					Log: &pb.Log{
						Source:    "test-source",
						Level:     pb.Level_LEVEL_INFO,
						Message:   "test message 1",
						Timestamp: time.Now().Unix(),
					},
				},
				{
					Log: &pb.Log{
						Source:    "test-source",
						Level:     pb.Level_LEVEL_INFO,
						Message:   "test message 2",
						Timestamp: time.Now().Unix(),
					},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO logs (source, lvl, message, created_at) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs("test-source", pb.Level_LEVEL_INFO, "test message 1", time.Now().Unix()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO logs (source, lvl, message, created_at) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs("test-source", pb.Level_LEVEL_INFO, "test message 2", time.Now().Unix()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
			expectedResp: []*pb.SaveLogResponse{
				{
					Id: 1,
				},
				{
					Id: 2,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mockSetup(s.mock)

			stream, err := s.client.SaveLogStream(t.Context())
			require.NoError(t, err)

			doneCh := make(chan struct{})
			actualResp := make([]*pb.SaveLogResponse, 0, len(tc.expectedResp))

			go func() {
				defer close(doneCh)

				resp, err := stream.Recv()
				if err == io.EOF {
					return
				}
				require.NoError(t, err)

				actualResp = append(actualResp, resp)
			}()

			for _, req := range tc.reqs {
				err := stream.Send(req)
				require.NoError(t, err)
			}

			err = stream.CloseSend()
			require.NoError(t, err)

			<-doneCh

			for i, resp := range tc.expectedResp {
				t.Logf("%+v", resp)
				t.Logf("%+v", actualResp[i])
				assert.True(t, reflect.DeepEqual(resp, actualResp[i]))
			}
		})
	}
}

func (s *ServerStreamSuite) TestListLogStream() {}
