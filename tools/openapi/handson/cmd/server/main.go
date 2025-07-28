package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
)

const listenAddr = "0.0.0.0:8080"

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGKILL,
	)
	defer cancel()

	var router = mux.NewRouter()

	httpSrv := &http.Server{
		Handler: router,
		Addr:    listenAddr,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		return httpSrv.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		log.Printf("serving on: %s..\n", listenAddr)
		return httpSrv.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
