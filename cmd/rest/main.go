package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sn/internal/business/cache"
	"sn/internal/server"
	"sn/internal/server/handler"
	"sn/pkg/config"
	"sn/pkg/logger"
	"syscall"
	"time"
	//"sn/internal/pkg/log"
)

func main() {
	log, err := logger.New("rest-api")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	log.Infow("starting service", "version", "1")
	defer log.Infow("shutdown complete")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cacheCore := cache.NewCore(log, 10)

	cfg := handler.Config{
		Shutdown:   shutdown,
		Log:        log,
		Cache:      cacheCore,
		TimeoutSec: 10,
	}

	r := server.APIMux(&cfg)

	api := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetByKey("API_PORT")),
		Handler: r,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}
	return nil
}
