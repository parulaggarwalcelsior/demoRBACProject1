package config

import (
	"os"
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
	return FromDSN(os.Getenv(EnvSchemaTestDSN))
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
		return Config{}, err
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
	return nil
}
