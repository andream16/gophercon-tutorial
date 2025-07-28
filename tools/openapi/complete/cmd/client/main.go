package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	openapi "github.com/andream16/gophercon-tutorial/tools/openapi/complete/gopher/server/gen"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client, err := openapi.NewClientWithResponses("yo", openapi.WithBaseURL("http://localhost:8080"))
	if err != nil {
		log.Fatalf("could not create client: %v", err)
	}

	resp, err := client.ListGophers(ctx)
	switch {
	case err != nil:
		log.Fatalf("could not list gophers: %v", err)
	case resp.StatusCode != http.StatusOK:
		log.Fatalf("could not list gophers, got: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var gophers []openapi.Gopher
	if err := json.NewDecoder(resp.Body).Decode(&gophers); err != nil {
		log.Fatalf("could not decode gophers: %v", err)
	}

	log.Printf("gophers: %+v", gophers)

	resp, err = client.BuyGopher(
		ctx,
		1,
		openapi.BuyGopherJSONRequestBody{
			CustomerEmail: "andrea@gitpod.io",
			Quantity:      10,
		},
	)
	switch {
	case err != nil:
		log.Fatalf("could not buy gopher: %v", err)
	case resp.StatusCode != http.StatusOK:
		log.Fatalf("could not list gophers, got: %d", resp.StatusCode)
	}

}
