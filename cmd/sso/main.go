package main

import (
	"app/internal/app"
	"app/internal/config"
	"app/pkg/common/logging"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	logger := logging.NewLogger(
		logging.WithLevel(cfg.AppConfig.LogLevel),
		logging.WithIsJSON(cfg.AppConfig.LogJSON),
	)
	logger.Info("starting logger")

	ctx = logging.ContextWithLogger(ctx, logger)
	application := app.New(ctx, cfg)

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	application.MustRun()

	sign := <-done

	logger.Info("stopped application", slog.String("signal", sign.String()))

	//application.Stop()

	logger.Info("app stopped")

}
