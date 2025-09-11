package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "example.com/hello/hello" // keep this in sync with your module path
)

func main() {
	// Trust the server's self-signed cert
	pem, err := os.ReadFile("server.crt")
	if err != nil { log.Fatalf("read cert: %v", err) }
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) { log.Fatal("bad cert") }

	creds := credentials.NewClientTLSFromCert(pool, "localhost") // must match CN

  	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(creds))
  	if err != nil { log.Fatalf("dial: %v", err) }
	defer conn.Close()
	client := pb.NewGreeterClient(conn)

	// Add a deadline (timeout) and some metadata (like headers)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-client", "go-grpc-demo")

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "gopher"})
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
	// --- Server streaming example ---
	streamCtx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	stream, err := client.GreetMany(streamCtx, &pb.GreetManyRequest{Name: "streamy gopher"})
	if err != nil {
		log.Fatalf("GreetMany start: %v", err)
	}
	for {
		chunk, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			st, ok := status.FromError(recvErr)
			if ok {
				log.Printf("stream error: %s %s", st.Code(), st.Message())
			} else {
				log.Printf("recv error: %v", recvErr)
			}
			break
		}
		fmt.Println("chunk:", chunk.GetMessage())
	}
	// --- Client streaming example ---
	cctx, ccancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer ccancel()

	up, err := client.UploadNames(cctx)
	if err != nil {
		log.Fatalf("start UploadNames: %v", err)
	}

	// Send several names
	names := []string{"Ada", "Brian", "Charlotte"}
	for _, v := range names {
		if err := up.Send(&pb.Name{Value: v}); err != nil {
			log.Fatalf("send: %v", err)
		}
	}

	// Close the send side and receive the summary
	sum, err := up.CloseAndRecv()
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("upload error: %s %s", st.Code(), st.Message())
		} else {
			log.Printf("upload error: %v", err)
		}
		return
	}
	fmt.Printf("uploaded count: %d\n", sum.GetCount())
	// --- Bi-directional streaming example ---
chatCtx, chatCancel := context.WithTimeout(context.Background(), 5*time.Second)
defer chatCancel()

chat, err := client.Chat(chatCtx)
if err != nil {
	log.Fatalf("start Chat: %v", err)
}

// Interleave send/recv for a simple request/response rhythm
messages := []string{"hi", "how are you?", "", "great, thanks", "bye"}
for _, m := range messages {
	if err := chat.Send(&pb.ChatMsg{Text: m}); err != nil {
		log.Fatalf("chat send: %v", err)
	}
	reply, rerr := chat.Recv()
	if rerr == io.EOF {
		break
	}
	if rerr != nil {
		st, ok := status.FromError(rerr)
		if ok {
			log.Printf("chat recv error: %s %s", st.Code(), st.Message())
		} else {
			log.Printf("chat recv error: %v", rerr)
		}
		break
	}
	fmt.Println("chat reply:", reply.GetText())
}

// Close the send side and try to read any trailing messages (usually none here)
if err := chat.CloseSend(); err != nil {
	log.Printf("close send: %v", err)
}


}
