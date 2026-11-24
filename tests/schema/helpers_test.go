package schema_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"demoRBACProject1/internal/config"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type EnvironmentSnapshot struct {
	Database         string
	Version          string
	SQLMode          string
	TimeZone         string
	ForeignKeyChecks string
	CharacterSet     string
	Collation        string
	InnoDBSupport    string
}

type MigrationTrack struct {
	Name         string
	SourcePath   string
	HistoryTable string
	Target       uint
}

type ObjectSnapshot struct {
	Tables   []string
	Routines []string
	Rows     map[string]int64
	Version  uint
	Dirty    bool
}

var (
	schemaTrack = MigrationTrack{
		Name:         "schema",
		SourcePath:   "migrations/schema",
		HistoryTable: "schema_migrations_schema",
		Target:       1,
	}
	credentialsTrack = MigrationTrack{
		Name:         "credentials",
		SourcePath:   "migrations/credentials",
		HistoryTable: "schema_migrations_credentials",
		Target:       1,
	}
	finalizeCredentialsTrack = MigrationTrack{
		Name:         "credentials-finalize",
		SourcePath:   "migrations/credentials-finalize",
		HistoryTable: "schema_migrations_credentials_finalize",
		Target:       1,
	}
	legacyFixtureTrack = MigrationTrack{
		Name:         "fixture-legacy",
		SourcePath:   "tests/schema/testdata/legacy",
		HistoryTable: "schema_migrations_fixture_legacy",
		Target:       1,
	}
	sentinelFixtureTrack = MigrationTrack{
		Name:         "fixture-sentinel",
		SourcePath:   "tests/schema/testdata/sentinel",
		HistoryTable: "schema_migrations_fixture_sentinel",
		Target:       1,
	}
)

func openReadOnly(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	return openConfiguredDB(ctx, cfg)
}

func openConfiguredDB(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}

	dsn, err := configuredDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open configured connection %s: %w", config.RedactedDSN(cfg), err)
	}
	if err := applySessionSettings(ctx, db); err != nil {
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
	@@session.foreign_key_checks,
	@@character_set_connection,
	@@collation_connection
`).Scan(
		&snapshot.Database,
		&snapshot.Version,
		&snapshot.SQLMode,
		&snapshot.TimeZone,
		&snapshot.ForeignKeyChecks,
		&snapshot.CharacterSet,
		&snapshot.Collation,
	)
	if err != nil {
		return snapshot, err
	}

	err = db.QueryRowContext(ctx, `
SELECT SUPPORT
FROM INFORMATION_SCHEMA.ENGINES
WHERE ENGINE = 'InnoDB'
`).Scan(&snapshot.InnoDBSupport)
	return snapshot, err
}

func validateEnvironment(cfg config.Config, snapshot EnvironmentSnapshot) error {
	if err := config.Validate(cfg); err != nil {
		return err
	}

	var problems []string
	if snapshot.Database != config.RequiredDatabase {
		problems = append(problems, fmt.Sprintf("database must be %q, got %q", config.RequiredDatabase, snapshot.Database))
	}
	if !strings.HasPrefix(snapshot.Version, "8.4.") {
		problems = append(problems, "mysql version must be 8.4.x, got "+snapshot.Version)
	}
	for _, required := range config.RequiredSQLModes {
		if !csvContains(snapshot.SQLMode, required) {
			problems = append(problems, "sql_mode missing "+required)
		}
	}
	if snapshot.TimeZone != config.RequiredTimeZone && !strings.EqualFold(snapshot.TimeZone, "UTC") {
		problems = append(problems, "time zone must be UTC/+00:00, got "+snapshot.TimeZone)
	}
	if snapshot.ForeignKeyChecks != "1" {
		problems = append(problems, "foreign_key_checks must be 1")
	}
	if snapshot.InnoDBSupport != "" && snapshot.InnoDBSupport != "YES" && snapshot.InnoDBSupport != "DEFAULT" {
		problems = append(problems, "InnoDB must be supported, got "+snapshot.InnoDBSupport)
	}
	if snapshot.CharacterSet != "" && snapshot.CharacterSet != config.RequiredCharset {
		problems = append(problems, fmt.Sprintf("connection charset must be %q, got %q", config.RequiredCharset, snapshot.CharacterSet))
	}
	if snapshot.Collation != "" && snapshot.Collation != config.RequiredCollation {
		problems = append(problems, fmt.Sprintf("connection collation must be %q, got %q", config.RequiredCollation, snapshot.Collation))
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return errors.New("unsupported schema test environment: " + strings.Join(problems, "; "))
	}
	return nil
}

func newMigrator(ctx context.Context, cfg config.Config, track MigrationTrack) (*migrate.Migrate, *sql.DB, error) {
	db, err := openConfiguredDB(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	driver, err := mysqlmigrate.WithInstance(db, &mysqlmigrate.Config{
		MigrationsTable: track.HistoryTable,
		DatabaseName:    cfg.Database,
	})
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	sourcePath, err := migrationSourcePath(track.SourcePath)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://"+sourcePath, cfg.Database, driver)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return migrator, db, nil
}

func migrationSourcePath(sourcePath string) (string, error) {
	if filepath.IsAbs(sourcePath) {
		return sourcePath, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("inspect working directory for migration source %s: %w", sourcePath, err)
	}
	for {
		candidate := filepath.Join(wd, sourcePath)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", fmt.Errorf("migration source %s not found from working directory", sourcePath)
}

func migrateToVersion(ctx context.Context, cfg config.Config, track MigrationTrack, version uint) error {
	migrator, db, err := newMigrator(ctx, cfg, track)
	if err != nil {
		return err
	}
	defer db.Close()
	defer migrator.Close()

	if err := migrator.Migrate(version); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func migrationVersion(ctx context.Context, cfg config.Config, track MigrationTrack) (uint, bool, error) {
	migrator, db, err := newMigrator(ctx, cfg, track)
	if err != nil {
		return 0, false, err
	}
	defer db.Close()
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}

func cleanupOwnedObjects(ctx context.Context, db *sql.DB) error {
	statements := []string{
		"DROP TABLE IF EXISTS orders",
		"DROP TABLE IF EXISTS usercart",
		"DROP TABLE IF EXISTS user_demand",
		"DROP TABLE IF EXISTS transactions",
		"DROP TABLE IF EXISTS product",
		"DROP TABLE IF EXISTS `user`",
		"DROP TABLE IF EXISTS sentinel_owned",
		"DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials",
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("cleanup owned object with %q: %w", statement, err)
		}
	}
	return nil
}

func snapshotOwnedObjects(ctx context.Context, db *sql.DB, cfg config.Config, track MigrationTrack) (ObjectSnapshot, error) {
	rows, err := db.QueryContext(ctx, `
SELECT TABLE_NAME
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('product','user','orders','usercart','user_demand','transactions','sentinel_owned')
ORDER BY TABLE_NAME`)
	if err != nil {
		return ObjectSnapshot{}, err
	}
	defer rows.Close()

	snapshot := ObjectSnapshot{Rows: map[string]int64{}}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return ObjectSnapshot{}, err
		}
		snapshot.Tables = append(snapshot.Tables, name)
	}
	if err := rows.Err(); err != nil {
		return ObjectSnapshot{}, err
	}
	if err := rows.Close(); err != nil {
		return ObjectSnapshot{}, err
	}

	for _, name := range snapshot.Tables {
		var count int64
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `"+strings.ReplaceAll(name, "`", "``")+"`").Scan(&count); err != nil {
			return ObjectSnapshot{}, err
		}
		snapshot.Rows[name] = count
	}

	routineRows, err := db.QueryContext(ctx, `
SELECT ROUTINE_NAME
FROM INFORMATION_SCHEMA.ROUTINES
WHERE ROUTINE_SCHEMA = DATABASE()
  AND ROUTINE_NAME = 'schema_guard_finalize_credentials'
ORDER BY ROUTINE_NAME`)
	if err != nil {
		return ObjectSnapshot{}, err
	}
	defer routineRows.Close()
	for routineRows.Next() {
		var name string
		if err := routineRows.Scan(&name); err != nil {
			return ObjectSnapshot{}, err
		}
		snapshot.Routines = append(snapshot.Routines, name)
	}
	if err := routineRows.Err(); err != nil {
		return ObjectSnapshot{}, err
	}

	version, dirty, err := migrationVersion(ctx, cfg, track)
	if err != nil {
		return ObjectSnapshot{}, err
	}
	snapshot.Version = version
	snapshot.Dirty = dirty
	return snapshot, nil
}

func snapshotBusinessAndHistory(ctx context.Context, db *sql.DB, cfg config.Config) (ObjectSnapshot, error) {
	rows, err := db.QueryContext(ctx, `
SELECT TABLE_NAME
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN (
	'product','user','orders','usercart','user_demand','transactions',
	'schema_migrations_schema','schema_migrations_credentials',
	'schema_migrations_credentials_finalize','schema_migrations_fixture_legacy',
	'schema_migrations_fixture_sentinel'
  )
ORDER BY TABLE_NAME`)
	if err != nil {
		return ObjectSnapshot{}, err
	}
	defer rows.Close()

	snapshot := ObjectSnapshot{Rows: map[string]int64{}}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return ObjectSnapshot{}, err
		}
		snapshot.Tables = append(snapshot.Tables, name)
	}
	if err := rows.Err(); err != nil {
		return ObjectSnapshot{}, err
	}
	if err := rows.Close(); err != nil {
		return ObjectSnapshot{}, err
	}

	for _, name := range snapshot.Tables {
		var count int64
		query := "SELECT COUNT(*) FROM `" + strings.ReplaceAll(name, "`", "``") + "`"
		if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			return ObjectSnapshot{}, err
		}
		snapshot.Rows[name] = count
	}
	return snapshot, nil
}

func assertMySQLError(t testingT, err error, number uint16, state string) {
	t.Helper()
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		t.Fatalf("expected MySQL error %d/%s, got %T: %v", number, state, err, err)
	}
	gotState := string(mysqlErr.SQLState[:])
	if mysqlErr.Number != number || gotState != state {
		t.Fatalf("mysql error = %d/%s, want %d/%s", mysqlErr.Number, gotState, number, state)
	}
}

func checkOccupiedBootstrap(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `
SELECT TABLE_NAME
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_TYPE = 'BASE TABLE'
  AND TABLE_NAME IN ('product','user','orders','usercart','user_demand','transactions')
ORDER BY TABLE_NAME`)
	if err != nil {
		return fmt.Errorf("inspect bootstrap inventory: %w", err)
	}
	defer rows.Close()

	var occupied []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan bootstrap inventory: %w", err)
		}
		occupied = append(occupied, name)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read bootstrap inventory: %w", err)
	}
	if len(occupied) > 0 {
		return fmt.Errorf("occupied bootstrap inventory: %s", strings.Join(occupied, ", "))
	}
	return nil
}

func checkCredentialEntry(ctx context.Context, db *sql.DB) error {
	var tables int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_TYPE = 'BASE TABLE'
  AND TABLE_NAME = 'user'`).Scan(&tables); err != nil {
		return fmt.Errorf("inspect credential entry inventory: %w", err)
	}
	if tables == 0 {
		return errors.New("credential entry refused: legacy user table is absent")
	}

	rows, err := db.QueryContext(ctx, `
SELECT COLUMN_NAME
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'user'`)
	if err != nil {
		return fmt.Errorf("inspect credential entry columns: %w", err)
	}
	defer rows.Close()

	var hasLegacyPassword bool
	var targets []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan credential entry columns: %w", err)
		}
		switch strings.ToLower(name) {
		case "password":
			hasLegacyPassword = true
		case "password_hash", "must_reset_password":
			targets = append(targets, strings.ToLower(name))
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read credential entry columns: %w", err)
	}

	if len(targets) > 0 {
		sort.Strings(targets)
		return fmt.Errorf("credential entry refused: user table already carries target credential columns: %s", strings.Join(targets, ", "))
	}
	if !hasLegacyPassword {
		return errors.New("credential entry refused: user table missing legacy password column")
	}
	return nil
}

func configuredDSN(cfg config.Config) (string, error) {
	parsed, err := mysql.ParseDSN(cfg.DSN)
	if err != nil {
		return "", err
	}
	parsed.DBName = config.RequiredDatabase
	parsed.ParseTime = true
	parsed.Collation = ""
	parsed.Loc = time.UTC
	parsed.MultiStatements = true
	if parsed.Params == nil {
		parsed.Params = map[string]string{}
	}
	parsed.Params["charset"] = config.RequiredCharset
	parsed.Params["time_zone"] = "'+00:00'"
	parsed.Params["foreign_key_checks"] = "1"
	parsed.Params["sql_mode"] = "'STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO'"
	parsed.ReadTimeout = cfg.Runtime.ReadTimeout
	parsed.WriteTimeout = cfg.Runtime.WriteTimeout
	return parsed.FormatDSN(), nil
}

func applySessionSettings(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
SET SESSION
	sql_mode = 'STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO',
	time_zone = '+00:00',
	foreign_key_checks = 1`)
	if err != nil {
		return fmt.Errorf("apply required session settings: %w", err)
	}
	if _, err := db.ExecContext(ctx, "SET NAMES utf8mb4 COLLATE utf8mb4_0900_bin"); err != nil {
		return fmt.Errorf("apply required connection charset: %w", err)
	}
	return nil
}

func csvContains(csv string, want string) bool {
	for _, value := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(value), want) {
			return true
		}
	}
	return false
}

type testingT interface {
	Helper()
	Fatalf(string, ...any)
}
