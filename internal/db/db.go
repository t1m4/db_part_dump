package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/t1m4/db_part_dump/config"

	_ "github.com/lib/pq"
)

func NewDB(ctx context.Context, c *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.GetDSN(c))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxIdleConns(c.Database.MaxIdleConns)
	db.SetMaxOpenConns(c.Database.MaxOpenConns)
	db.SetConnMaxLifetime(c.Database.ConnMaxLifetime)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
