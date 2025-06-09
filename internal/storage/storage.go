package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	SaveUser(ctx context.Context, email, last_name, first_name string, passHash []byte) error
	SaveApp(ctx context.Context, name, secret string) error
}

type PGStorage struct {
	pool *pgxpool.Pool
}

type PGConnectionInfo struct {
	Name     string
	User     string
	Password string
	Host     string
	Port     string
}

func initEnv() (*PGConnectionInfo, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	dbName, exists := os.LookupEnv("POSTGRES_NAME")
	if !exists {
		return nil, fmt.Errorf("not found dbName in env file")
	}
	dbUsername, exists := os.LookupEnv("POSTGRES_USER")
	if !exists {
		return nil, fmt.Errorf("not found dbUser in env file")
	}
	dbPassword, exists := os.LookupEnv("POSTGRES_PASSWORD")
	if !exists {
		return nil, fmt.Errorf("not found dbPassword in env file")
	}
	dbPort, exists := os.LookupEnv("POSTGRES_PORT")
	if !exists {
		return nil, fmt.Errorf("not found dbPort in env file")
	}
	dbHost, exists := os.LookupEnv("POSTGRES_HOST")
	if !exists {
		return nil, fmt.Errorf("not found dbHost in env file")
	}

	return &PGConnectionInfo{
		Name:     dbName,
		User:     dbUsername,
		Password: dbPassword,
		Port:     dbPort,
		Host:     dbHost,
	}, nil
}

func NewPGHandler(ctx context.Context) (*PGStorage, error) {
	config, err := initEnv()

	if err != nil {
		return nil, err
	}

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

func (st *PGStorage) SaveUser(ctx context.Context, email, last_name, first_name string, passHash []byte) error {
	tx, err := st.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	querySaveUser := `
		INSERT INTO users(email, last_name, first_name, password_hash)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, querySaveUser, email, last_name, first_name, passHash)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("%w", ErrUserExists)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
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
