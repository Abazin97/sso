package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/otp"
	"sso/internal/repository/redis"
	"sso/internal/repository/sqlite"
	"sso/internal/services"
	"sso/internal/services/auth"
	"sso/internal/services/email/smtp"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	config *config.Config,
	grpcPort int,
	smtpPort int,
	from string,
	pass string,
	host string,
	storagePath string,
	tokenTTL time.Duration,
	verificationCodeLength int,
	redisHost config.RedisConfig,
	verificationCodeTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		log.Error("sqlite unavailable")
	}
	smtpService, err := smtp.NewSMTPService(from, pass, host, smtpPort)
	if err != nil {
		log.Error("smtp unavailable")
	}

	emails, err := services.NewEmailService(log, smtpService, config.Email)
	if err != nil {
		log.Error("emails unavailable")
	}

	otpGenerator := otp.NewGOTPGenerator()

	redisRepo, err := redis.New(redisHost)
	if err != nil {
		log.Error("redis unavailable")
	}

	authService := auth.New(
		log,
		storage,
		storage,
		storage,
		tokenTTL,
		emails,
		otpGenerator,
		verificationCodeLength,
		redisRepo,
		verificationCodeTTL)

	grpcApp := grpcapp.New(log, grpcPort, authService)

	return &App{
		GRPCSrv: grpcApp,
	}
}
