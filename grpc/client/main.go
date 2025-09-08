package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"

	pb "example.com/hello/hello" // keep this in sync with your module path
)

func main() {
	// For local dev we’ll use insecure; we’ll switch to TLS later.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	// Add a deadline (timeout) and some metadata (like headers)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-client", "go-grpc-demo")

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: ""})
	if err != nil {
		// Proper gRPC error handling
		st, ok := status.FromError(err)
		if ok {
			log.Printf("rpc failed: code=%s msg=%s", st.Code(), st.Message())
			if st.Code() == codes.DeadlineExceeded {
				log.Println("hint: increase timeout or speed up server")
			}
		} else {
			log.Printf("non-gRPC error: %v", err)
		}
		return
	}

	fmt.Println(resp.GetMessage())
}
