package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/ItserX/subpub/api/proto"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "server address")
	mode := flag.String("mode", "both", "client mode: sub, pub, or both")
	key := flag.String("key", "test-key", "subscription/publish key")
	message := flag.String("message", "Hello!", "message to publish")
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPubSubClient(conn)
	ctx := context.Background()

	if *mode == "sub" || *mode == "both" {
		go func() {
			subscribeToEvents(ctx, client, *key)
		}()
	}

	if *mode == "pub" || *mode == "both" {
		time.Sleep(time.Second)
		publishEvents(ctx, client, *key, *message)
	}

	if *mode == "sub" || *mode == "both" {
		select {}
	}
}

func subscribeToEvents(ctx context.Context, client pb.PubSubClient, key string) {
	req := &pb.SubscribeRequest{Key: key}
	stream, err := client.Subscribe(ctx, req)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Printf("Subscribed to key: %s", key)
	for {
		event, err := stream.Recv()
		if err != nil {
			log.Printf("Stream error: %v", err)
			return
		}
		log.Printf("Received event: %s", event.Data)
	}
}

func publishEvents(ctx context.Context, client pb.PubSubClient, key, message string) {
	req := &pb.PublishRequest{
		Key:  key,
		Data: message,
	}

	_, err := client.Publish(ctx, req)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Printf("Published message '%s' to key '%s'\n", message, key)
}
