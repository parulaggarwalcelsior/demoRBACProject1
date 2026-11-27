package schema_test

import (
	"context"
	"os"
	"reflect"
	"strings"
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

func TestEnvironmentCharacterization(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open characterization connection: %v", err)
	}
	defer db.Close()

	t.Run("pad_attribute is NO PAD", func(t *testing.T) {
		var padAttribute string
		if err := db.QueryRowContext(ctx, `
SELECT PAD_ATTRIBUTE
FROM INFORMATION_SCHEMA.COLLATIONS
WHERE COLLATION_NAME = ?`, config.RequiredCollation).Scan(&padAttribute); err != nil {
			t.Fatalf("read PAD_ATTRIBUTE for %s: %v", config.RequiredCollation, err)
		}
		if padAttribute != "NO PAD" {
			t.Fatalf("PAD_ATTRIBUTE for %s = %q, want %q; target environment contract documents NO PAD and must be reconciled before this assertion is relaxed",
				config.RequiredCollation, padAttribute, "NO PAD")
		}
	})

	t.Run("case sensitive upper vs lower", func(t *testing.T) {
		var equal int
		var cmp int
		if err := db.QueryRowContext(ctx, `
SELECT
  (_utf8mb4'A' COLLATE utf8mb4_0900_bin = _utf8mb4'a' COLLATE utf8mb4_0900_bin),
  STRCMP(_utf8mb4'A' COLLATE utf8mb4_0900_bin, _utf8mb4'a' COLLATE utf8mb4_0900_bin)
`).Scan(&equal, &cmp); err != nil {
			t.Fatalf("compare case characterization: %v", err)
		}
		if equal != 0 {
			t.Fatalf("'A' = 'a' under %s returned equal=%d (STRCMP=%d), want equal=0; target environment contract records case-sensitive identity equality",
				config.RequiredCollation, equal, cmp)
		}
		if cmp == 0 {
			t.Fatalf("STRCMP('A','a') under %s = 0, want nonzero to characterize case distinction",
				config.RequiredCollation)
		}
	})

	t.Run("accent sensitive e vs e-acute", func(t *testing.T) {
		var equal int
		var cmp int
		if err := db.QueryRowContext(ctx, `
SELECT
  (_utf8mb4'é' COLLATE utf8mb4_0900_bin = _utf8mb4'e' COLLATE utf8mb4_0900_bin),
  STRCMP(_utf8mb4'é' COLLATE utf8mb4_0900_bin, _utf8mb4'e' COLLATE utf8mb4_0900_bin)
`).Scan(&equal, &cmp); err != nil {
			t.Fatalf("compare accent characterization: %v", err)
		}
		if equal != 0 {
			t.Fatalf("'é' = 'e' under %s returned equal=%d (STRCMP=%d), want equal=0; target environment contract records accent-sensitive identity equality",
				config.RequiredCollation, equal, cmp)
		}
		if cmp == 0 {
			t.Fatalf("STRCMP('é','e') under %s = 0, want nonzero to characterize accent distinction",
				config.RequiredCollation)
		}
	})

	t.Run("trailing space preserved no pad", func(t *testing.T) {
		var equal int
		var cmp int
		var paddedLength int
		if err := db.QueryRowContext(ctx, `
SELECT
  (_utf8mb4'abc' COLLATE utf8mb4_0900_bin = _utf8mb4'abc ' COLLATE utf8mb4_0900_bin),
  STRCMP(_utf8mb4'abc' COLLATE utf8mb4_0900_bin, _utf8mb4'abc ' COLLATE utf8mb4_0900_bin),
  CHAR_LENGTH(_utf8mb4'abc ' COLLATE utf8mb4_0900_bin)
`).Scan(&equal, &cmp, &paddedLength); err != nil {
			t.Fatalf("compare trailing-space characterization: %v", err)
		}
		if equal != 0 {
			t.Fatalf("'abc' = 'abc ' under %s returned equal=%d (STRCMP=%d), want equal=0; target environment contract records NO PAD trailing-space preservation",
				config.RequiredCollation, equal, cmp)
		}
		if cmp == 0 {
			t.Fatalf("STRCMP('abc','abc ') under %s = 0, want nonzero to characterize NO PAD trailing-space preservation",
				config.RequiredCollation)
		}
		if paddedLength != 4 {
			t.Fatalf("CHAR_LENGTH('abc ') under %s = %d, want 4; NO PAD retains the trailing space in the stored value",
				config.RequiredCollation, paddedLength)
		}
	})

	t.Run("session collation matches required", func(t *testing.T) {
		snapshot, err := inspectEnvironment(ctx, db)
		if err != nil {
			t.Fatalf("inspect session environment: %v", err)
		}
		if !strings.EqualFold(snapshot.Collation, config.RequiredCollation) {
			t.Fatalf("session collation = %q, want %q; characterization applies only under the selected target collation",
				snapshot.Collation, config.RequiredCollation)
		}
	})
}
