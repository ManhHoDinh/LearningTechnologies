package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	 "google.golang.org/grpc/credentials"
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

func (s *greeterServer) GreetMany(req *pb.GreetManyRequest, srv pb.Greeter_GreetManyServer) error {
    if req.GetName() == "" {
        return status.Error(codes.InvalidArgument, "name must not be empty")
    }

    ctx := srv.Context() // you can still access context/cancellation

    for i := 1; i <= 5; i++ {
        select {
        case <-ctx.Done():
            return status.Error(codes.Canceled, "client canceled")
        default:
        }

        msg := fmt.Sprintf("Hello #%d, %s 👋", i, req.GetName())

        // Use the generated Send method (type-safe), not SendMsg
        if err := srv.Send(&pb.GreetChunk{Message: msg}); err != nil {
            return err
        }

        time.Sleep(300 * time.Millisecond)
    }
    return nil
}

func (s *greeterServer) UploadNames(stream pb.Greeter_UploadNamesServer) error {
    var count int32
    for {
        n, err := stream.Recv()
        if err == io.EOF {
            // One final response
            return stream.SendAndClose(&pb.Summary{Count: count})
        }
        if err != nil {
            return status.Errorf(codes.Internal, "recv error: %v", err)
        }
        if n.GetValue() == "" {
            return status.Error(codes.InvalidArgument, "name must not be empty")
        }
        count++
        // Optional: bail out if client cancels
        select {
			case <-stream.Context().Done():
				return status.Error(codes.Canceled, "client canceled")
			default:
        }
    }
}
func (s *greeterServer) Chat(stream pb.Greeter_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil // client closed send
		}
		if err != nil {
			return status.Errorf(codes.Internal, "recv error: %v", err)
		}
		if in.GetText() == "" {
			if sendErr := stream.Send(&pb.ChatMsg{Text: "please send non-empty text"}); sendErr != nil {
				return sendErr
			}
			continue
		}
		// simple echo + timestamp
		out := &pb.ChatMsg{Text: fmt.Sprintf("server echo: %q @ %s", in.GetText(), time.Now().Format(time.RFC3339))}
		if err := stream.Send(out); err != nil {
			return err
		}
	}
}

func main() {
  lis, err := net.Listen("tcp", ":50051")
  if err != nil { log.Fatalf("listen: %v", err) }

  // NEW: TLS creds
  creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
  if err != nil { log.Fatalf("tls: %v", err) }

  s := grpc.NewServer(grpc.Creds(creds)) // secure server
  pb.RegisterGreeterServer(s, &greeterServer{})

  log.Println("gRPC TLS server listening on :50051")
  if err := s.Serve(lis); err != nil { log.Fatalf("serve: %v", err) }
}