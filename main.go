package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zenpk/chatbone/repo"
	"github.com/zenpk/chatbone/util"
)

var mode = flag.String("mode", "dev", "define program mode")

func main() {
	flag.Parse()
	// graceful exit
	var cleanUpErr error
	defer func() {
		if cleanUpErr != nil {
			panic(cleanUpErr)
		}
		log.Println("gracefully exited")
	}()

	conf := new(util.Configuration)
	if err := conf.Init(*mode); err != nil {
		panic(err)
	}

	logger := new(util.Logger)
	if err := logger.Init(conf); err != nil {
		panic(err)
	}
	defer func() {
		log.Println("closing logger")
		if err := logger.Close(); err != nil {
			cleanUpErr = errors.Join(cleanUpErr, err)
		}
	}()

	mongoDb := new(repo.Mongo)
	if err := mongoDb.Init(conf); err != nil {
		panic(err)
	}

	// clean up
	osSignalChan := make(chan os.Signal, 2)
	signal.Notify(osSignalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-osSignalChan
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := hd.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	log.Println("started")
	if err := hd.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
