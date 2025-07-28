package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"

	openapi "github.com/andream16/gophercon-tutorial/tools/openapi/complete/gopher/server/gen"
)

const listenAddr = "0.0.0.0:8080"

type server struct {
	gophers []openapi.Gopher
}

func newServer() *server {
	return &server{
		gophers: []openapi.Gopher{
			{
				Color: openapi.GopherColorBlack,
				Name:  "Jess",
				Price: 10.32,
				Id:    1,
			},
			{
				Color: openapi.GopherColorBlack,
				Name:  "Yuno",
				Price: 60,
				Id:    2,
			},
		},
	}
}

func (s *server) ListGophers(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(s.gophers)
}

func (s *server) BuyGopher(w http.ResponseWriter, r *http.Request, gopherId int64) {
	var gopher openapi.Gopher
	if err := json.NewDecoder(r.Body).Decode(&gopher); err != nil {
		_ = json.NewEncoder(w).Encode(openapi.UnmarshalingParamError{
			ParamName: "gopher",
			Err:       err,
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if gopherId <= 0 {
		_ = json.NewEncoder(w).Encode(openapi.InvalidParamFormatError{
			ParamName: "gopher_id",
			Err:       errors.New("must be a positive integer"),
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(openapi.PurchaseResponse{
		OrderId: strconv.FormatInt(gopher.Id, 10),
	})

	w.WriteHeader(http.StatusOK)
}

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

	var (
		router  = mux.NewRouter()
		srv     = newServer()
		handler = openapi.HandlerFromMux(srv, router)
	)

	httpSrv := &http.Server{
		Handler: handler,
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
