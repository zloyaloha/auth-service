package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zloyaloha/auth-service/internal/config"
	"github.com/zloyaloha/auth-service/internal/domain/models"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

type Storage interface {
	GetUser(ctx context.Context, email string) (*models.User, error)
	GetApp(ctx context.Context, id int) (*models.App, error)
	SaveUser(ctx context.Context, email, last_name, first_name string, passHash []byte) (int64, error)
	SaveApp(ctx context.Context, name, secret string) error
}

type PGStorage struct {
	pool *pgxpool.Pool
}

func NewPGHandler(ctx context.Context, config config.PGConnectionConfig) (*PGStorage, error) {
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Name,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error while parsing connection string: %w", err)
	}

	poolConfig.MaxConns = 5
	poolConfig.ConnConfig.ConnectTimeout = 60 * time.Second
	poolConfig.HealthCheckPeriod = 60 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error while creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error while pinging database: %w", err)
	}

	return &PGStorage{pool: pool}, nil
}

func (st *PGStorage) SaveUser(ctx context.Context, email, last_name, first_name string, passHash []byte) (int64, error) {
	tx, err := st.pool.Begin(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	querySaveUser := `
		INSERT INTO users(email, last_name, first_name, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var userID int64
	err = tx.QueryRow(ctx, querySaveUser, email, last_name, first_name, passHash).Scan(&userID)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%w", ErrUserExists)
		}
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return userID, nil
}

func (st *PGStorage) SaveApp(ctx context.Context, name, secret string) error {
	tx, err := st.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	querySaveApp := `
		INSERT INTO apps (name, secret)
		VALUES ($1, $2)
	`
	_, err = tx.Exec(ctx, querySaveApp, name, secret)

	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (st *PGStorage) GetUser(ctx context.Context, email string) (*models.User, error) {
	tx, err := st.pool.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queryGetUser := `
		SELECT ID, email, first_name, last_name, password_hash
		FROM users
		WHERE email = $1
	`

	user := &models.User{}

	err = tx.QueryRow(ctx, queryGetUser, email).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PassHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrUserNotFound)
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction")
	}
	return user, nil
}

func (st *PGStorage) GetApp(ctx context.Context, id int) (*models.App, error) {
	tx, err := st.pool.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queryGetApp := `
		SELECT id, name, secret
		FROM apps
		WHERE id = $1
	`

	app := &models.App{}

	err = tx.QueryRow(ctx, queryGetApp, id).Scan(&app.ID, &app.Name, &app.Secret)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrAppNotFound)
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction")
	}
	return app, nil
}
