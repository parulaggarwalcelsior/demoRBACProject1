package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	EnvSchemaTestDSN = "SCHEMA_TEST_DSN"

	RequiredDatabase  = "shopping-cart"
	RequiredCharset   = "utf8mb4"
	RequiredCollation = "utf8mb4_0900_bin"
	RequiredTimeZone  = "+00:00"
)

var RequiredSQLModes = []string{
	"STRICT_ALL_TABLES",
	"NO_ENGINE_SUBSTITUTION",
	"NO_AUTO_VALUE_ON_ZERO",
}

type Config struct {
	DSN      string
	Database string
	Runtime  RuntimeSettings
}

type RuntimeSettings struct {
	Charset          string
	Collation        string
	Location         string
	SQLModes         []string
	ForeignKeyChecks bool
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
}

func Load() (Config, error) {
	cfg, err := FromDSN(os.Getenv(EnvSchemaTestDSN))
	if err != nil {
		return Config{}, err
	}
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func FromDSN(dsn string) (Config, error) {
	if strings.TrimSpace(dsn) == "" {
		return Config{
			DSN:     dsn,
			Runtime: RequiredRuntimeSettings(),
		}, nil
	}

	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		return Config{}, fmt.Errorf("parse schema test DSN: %w", err)
	}

	return Config{
		DSN:      dsn,
		Database: parsed.DBName,
		Runtime:  RequiredRuntimeSettings(),
	}, nil
}

func RequiredRuntimeSettings() RuntimeSettings {
	return RuntimeSettings{
		Charset:          RequiredCharset,
		Collation:        RequiredCollation,
		Location:         RequiredTimeZone,
		SQLModes:         append([]string(nil), RequiredSQLModes...),
		ForeignKeyChecks: true,
		ReadTimeout:      5 * time.Second,
		WriteTimeout:     5 * time.Second,
	}
}

func Validate(cfg Config) error {
	var problems []string
	if strings.TrimSpace(cfg.DSN) == "" {
		problems = append(problems, EnvSchemaTestDSN+" is required")
	}
	if cfg.Database != RequiredDatabase {
		problems = append(problems, fmt.Sprintf("database must be %q, got %q", RequiredDatabase, cfg.Database))
	}
	if cfg.Runtime.Charset != RequiredCharset {
		problems = append(problems, fmt.Sprintf("charset must be %q, got %q", RequiredCharset, cfg.Runtime.Charset))
	}
	if cfg.Runtime.Collation != RequiredCollation {
		problems = append(problems, fmt.Sprintf("collation must be %q, got %q", RequiredCollation, cfg.Runtime.Collation))
	}
	if cfg.Runtime.Location != RequiredTimeZone {
		problems = append(problems, fmt.Sprintf("time zone must be %q, got %q", RequiredTimeZone, cfg.Runtime.Location))
	}
	if !cfg.Runtime.ForeignKeyChecks {
		problems = append(problems, "foreign_key_checks must be enabled")
	}
	for _, required := range RequiredSQLModes {
		if !containsFold(cfg.Runtime.SQLModes, required) {
			problems = append(problems, "sql_mode missing "+required)
		}
	}
	if cfg.Runtime.ReadTimeout <= 0 {
		problems = append(problems, "read timeout must be positive")
	}
	if cfg.Runtime.WriteTimeout <= 0 {
		problems = append(problems, "write timeout must be positive")
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return errors.New("invalid schema test configuration: " + strings.Join(problems, "; "))
	}
	return nil
}

func RedactedDSN(cfg Config) string {
	if strings.TrimSpace(cfg.DSN) == "" {
		return "<empty>"
	}

	parsed, err := mysql.ParseDSN(cfg.DSN)
	if err != nil {
		return "<invalid>"
	}
	parsed.Passwd = "xxxxx"
	return parsed.FormatDSN()
}

func containsFold(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), want) {
			return true
		}
	}
	return false
}
