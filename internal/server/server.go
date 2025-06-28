package server

import (
	"context"
	"database/sql"
	"errors"
	"io"
	logger "log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"logstream/internal/database"
	"logstream/internal/repo"
	pb "logstream/pkg/api/logstream"
)

type Server struct {
	pb.UnimplementedLogsServiceServer

	r repo.Repo
}

func NewServer(db *sql.DB) *Server {
	r := repo.NewRepo(db)
	return &Server{
		r: r,
	}
}

// SaveLog implements pb.LogsServiceServer
func (s *Server) SaveLog(ctx context.Context, req *pb.SaveLogRequest) (*pb.SaveLogResponse, error) {
	logger.Println("SaveLog: received")

	if err := validateSaveLogRequest(req); err != nil {
		return nil, err
	}

	id, err := s.r.AddLog(ctx, repo.FromPbLog(req.Log))
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	return &pb.SaveLogResponse{
		Id: id,
	}, nil
}

// SaveLogStream implements pb.LogsServiceServer
func (s *Server) SaveLogStream(stream pb.LogsService_SaveLogStreamServer) error {
	logger.Println("SaveLogsStream: received")

	for {
		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "client context is done")
		default:
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			if err := validateSaveLogRequest(req); err != nil {
				return err
			}

			id, err := s.r.AddLog(stream.Context(), repo.FromPbLog(req.GetLog()))
			if err != nil {
				return status.Error(codes.Aborted, err.Error())
			}

			resp := &pb.SaveLogResponse{
				Id: id,
			}
			if err := stream.Send(resp); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

// ListLog implements pb.LogsServiceServer
func (s *Server) ListLog(ctx context.Context, req *pb.ListLogRequest) (*pb.ListLogResponse, error) {
	logger.Println("ListLog: received")

	if err := validateListLogRequest(req); err != nil {
		return nil, err
	}

	log, err := s.r.GetLog(ctx, req.Id)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListLogResponse{
		Log: log.ToPbLog(),
	}, nil
}

func (s *Server) ListLogStream(stream pb.LogsService_ListLogStreamServer) error {
	logger.Println("ListLogStream: received")

	for {
		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "client context is done")
		default:
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			if err := validateListLogRequest(req); err != nil {
				return err
			}

			log, err := s.r.GetLog(stream.Context(), req.GetId())
			if err != nil {
				if database.IsRecordNotFoundError(err) {
					return status.Error(codes.NotFound, err.Error())
				}
				return status.Error(codes.Internal, err.Error())
			}

			resp := &pb.ListLogResponse{
				Log: log.ToPbLog(),
			}
			if err := stream.Send(resp); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

// ListLogs implements pb.LogsServiceServer
func (s *Server) ListLogs(ctx context.Context, req *pb.ListLogsRequest) (*pb.ListLogsResponse, error) {
	logger.Println("ListLogs: received")

	if err := validateListLogsRequest(req); err != nil {
		return nil, err
	}

	logs, err := s.r.GetLogs(ctx, req.GetSource(), int32(req.GetLevel()), req.GetStartTime(), req.GetEndTime())
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	respLogs := make([]*pb.Log, len(logs))
	for idx, log := range logs {
		respLogs[idx] = log.ToPbLog()
	}

	return &pb.ListLogsResponse{
		Logs: respLogs,
	}, nil
}

// ListLogsStream implements pb.LogsServiceServer
func (s *Server) ListLogsStream(req *pb.ListLogsStreamRequest, stream pb.LogsService_ListLogsStreamServer) error {
	logger.Println("ListLogsStream: received")

	if err := validateListLogsStreamRequest(req); err != nil {
		return err
	}

	logs, err := s.r.GetLogs(stream.Context(), req.GetSource(), int32(req.GetLevel()), req.GetStartTime(), req.GetEndTime())
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return status.Error(codes.NotFound, err.Error())
		}
		return status.Error(codes.Internal, err.Error())
	}

	for _, log := range logs {
		resp := &pb.ListLogsStreamResponse{
			Log: log.ToPbLog(),
		}
		if err := stream.Send(resp); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}
