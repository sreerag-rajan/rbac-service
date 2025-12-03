package repository

import (
	"context"
	"fmt"
	"os"
	"sync"

	"rbac-service/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

func InitDB(ctx context.Context) error {
	var err error
	once.Do(func() {
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)

		config, cfgErr := pgxpool.ParseConfig(dsn)
		if cfgErr != nil {
			err = fmt.Errorf("unable to parse database config: %w", cfgErr)
			return
		}

		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}

		if pingErr := pool.Ping(ctx); pingErr != nil {
			err = fmt.Errorf("unable to ping database: %w", pingErr)
			return
		}

		logger.Info(ctx, "Database connection established", nil, "db", "init")
	})

	return err
}

func GetPool() *pgxpool.Pool {
	return pool
}

func CloseDB() {
	if pool != nil {
		pool.Close()
	}
}

func RunMigrations(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := os.ReadFile(dir + "/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		logger.Info(ctx, "Running migration", nil, "file", entry.Name())
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
