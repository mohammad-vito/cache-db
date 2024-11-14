package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sn/internal/pkg/cache"
	"sn/internal/server"
	"sn/internal/server/handlers"
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
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	log.Infow("starting service", "version", "1")
	defer log.Infow("shutdown complete")

	cacheCore := cache.NewCore(log, 10)

	cfg := handlers.Config{
		Log:        log,
		Cache:      cacheCore,
		TimeoutSec: 10,
	}

	r := server.APIMux(&cfg)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetByKey("API_PORT")),
		Handler: r,
	}

	go func() {
		log.Infow("startup", "status", "api router started", "host", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}

	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	select {
	case <-ctx.Done():
		log.Warn("shutdown timeout failed:")
	}
	log.Info("Server exiting")
	return nil
}
