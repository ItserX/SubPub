package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/ItserX/subpub/api/proto"
	"github.com/ItserX/subpub/internal/config"
	"github.com/ItserX/subpub/internal/service"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := log.New(os.Stdout, "[PubSub] ", log.LstdFlags)
	if cfg.Log.File != "" {
		f, err := os.OpenFile(cfg.Log.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		logger.SetOutput(f)
	}

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pubSubService := service.NewPubSubService(logger)
	pb.RegisterPubSubServer(grpcServer, pubSubService)
	reflection.Register(grpcServer)

	go func() {
		logger.Printf("Server listening on %s", addr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("Failed to serve: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Println("Initiating graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan bool)
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		logger.Println("Shutdown timeout exceeded, forcing server stop")
		grpcServer.Stop()
	case <-done:
		logger.Println("Server stopped gracefully")
	}
}
