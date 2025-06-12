package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	logger "log"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"logstream/internal/repo"
	pb "logstream/pkg/api/logstream"
)

type server struct {
	pb.UnimplementedLogsServiceServer

	r repo.Repo
}

func NewServer(db *sql.DB) *server {
	r := repo.NewRepo(db)
	return &server{
		r: r,
	}
}

// SaveLog implements pb.LogsServiceServer
func (s *server) SaveLog(ctx context.Context, req *pb.SaveLogRequest) (*pb.SaveLogResponse, error) {
	logger.Println("SaveLog: received")

	if err := validateSaveLogRequest(req); err != nil {
		return nil, err
	}

	id, err := s.r.AddLog(ctx, req.Log)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	return &pb.SaveLogResponse{
		Id: id,
	}, nil
}

// SaveLogsStream implements pb.LogsServiceServer
func (s *server) SaveLogsStream(stream pb.LogsService_SaveLogsStreamServer) error {
	logger.Println("SaveLogsStream: received")

	var (
		rpcError error
	)
	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				rpcError = err
			}
			break
		}

		log := msg.GetLog()

		id, err := s.r.AddLog(stream.Context(), log)
		if err != nil {
			rpcError = err
		}

		if err := stream.Send(&pb.SaveLogsStreamResponse{Id: id}); err != nil {
			rpcError = err
		}

		if rpcError != nil {
			break
		}
	}

	return rpcError
}

// ListLogs implements pb.LogsServiceServer
func (s *server) ListLogs(ctx context.Context, req *pb.ListLogsRequest) (*pb.ListLogsResponse, error) {
	logger.Println("ListLogs: received")

	if err := validateListLogsRequest(req); err != nil {
		return nil, err
	}

	level := int32(req.GetLevel().Number())
	keyword := req.GetKeyword()
	startTime := req.GetStartTime()
	endTime := req.GetEndTime()

	logs, err := s.r.GetLogs(ctx, level, keyword, startTime, endTime)
	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}

	if len(logs) == 0 {
		st := status.New(codes.NotFound, "logs not found")
		return nil, st.Err()
	}

	return &pb.ListLogsResponse{
		Logs: logs,
	}, nil
}

// ListLogsStream implements pb.LogsServiceServer
func (s *server) ListLogsStream(req *pb.ListLogsStreamRequest, stream pb.LogsService_ListLogsStreamServer) error {
	logger.Println("ListLogsStream: received")

	var (
		rpcError error
		wg       sync.WaitGroup
	)

	level := int32(req.GetLevel().Number())
	keyword := req.GetKeyword()
	startTime := req.GetStartTime()
	endTime := req.GetEndTime()

	logCh := make(chan *pb.Log, 1)
	logs, err := s.r.GetLogs(stream.Context(), level, keyword, startTime, endTime)
	if err != nil {
		rpcError = err
		return rpcError
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, log := range logs {
			logCh <- log
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for localLog := range logCh {
			log := &pb.Log{
				Id:        localLog.Id,
				Source:    localLog.Source,
				Level:     localLog.Level,
				Message:   localLog.Message,
				Timestamp: localLog.Timestamp,
			}
			resp := &pb.ListLogsStreamResponse{
				Log: log,
			}
			if err := stream.Send(resp); err != nil {
				rpcError = err
				return
			}
		}
	}()

	wg.Wait()
	close(logCh)

	return rpcError
}
