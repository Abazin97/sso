package redis

import (
	"context"
	"fmt"
	"sso/internal/config"
	"sso/internal/domain/models"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	db *redis.Client
}

func New(Redis config.RedisConfig) (*Repository, error) {
	const op = "repository.redis.New"

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", Redis.Host, Redis.Port),
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}

	return &Repository{
		db: rdb,
	}, nil
}

func (s *Repository) SaveCode(ctx context.Context, code string) (int64, error) {
	const op = "repository.redis.SaveCode"

	id, err := s.db.Incr(ctx, "code:id").Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %s", err, op)
	}

	key := fmt.Sprintf("code:%d", id)

	if err := s.db.Set(ctx, key, code, 0).Err(); err != nil {
		return 0, fmt.Errorf("%w: %s", err, op)
	}

	return id, nil
}

func (s *Repository) Code(ctx context.Context, id int64) (models.Code, error) {
	const op = "repository.redis.Code"

	key := fmt.Sprintf("code:%d", id)

	code, err := s.db.Get(ctx, key).Result()
	if err == redis.Nil {
		return models.Code{}, fmt.Errorf("%w: %s", err, op)
	}
	if err != nil {
		return models.Code{}, err
	}

	return models.Code{
		UserID: id,
		Code:   code,
	}, nil
}

// todo: code deletion after 5 mins of issuing
