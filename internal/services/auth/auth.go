package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/otp"
	"sso/internal/repository"
	"sso/internal/services"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log                    *slog.Logger
	usrSaver               UserSaver
	usrProvider            UserProvider
	appProvider            AppProvider
	tokenTTL               time.Duration
	emailService           *services.EmailService
	otpGenerator           otp.Generator
	verificationCodeLength int
	repo                   repository.Redis
	verCodeTTL             time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		title string,
		birthDate string,
		name string,
		lastName string,
		email string,
		passHash []byte,
		phone string,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string, phone string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	SetPassword(ctx context.Context, email string, newPassword []byte) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	//ErrNotValidCode       = errors.New("invalid code")
)

// New returns a new instance of the Auth service.
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
	emailService *services.EmailService,
	otpGenerator otp.Generator,
	verificationCodeLength int,
	repo repository.Redis,
	verCodeTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:               userSaver,
		usrProvider:            userProvider,
		log:                    log,
		emailService:           emailService,
		appProvider:            appProvider,
		tokenTTL:               tokenTTL,
		otpGenerator:           otpGenerator,
		verificationCodeLength: verificationCodeLength,
		repo:                   repo,
		verCodeTTL:             verCodeTTL,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error. If user doesn’t exist, returns error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	phone string,
	appID int,
) (models.User, string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
		slog.String("phone", phone),
	)

	log.Info("logining user")

	user, err := a.usrProvider.User(ctx, email, phone)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return models.User{}, "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return models.User{}, "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))
		return models.User{}, "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, repository.ErrAppNotFound) {
			a.log.Warn("app not found", sl.Err(err))
			return models.User{}, "", fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return models.User{}, "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)

	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return models.User{}, "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in succesfully")

	return user, token, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context,
	title string,
	birthDate string,
	name string,
	lastName string,
	email string,
	pass string,
	phone string,
) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	// todo: add salt into password
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, title, birthDate, name, lastName, email, passHash, phone)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registered!")

	return id, nil
}

// IsAdmin checks if user is admin.
func (a *Auth) IsAdmin(ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("userID", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Warn("user not found")
		}
		return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
	}

	log.Info("checked if user is admin", slog.Bool("is admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) ChangePasswordInit(ctx context.Context, email string, phone string, oldPassword string) (string, int64, error) {
	const op = "auth.ChangePasswordInit"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("getting user...")

	user, err := a.usrProvider.User(ctx, email, phone)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", 0, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", sl.Err(err))

		return "", 0, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(oldPassword)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))
		return "", 0, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("got user, sending verification email")

	verificationCode := a.otpGenerator.RandomSecret(a.verificationCodeLength)

	uid := user.ID
	//a.log.Info("uid:", slog.Int64("uid", uid))
	err = a.repo.SaveCode(ctx, verificationCode, uid)
	if err != nil {
		a.log.Error("failed to save verification code", sl.Err(err))
	}

	log.Info("code saved in redis")

	redisTTL := a.verCodeTTL
	expiresAt := time.Now().UTC().Add(redisTTL).Format(time.RFC3339)
	//log.Info("expires at:", expiresAt)

	err = a.emailService.SendVerificationEmail(services.VerificationEmailInput{
		Email:            email,
		Name:             user.Name,
		VerificationCode: verificationCode,
	})
	if err != nil {
		a.log.Info("failed to send verification email", sl.Err(err))
	}

	log.Info("email has sent")

	return expiresAt, uid, nil
}

func (a *Auth) ChangePasswordConfirm(ctx context.Context, verificationCode string, uid int64, email string, newPassword string) (bool, error) {
	const op = "auth.ChangePasswordConfirm"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("comparing verification code")

	code, err := a.repo.Code(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrCodeNotFound) {
			a.log.Warn("invalid verification code", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get code", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	if code.Code != verificationCode {
		a.log.Warn("codes doesn't match")
		return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("correct code!")

	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Info("failed to generate password hash", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	success, err := a.usrProvider.SetPassword(ctx, email, passHash)
	if err != nil {
		a.log.Info("failed to change password", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("password changed successfully")

	return success, nil
}
