package repository

import (
	"context"
	"errors"
	"sso/internal/domain/models"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrCodeNotFound = errors.New("code not found")
	ErrAppNotFound  = errors.New("app not found")
)

type Redis interface {
	SaveCode(ctx context.Context, code string, uid int64) error
	Code(ctx context.Context, uid int64) (models.Code, error)
}
