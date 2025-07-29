package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	greetv1 "github.com/andream16/gophercon-tutorial/tools/grpc/proto/greet/v1"
)

type server struct {
	greetv1.UnimplementedGreetServiceServer
}

func (s *server) Greet(_ context.Context, req *greetv1.GreetRequest) (*greetv1.GreetResponse, error) {
	return &greetv1.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := Main(ctx); err != nil {
		log.Fatal(err)
	}
}

func Main(ctx context.Context) error {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return errors.Errorf("could not create listener: %w", err)
	}

	grpcServer := grpc.NewServer()
	greetv1.RegisterGreetServiceServer(grpcServer, &server{})

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-gCtx.Done()
		grpcServer.GracefulStop()
		return nil
	})

	g.Go(func() error {
		log.Println("serving gRPC server at localhost:50051...")
		return grpcServer.Serve(lis)
	})

	return g.Wait()
}
