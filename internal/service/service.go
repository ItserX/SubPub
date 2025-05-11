package service

import (
	"context"
	"log"

	"github.com/ItserX/subpub/internal/subpub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/ItserX/subpub/api/proto"
)

type PubSubService struct {
	pb.UnimplementedPubSubServer
	subPub subpub.SubPub
	logger *log.Logger
}

func NewPubSubService(logger *log.Logger) *PubSubService {
	return &PubSubService{
		subPub: subpub.NewSubPub(),
		logger: logger,
	}
}

func (s *PubSubService) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	if req.Key == "" {
		return status.Error(codes.InvalidArgument, "key is empty")
	}

	s.logger.Printf("New subscription key: %s", req.Key)

	sub, err := s.subPub.Subscribe(req.Key, func(msg interface{}) {
		if data, ok := msg.(string); ok {
			event := &pb.Event{Data: data}
			if err := stream.Send(event); err != nil {
				s.logger.Printf("Error sending event: %v", err)
			}
		}
	})
	if err != nil {
		return status.Error(codes.Internal, "failed to subscribe")
	}
	defer sub.Unsubscribe()

	<-stream.Context().Done()
	return nil
}

func (s *PubSubService) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is empty")
	}

	s.logger.Printf("Publishing message for key: %s", req.Key)

	if err := s.subPub.Publish(req.Key, req.Data); err != nil {
		return nil, status.Error(codes.Internal, "failed to publish message")
	}

	return &emptypb.Empty{}, nil
}
