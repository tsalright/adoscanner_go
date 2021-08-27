// Package main starts the server
package main

import (
	"adoscanner/ado"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	srv := ado.InitializeServer()

	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			logger := ado.AppInsightsLogger{}
			logger.LogFatal(err)
			log.Fatal(err)
		}
	}()

	waitForShutdown(srv)

	os.Exit(0)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		logger := ado.AppInsightsLogger{}
		logger.LogFatal(err)
		log.Panic(err)
	}
	log.Println("Shutting down")
}