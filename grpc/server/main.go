package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "example.com/hello/hello" // adjust if your module path differs
)

type greeterServer struct {
	pb.UnimplementedGreeterServer
}

func (s *greeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	// Basic validation + gRPC-style error
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name must not be empty")
	}

	// Respect deadlines (if any)
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "request canceled or timed out")
	default:
	}

	msg := fmt.Sprintf("Hello, %s 👋", req.GetName())
	return &pb.HelloReply{Message: msg}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &greeterServer{})

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
