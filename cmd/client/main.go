package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	logger "log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"logstream/internal/config"
	pb "logstream/pkg/api/logstream"
)

const (
	configPath = "config/local.yaml"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		logger.Fatalf("failed to init connection: %v", err)
	}
	defer conn.Close()

	client := pb.NewLogsServiceClient(conn)

	//saveLog(client, &pb.SaveLogRequest{
	//	Log: &pb.Log{
	//		Source:    "api",
	//		Level:     0,
	//		Message:   "Something happened",
	//		Timestamp: 100000,
	//	},
	//})

	//reqs := []*pb.SaveLogRequest{
	//	{
	//		Log: &pb.Log{Source: "API", Level: 0, Message: "First log", Timestamp: time.Now().Unix()},
	//	},
	//	{
	//		Log: &pb.Log{Source: "API", Level: 0, Message: "Second log", Timestamp: time.Now().Unix()},
	//	},
	//	{
	//		Log: &pb.Log{Source: "API", Level: 0, Message: "Third log", Timestamp: time.Now().Unix()},
	//	},
	//	{
	//		Log: &pb.Log{Source: "API", Level: 0, Message: "Fourth log", Timestamp: time.Now().Unix()},
	//	},
	//	{
	//		Log: &pb.Log{Source: "API", Level: 0, Message: "Fifth log", Timestamp: time.Now().Unix()},
	//	},
	//}
	//saveLogStream(client, reqs)

	listLog(client, &pb.ListLogRequest{
		Id: 7,
	})

	//listLogStream(client, []*pb.ListLogRequest{
	//	{Id: 7},
	//	{Id: 8},
	//	{Id: 9},
	//	{Id: 10},
	//	{Id: 11},
	//})
	//
	//listLogs(client, &pb.ListLogsRequest{
	//	StartTime: 1,
	//	EndTime:   10000000000,
	//})
	//
	//listLogsStream(client, &pb.ListLogsStreamRequest{
	//	StartTime: time.Now().AddDate(0, 0, -1).Unix(),
	//	EndTime:   time.Now().Unix(),
	//})
}

func saveLog(client pb.LogsServiceClient, req *pb.SaveLogRequest) {
	logger.Println("SaveLog: started...")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := client.SaveLog(ctx, req)
	if err != nil {
		logger.Printf("SaveLog: failed: %v\n", err)
		return
	}

	logger.Printf("SaveLog: received ID %d", resp.GetId())
}

func saveLogStream(client pb.LogsServiceClient, reqs []*pb.SaveLogRequest) {
	logger.Println("SaveLogStream: started...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.SaveLogStream(ctx)
	if err != nil {
		logger.Printf("SaveLogStream: failed: %v\n", err)
		return
	}

	doneCh := make(chan error)

	go func() {
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				doneCh <- nil
				return
			}
			if err != nil {
				doneCh <- fmt.Errorf("SaveLogStream: failed to receive resp: %v", err)
				return
			}

			logger.Printf("SaveLogStream: received ID %d", resp.GetId())
		}
	}()

	for _, req := range reqs {
		if err := stream.Send(req); err != nil {
			logger.Printf("SaveLogStream: failed to send req: %v\n", err)
		}
		logger.Printf("SaveLogStream: sent req with log: %v\n", req.GetLog())
		time.Sleep(500 * time.Millisecond)
	}

	if err := stream.CloseSend(); err != nil {
		logger.Printf("SaveLogStream: failed to close stream: %v\n", err)
	}

	if err := <-doneCh; err != nil {
		logger.Printf("SaveLogStream: %v\n", err)
	}
}

func listLog(client pb.LogsServiceClient, req *pb.ListLogRequest) {
	logger.Println("ListLog: started...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListLog(ctx, req)
	if err != nil {
		logger.Printf("ListLog: failed: %v\n", err)
		return
	}

	logger.Printf("ListLog: received log: %+v", resp.GetLog())
}

func listLogStream(client pb.LogsServiceClient, reqs []*pb.ListLogRequest) {
	logger.Println("ListLogStream: started...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.ListLogStream(ctx)
	if err != nil {
		logger.Printf("ListLogStream: failed: %v\n", err)
		return
	}

	doneCh := make(chan error)

	go func() {
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				doneCh <- nil
				return
			}
			if err != nil {
				doneCh <- fmt.Errorf("ListLogStream: failed to receive resp: %v", err)
				return
			}

			logger.Printf("ListLogStream: received log: %+v", resp.GetLog())
		}
	}()

	for _, req := range reqs {
		if err := stream.Send(req); err != nil {
			logger.Printf("ListLogStream: failed to send req: %v\n", err)
		}
		logger.Printf("ListLogStream: sent req with id: %v\n", req.GetId())
		time.Sleep(500 * time.Millisecond)
	}

	if err := stream.CloseSend(); err != nil {
		logger.Printf("ListLogStream: failed to close stream: %v\n", err)
	}

	if err := <-doneCh; err != nil {
		logger.Printf("ListLogStream: %v\n", err)
	}
}

func listLogs(client pb.LogsServiceClient, req *pb.ListLogsRequest) {
	logger.Println("ListLogs: started...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListLogs(ctx, req)
	if err != nil {
		logger.Printf("ListLogs: failed: %v\n", err)
		return
	}

	logs := resp.GetLogs()

	logger.Println("ListLogs: received logs:")
	for _, log := range logs {
		logger.Printf("Log: %+v", log)
	}
}

func listLogsStream(client pb.LogsServiceClient, req *pb.ListLogsStreamRequest) {
	logger.Println("ListLogsStream: started...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.ListLogsStream(ctx, req)
	if err != nil {
		logger.Printf("ListLogsStream: failed: %v\n", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			logger.Printf("ListLogsStream: failed: %v\n", err)
			return
		}

		logger.Printf("ListLogsStream: received log: %+v", resp.GetLog())
	}
}
