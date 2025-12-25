package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"syscall"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	//fmt.Println("from yaml:", cfg.SMTP.Pass)

	log := setupLogger(cfg.Env)

	log.Info("starting app", slog.Any("config", cfg))

	application := app.New(log, cfg, cfg.GRPC.Port, cfg.SMTP.Port, cfg.SMTP.From, cfg.SMTP.Pass, cfg.SMTP.Host, cfg.StoragePath, cfg.TokenTTL, cfg.SMTP.VerificationCodeLength, cfg.Redis, cfg.Redis.VerTokenTTL)

	go application.GRPCSrv.MustRun()

	//Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))

	application.GRPCSrv.Stop()
	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
