package schema_test

import (
	"context"
	"os"
	"testing"
	"time"

	"demoRBACProject1/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

var startupConfig config.Config
var startupConfigErr error

func TestMain(m *testing.M) {
	startupConfig, startupConfigErr = config.Load()
	os.Exit(m.Run())
}

func TestEnvironmentPreflight(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	tests := []struct {
		name        string
		cfg         config.Config
		snapshot    EnvironmentSnapshot
		wantRefusal bool
	}{
		{
			name: "valid runtime settings",
			cfg:  validConfigForPreflight(),
			snapshot: EnvironmentSnapshot{
				Database:         config.RequiredDatabase,
				Version:          "8.4.7",
				SQLMode:          "STRICT_ALL_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO",
				TimeZone:         config.RequiredTimeZone,
				ForeignKeyChecks: "1",
			},
		},
		{
			name:        "wrong namespace refused",
			cfg:         validConfigForPreflight(),
			snapshot:    EnvironmentSnapshot{Database: "other", Version: "8.4.7", SQLMode: "STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO", TimeZone: config.RequiredTimeZone, ForeignKeyChecks: "1"},
			wantRefusal: true,
		},
		{
			name:        "unsupported server refused",
			cfg:         validConfigForPreflight(),
			snapshot:    EnvironmentSnapshot{Database: config.RequiredDatabase, Version: "5.7.44", SQLMode: "STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO", TimeZone: config.RequiredTimeZone, ForeignKeyChecks: "1"},
			wantRefusal: true,
		},
		{
			name:        "missing strict settings refused",
			cfg:         validConfigForPreflight(),
			snapshot:    EnvironmentSnapshot{Database: config.RequiredDatabase, Version: "8.4.7", SQLMode: "NO_ENGINE_SUBSTITUTION", TimeZone: config.RequiredTimeZone, ForeignKeyChecks: "1"},
			wantRefusal: true,
		},
		{
			name:        "disabled foreign keys refused",
			cfg:         validConfigForPreflight(),
			snapshot:    EnvironmentSnapshot{Database: config.RequiredDatabase, Version: "8.4.7", SQLMode: "STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO", TimeZone: config.RequiredTimeZone, ForeignKeyChecks: "0"},
			wantRefusal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnvironment(tt.cfg, tt.snapshot)
			if tt.wantRefusal && err == nil {
				t.Fatalf("expected environment refusal, got nil")
			}
			if !tt.wantRefusal && err != nil {
				t.Fatalf("unexpected environment refusal: %v", err)
			}
		})
	}
}

func TestEnvironmentReadOnlyConnection(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := openReadOnly(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open read-only connection: %v", err)
	}
	defer db.Close()

	snapshot, err := inspectEnvironment(ctx, db)
	if err != nil {
		t.Fatalf("inspect environment: %v", err)
	}

	if snapshot.Database == "" {
		t.Fatalf("expected selected database, got empty value")
	}
}

func validConfigForPreflight() config.Config {
	return config.Config{
		DSN:      "schema_test:secret@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true",
		Database: config.RequiredDatabase,
		Runtime:  config.RequiredRuntimeSettings(),
	}
}
