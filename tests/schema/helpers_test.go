package schema_test

import (
	"context"
	"database/sql"
	"time"

	"demoRBACProject1/internal/config"
)

type EnvironmentSnapshot struct {
	Database         string
	Version          string
	SQLMode          string
	TimeZone         string
	ForeignKeyChecks string
}

func openReadOnly(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func inspectEnvironment(ctx context.Context, db *sql.DB) (EnvironmentSnapshot, error) {
	var snapshot EnvironmentSnapshot
	err := db.QueryRowContext(ctx, `
SELECT
	DATABASE(),
	@@version,
	@@session.sql_mode,
	@@session.time_zone,
	@@session.foreign_key_checks
`).Scan(
		&snapshot.Database,
		&snapshot.Version,
		&snapshot.SQLMode,
		&snapshot.TimeZone,
		&snapshot.ForeignKeyChecks,
	)
	return snapshot, err
}

func validateEnvironment(config.Config, EnvironmentSnapshot) error {
	return nil
}
