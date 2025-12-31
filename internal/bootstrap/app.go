package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/logger/sl"
	"sso/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AppBootstrapRepository interface {
	CreateApp(ctx context.Context, name string, secret []byte) (int, error)
	UpdateApp(ctx context.Context, id int, name string, secret []byte) error
	App(ctx context.Context, id int) (models.App, error)
}

func InitApp(
	ctx context.Context,
	log *slog.Logger,
	repo AppBootstrapRepository,
	id int,
	name string,
	secret string,
) error {
	const op = "bootstrap.initApp"

	app, err := repo.App(ctx, id)
	if err == nil {
		nameChanged := app.Name != name

		secretChanged := bcrypt.CompareHashAndPassword(app.SecretHash, []byte(secret)) != nil
		if nameChanged || secretChanged {
			secretHash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
			if err != nil {
				log.Error("failed to generate app secret hash", sl.Err(err))

				return fmt.Errorf("%s: %w", op, err)
			}

			if err := repo.UpdateApp(ctx, id, name, secretHash); err != nil {
				log.Error("failed to update app secret hash", sl.Err(err))

				return fmt.Errorf("%s: %w", op, err)
			}

			log.Info("app secret updated", slog.String("name", name))
		}
		return nil
	}

	if errors.Is(err, repository.ErrAppNotFound) {
		secretHash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to generate password hash", sl.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}

		if _, err := repo.CreateApp(ctx, name, secretHash); err != nil {
			log.Error("failed to create app", sl.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}

		log.Info("app created", slog.String("name", name))
		return nil
	}

	log.Info("failed to get app", slog.String("name", name))

	return fmt.Errorf("%s: %w", op, err)
}
