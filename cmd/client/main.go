package main

import (
	"context"
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

	saveLog(client, &pb.SaveLogRequest{
		Log: &pb.Log{
			Id:        "zxc-123",
			Source:    "api",
			Level:     0,
			Message:   "Something happened",
			Timestamp: 100000,
		},
	})

	logs := []*pb.Log{
		{Id: "api-1", Source: "API", Level: 0, Message: "First log", Timestamp: time.Now().Unix()},
		{Id: "api-2", Source: "API", Level: 0, Message: "Second log", Timestamp: time.Now().Unix()},
		{Id: "api-3", Source: "API", Level: 0, Message: "Third log", Timestamp: time.Now().Unix()},
		{Id: "api-4", Source: "API", Level: 0, Message: "Fourth log", Timestamp: time.Now().Unix()},
		{Id: "api-5", Source: "API", Level: 0, Message: "Fifth log", Timestamp: time.Now().Unix()},
		{Id: "api-6", Source: "API", Level: 0, Message: "Sixth log", Timestamp: time.Now().Unix()},
	}
	saveLogsStream(client, logs)

	listLogs(client, &pb.ListLogsRequest{
		StartTime: 1,
		EndTime:   10000000000,
	})

	listLogsStream(client, &pb.ListLogsStreamRequest{
		StartTime: time.Now().AddDate(0, 0, -1).Unix(),
		EndTime:   time.Now().Unix(),
	})
}

func saveLog(client pb.LogsServiceClient, req *pb.SaveLogRequest) {
	logger.Println("SaveLog started...")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := client.SaveLog(ctx, req)
	if err != nil {
		logger.Fatalf("SaveLog failed: %v\n", err)
	}

	id := res.GetId()
	logger.Printf("SaveLog result: Log ID %d", id)
}

func saveLogsStream(client pb.LogsServiceClient, logs []*pb.Log) {
	logger.Println("SaveLogsStream started...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.SaveLogsStream(ctx)
	if err != nil {
		logger.Fatalf("SaveLogsStream failed: %v\n", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					close(waitc)
					break
				}
				logger.Fatalf("SaveLogsStream failed: %v\n", err)
			}
			logger.Println(in)
		}
	}()

	for _, log := range logs {
		req := &pb.SaveLogsStreamRequest{
			Log: log,
		}
		if err := stream.Send(req); err != nil {
			logger.Fatalf("SaveLogsStream: stream.Send(%v) failed: %v", log, err)
		}
	}
	stream.CloseSend()
	<-waitc
}

func listLogs(client pb.LogsServiceClient, req *pb.ListLogsRequest) {
	logger.Println("ListLogs started...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.ListLogs(ctx, req)
	if err != nil {
		logger.Fatalf("ListLogs failed: %v\n", err)
	}

	logs := res.GetLogs()

	logger.Println("Received logs:")
	for _, log := range logs {
		logTime := time.Unix(log.Timestamp, 0)
		logger.Printf("ID: %s, Source: %s, Message: %s, Time: %s", log.Id, log.Source, log.Message, logTime)
	}
}

func listLogsStream(client pb.LogsServiceClient, req *pb.ListLogsStreamRequest) {
	logger.Println("ListLogsStream started...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.ListLogsStream(ctx, req)
	if err != nil {
		logger.Fatalf("ListLogsStream failed: %v\n", err)
	}

	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatalf("ListLogsStream failed: %v\n", err)
		}

		log := res.GetLog()

		logTime := time.Unix(log.Timestamp, 0)
		logger.Printf("ID: %s, Source: %s, Message: %s, Time: %s", log.Id, log.Source, log.Message, logTime)
	}
}
