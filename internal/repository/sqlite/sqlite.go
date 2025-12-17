package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/repository"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func New(storagePath string) (*Repository, error) {
	const op = "repository.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Repository{db: db}, nil
}

func (s *Repository) Stop() error {
	return s.db.Close()
}

// SaveUser saves user to db.
func (s *Repository) SaveUser(ctx context.Context, title string, birthDate string, name string, lastName string, email string, passHash []byte, phone string) (int64, error) {
	const op = "repository.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(title, birth_date, name, last_name, email, pass_hash, phone) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, title, birthDate, name, lastName, email, passHash, phone)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, repository.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Repository) User(ctx context.Context, email string, phone string) (models.User, error) {
	const op = "repository.sqlite.User"

	stmt, err := s.db.Prepare("SELECT id, title, birth_date, name, last_name, email, pass_hash, phone FROM users WHERE email = ? OR phone = ? LIMIT 1")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email, phone)

	var user models.User
	err = row.Scan(&user.ID, &user.Title, &user.BirthDate, &user.Name, &user.LastName, &user.Email, &user.PassHash, &user.Phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (s *Repository) SetPassword(ctx context.Context, email string, newPassword []byte) (bool, error) {
	const op = "repository.sqlite.ChangePassword"

	stmt, err := s.db.Prepare("UPDATE users SET pass_hash = ? WHERE email = ?")

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	//defer stmt.Close()

	_, err = stmt.ExecContext(ctx, newPassword, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Repository) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "repository.sqlite.IsAdmin"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userID)
	var isAdmin bool
	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Repository) App(ctx context.Context, id int) (models.App, error) {
	const op = "repository.sqlite.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, id)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}
