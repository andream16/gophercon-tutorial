package main

import (
	"context"
	"log"
	"time"
)

type server struct{}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := Main(ctx); err != nil {
		log.Fatal(err)
	}
}

func Main(ctx context.Context) error {
	return nil
}
