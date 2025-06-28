package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"logstream/internal/config"
	"logstream/internal/database"
	"logstream/internal/server"
	pb "logstream/pkg/api/logstream"
)

const (
	configPath = "config/local.yaml"
)

func main() {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewDB(&cfg.DBConfig)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Server is listening on %v", addr)

	s := grpc.NewServer()
	pb.RegisterLogsServiceServer(s, server.NewServer(db))

	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	s.GracefulStop()
	s.Stop()
}
