package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	greetv1 "github.com/andream16/gophercon-tutorial/tools/grpc/proto/greet/v1"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := Main(ctx); err != nil {
		log.Fatal(err)
	}
}

func Main(ctx context.Context) error {
	conn, err := grpc.NewClient(
		"0.0.0.0:50051",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		return fmt.Errorf("could not create gRPC client: %w", err)
	}

	greetClient := greetv1.NewGreetServiceClient(conn)

	resp, err := greetClient.Greet(ctx, &greetv1.GreetRequest{
		Name: "John Doe",
	})
	if err != nil {
		return fmt.Errorf("could not greet: %w", err)
	}

	log.Printf("greet response: %s", resp.Greeting)

	return nil
}
