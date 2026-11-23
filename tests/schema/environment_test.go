package schema_test

import (
	"context"
	"os"
	"reflect"
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
				CharacterSet:     config.RequiredCharset,
				Collation:        config.RequiredCollation,
				InnoDBSupport:    "YES",
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

func TestEnvironment(t *testing.T) {
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
	if err := validateEnvironment(startupConfig, snapshot); err != nil {
		t.Fatalf("unsupported SQL connection environment: %v", err)
	}

	migrationDB, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open migration connection before mutation: %v", err)
	}
	defer migrationDB.Close()

	migrationSnapshot, err := inspectEnvironment(ctx, migrationDB)
	if err != nil {
		t.Fatalf("inspect migration connection environment: %v", err)
	}
	if err := validateEnvironment(startupConfig, migrationSnapshot); err != nil {
		t.Fatalf("unsupported migration connection environment: %v", err)
	}
}

func TestEnvironmentRefusesInvalidBeforeMigration(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := openReadOnly(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open baseline connection: %v", err)
	}
	defer db.Close()

	before, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
	if err != nil {
		t.Fatalf("snapshot state before invalid migration input: %v", err)
	}

	invalid := startupConfig
	invalid.Database = "other"
	if migrator, migrationDB, err := newMigrator(ctx, invalid, schemaTrack); err == nil {
		if migrator != nil {
			_, _ = migrator.Close()
		}
		if migrationDB != nil {
			_ = migrationDB.Close()
		}
		t.Fatalf("expected invalid migration configuration refusal before migration invocation")
	}

	after, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
	if err != nil {
		t.Fatalf("snapshot state after invalid migration input: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("invalid migration input changed business objects/history: before=%+v after=%+v", before, after)
	}
}

func validConfigForPreflight() config.Config {
	return config.Config{
		DSN:      "schema_test:secret@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true",
		Database: config.RequiredDatabase,
		Runtime:  config.RequiredRuntimeSettings(),
	}
}
