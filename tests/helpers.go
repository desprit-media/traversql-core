package tests

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewPostgresContainer(ctx context.Context, t *testing.T, initFilePath ...string) (testcontainers.Container, *pgxpool.Pool) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithDatabase("test"),
		postgres.WithInitScripts(initFilePath...),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Second*5)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	connectionString, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}
	if connectionString[len(connectionString)-1] == '?' {
		connectionString += "sslmode=disable"
	} else {
		connectionString += "?sslmode=disable"
	}

	pgPool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		t.Fatalf("Failed to create pgx pool: %v", err)
	}
	t.Cleanup(func() {
		pgPool.Close()
	})

	return pgContainer, pgPool
}

func ExecuteOnTmpSchema(ctx context.Context, pgPool *pgxpool.Pool, tablesStmp string, sql string) error {
	_, err := pgPool.Exec(ctx, "CREATE SCHEMA tmp_schema")
	if err != nil {
		return err
	}
	defer func() {
		pgPool.Exec(ctx, "DROP SCHEMA tmp_schema CASCADE;")
		pgPool.Exec(ctx, "SET search_path TO public;")
	}()

	createTablesStmt := "SET search_path TO tmp_schema;" + string(tablesStmp)
	_, err = pgPool.Exec(ctx, createTablesStmt)
	if err != nil {
		return err
	}

	insertRecordsStmt := "SET search_path TO tmp_schema;\n" + sql
	_, err = pgPool.Exec(ctx, insertRecordsStmt)
	return err
}
