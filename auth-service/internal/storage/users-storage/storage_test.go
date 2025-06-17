package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(ctx context.Context) (testcontainers.Container, *PGStorage, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}
	testStorage := &PGStorage{
		pool: setupTestPool(ctx, host, port.Port()),
	}

	if err := initTestSchema(ctx, testStorage.pool); err != nil {
		return nil, nil, err
	}

	return pgContainer, testStorage, nil
}

func setupTestPool(ctx context.Context, host, port string) *pgxpool.Pool {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=testuser password=testpass dbname=testdb",
		host, port,
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic(err)
	}

	return pool
}

func initTestSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			email VARCHAR(50) NOT NULL UNIQUE,
			first_name VARCHAR(50) NOT NULL,
			last_name VARCHAR(50) NOT NULL,
			password_hash BYTEA NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS idx_email ON users(email);

		CREATE TABLE IF NOT EXISTS apps
		(
			id     SERIAL PRIMARY KEY,
			name   TEXT NOT NULL UNIQUE,
			secret TEXT NOT NULL UNIQUE
		);
	`)
	return err
}

func TestStorage(t *testing.T) {
	ctx := context.Background()

	pgContainer, testStorage, err := setupTestDB(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	t.Run("Save and Get User", func(t *testing.T) {
		email := "test@example.com"
		firstName := "John"
		lastName := "Doe"
		passHash := []byte("hashedpassword")

		usrid, err := testStorage.SaveUser(ctx, email, lastName, firstName, passHash)
		require.NoError(t, err)

		user, err := testStorage.GetUser(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, user)

		require.Equal(t, int64(1), usrid)
		require.Equal(t, email, user.Email)
		require.Equal(t, firstName, user.FirstName)
		require.Equal(t, lastName, user.LastName)
		require.Equal(t, passHash, user.PassHash)
	})

	t.Run("User Not Found", func(t *testing.T) {
		_, err := testStorage.GetUser(ctx, "nonexistent@example.com")
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("Save Duplicate User", func(t *testing.T) {
		email := "duplicate@example.com"
		usrid, err := testStorage.SaveUser(ctx, email, "Last", "First", []byte("hash"))
		require.NoError(t, err)
		require.Equal(t, int64(2), usrid)

		usrid, err = testStorage.SaveUser(ctx, email, "Last", "First", []byte("hash"))
		require.ErrorIs(t, err, ErrUserExists)
		require.Equal(t, int64(0), usrid)
	})

	t.Run("Save and Get App", func(t *testing.T) {
		appName := "testapp"
		appSecret := "secret123"

		err := testStorage.SaveApp(ctx, appName, appSecret)
		require.NoError(t, err)

		app, err := testStorage.GetApp(ctx, 1)
		require.NoError(t, err)
		require.NotNil(t, app)

		require.Equal(t, 1, app.ID)
		require.Equal(t, appName, app.Name)
		require.Equal(t, appSecret, app.Secret)
	})

	t.Run("App Not Found", func(t *testing.T) {
		_, err := testStorage.GetApp(ctx, 999)
		require.ErrorIs(t, err, ErrAppNotFound)
	})
}
