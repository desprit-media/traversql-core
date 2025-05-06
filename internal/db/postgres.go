package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConfig holds all the database connection parameters
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// checkPostgresEnvs verifies that all necessary PostgreSQL environment variables are set.
func checkPostgresEnvs() error {
	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		return fmt.Errorf("POSTGRES_HOST environment variable is not set")
	}
	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		return fmt.Errorf("POSTGRES_USER environment variable is not set")
	}
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		return fmt.Errorf("POSTGRES_PASS environment variable is not set")
	}
	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		return fmt.Errorf("POSTGRES_DB environment variable is not set")
	}
	postgresPort := os.Getenv("POSTGRES_PORT")
	if postgresPort == "" {
		return fmt.Errorf("POSTGRES_PORT environment variable is not set")
	}

	return nil
}

// NewPostgresConfigFromEnvs retrieves database configuration from environment variables
// and returns a PostgresConfig struct. It also checks if all required environment variables
// are set before returning the configuration.
func NewPostgresConfigFromEnvs() (PostgresConfig, error) {
	config := PostgresConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
	}

	if err := checkPostgresEnvs(); err != nil {
		return PostgresConfig{}, err
	}

	return config, nil
}

// InitPostgresPool creates and initializes a PostgreSQL connection pool using the provided configuration.
// ctx: The context for the pool initialization and ping check.
// config: The PostgresConfig containing the database connection parameters.
func InitPostgresPool(ctx context.Context, config PostgresConfig) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.Database)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return pool, nil
}
