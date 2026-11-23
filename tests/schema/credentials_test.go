package schema_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func TestCredentialUpgrade(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	db, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open credential upgrade connection: %v", err)
	}
	defer db.Close()

	t.Run("empty fixture refuses credential entry before migration", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		before, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot empty fixture before credential entry check: %v", err)
		}
		acceptedCredentialEntry := checkCredentialEntry(ctx, db) == nil
		after, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot empty fixture after credential entry check: %v", err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("empty credential entry check changed objects/history\nbefore=%+v\nafter=%+v", before, after)
		}
		if acceptedCredentialEntry {
			t.Fatalf("expected missing legacy user to refuse credential entry before migration invocation")
		}
	})

	t.Run("legacy only upgrade preserves profile and initializes protected target columns", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyCredentialLegacyFixture(t, ctx)
		applyCredentialSentinel(t, ctx)
		beforeProfiles := snapshotLegacyProfiles(t, ctx, db)
		beforeDefinitions := snapshotUserProfileDefinitions(t, ctx, db)
		beforeSentinel := snapshotCredentialSentinel(t, ctx, db)
		assertLegacyPasswordValuesAbsent(t, ctx, db)

		if err := checkCredentialEntry(ctx, db); err != nil {
			t.Fatalf("legacy-only credential entry refused before migration: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, credentialsTrack, credentialsTrack.Target); err != nil {
			t.Fatalf("upgrade legacy credentials: %v", err)
		}
		assertTrackVersion(t, ctx, credentialsTrack, credentialsTrack.Target, false)
		assertNoColumn(t, ctx, db, "user", "password")
		assertColumn(t, ctx, db, "user", "password_hash", "varchar(255)", true, "null")
		assertColumn(t, ctx, db, "user", "must_reset_password", "tinyint(1)", false, "1")
		assertCredentialProfilesUnchanged(t, beforeProfiles, snapshotLegacyProfiles(t, ctx, db))
		assertUserProfileDefinitionsUnchanged(t, beforeDefinitions, snapshotUserProfileDefinitions(t, ctx, db))
		assertInitialCredentialTargets(t, ctx, db)
		assertCredentialSentinelUnchanged(t, beforeSentinel, snapshotCredentialSentinel(t, ctx, db))
	})

	t.Run("incompatible entry refuses before migration", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		assertExecOK(t, ctx, db, "CREATE TABLE `user` (email VARCHAR(60) NOT NULL PRIMARY KEY, password_hash VARCHAR(255) NULL, must_reset_password BOOLEAN NOT NULL DEFAULT TRUE) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin")
		before, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot incompatible credential entry before refusal: %v", err)
		}
		if err := checkCredentialEntry(ctx, db); err == nil {
			t.Fatalf("expected existing target credential columns to refuse credential entry before migration invocation")
		}
		assertTrackVersion(t, ctx, credentialsTrack, 0, false)
		after, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot incompatible credential entry after refusal: %v", err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("incompatible credential entry check changed objects/history\nbefore=%+v\nafter=%+v", before, after)
		}
	})

	t.Run("initial target schema refuses credential entry before migration", func(t *testing.T) {
		if !schemaMigrationAuthored() {
			t.Skip("initial target schema migration is authored in T016-T019; T020 records this case red before implementing the preflight")
		}
		resetCredentialUpgrade(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply initial target schema before credential entry check: %v", err)
		}
		before, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot initial target before credential entry check: %v", err)
		}
		if err := checkCredentialEntry(ctx, db); err == nil {
			t.Fatalf("expected initial target schema to refuse credential entry before migration invocation")
		}
		assertTrackVersion(t, ctx, credentialsTrack, 0, false)
		after, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot initial target after credential entry check: %v", err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("initial target credential entry check changed objects/history\nbefore=%+v\nafter=%+v", before, after)
		}
	})

	t.Run("repeated cleanup leaves no credential owned objects", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyCredentialLegacyFixture(t, ctx)
		applyCredentialSentinel(t, ctx)
		resetCredentialUpgrade(t, ctx, db)
		resetCredentialUpgrade(t, ctx, db)
		assertTrackVersion(t, ctx, credentialsTrack, 0, false)
		assertNoBusinessTables(t, ctx, db)
	})

	t.Run("credential down refuses and retains protected data with dirty state", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyCredentialLegacyFixture(t, ctx)
		if err := migrateToVersion(ctx, startupConfig, credentialsTrack, credentialsTrack.Target); err != nil {
			t.Fatalf("upgrade legacy credentials before refused down: %v", err)
		}
		assertExecOK(t, ctx, db, "UPDATE `user` SET password_hash = ?, must_reset_password = FALSE WHERE email = ?", "synthetic-protected-hash", "legacy-one@example.test")
		beforeProfiles := snapshotLegacyProfiles(t, ctx, db)
		beforeCredentials := snapshotProtectedCredentials(t, ctx, db)

		err := migrateToVersion(ctx, startupConfig, credentialsTrack, 0)
		if err == nil {
			t.Fatalf("expected protected credential replacement down to refuse")
		}
		assertTrackDirty(t, ctx, credentialsTrack)
		assertNoColumn(t, ctx, db, "user", "password")
		assertCredentialProfilesUnchanged(t, beforeProfiles, snapshotLegacyProfiles(t, ctx, db))
		if got := snapshotProtectedCredentials(t, ctx, db); !reflect.DeepEqual(beforeCredentials, got) {
			t.Fatalf("refused credential down changed protected data\nbefore=%+v\nafter=%+v", beforeCredentials, got)
		}
	})
}

type legacyProfileRow struct {
	Email   string
	Name    sql.NullString
	Mobile  sql.NullInt64
	Address sql.NullString
	Pincode sql.NullInt64
}

type protectedCredentialRow struct {
	Email             string
	PasswordHash      sql.NullString
	MustResetPassword int
}

func resetCredentialUpgrade(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	if err := cleanupOwnedObjects(ctx, db); err != nil {
		t.Fatalf("reset credential owned objects: %v", err)
	}
	for _, track := range []MigrationTrack{schemaTrack, credentialsTrack, finalizeCredentialsTrack, legacyFixtureTrack, sentinelFixtureTrack} {
		_, _ = db.ExecContext(ctx, "DROP TABLE IF EXISTS `"+track.HistoryTable+"`")
	}
}

func applyCredentialLegacyFixture(t *testing.T, ctx context.Context) {
	t.Helper()
	if err := migrateToVersion(ctx, startupConfig, legacyFixtureTrack, legacyFixtureTrack.Target); err != nil {
		t.Fatalf("apply independent legacy user fixture: %v", err)
	}
}

func applyCredentialSentinel(t *testing.T, ctx context.Context) {
	t.Helper()
	if err := migrateToVersion(ctx, startupConfig, sentinelFixtureTrack, 1); err != nil {
		t.Fatalf("apply unrelated sentinel fixture: %v", err)
	}
}

func snapshotLegacyProfiles(t *testing.T, ctx context.Context, db *sql.DB) []legacyProfileRow {
	t.Helper()
	rows, err := db.QueryContext(ctx, "SELECT email, name, mobile, address, pincode FROM `user` ORDER BY email")
	if err != nil {
		t.Fatalf("snapshot credential profiles: %v", err)
	}
	defer rows.Close()

	var profiles []legacyProfileRow
	for rows.Next() {
		var row legacyProfileRow
		if err := rows.Scan(&row.Email, &row.Name, &row.Mobile, &row.Address, &row.Pincode); err != nil {
			t.Fatalf("scan credential profile: %v", err)
		}
		profiles = append(profiles, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate credential profiles: %v", err)
	}
	return profiles
}

func snapshotUserProfileDefinitions(t *testing.T, ctx context.Context, db *sql.DB) map[string]columnDefinition {
	t.Helper()
	snapshot, err := inspectSchemaDefinition(ctx, db)
	if err != nil {
		t.Fatalf("inspect user profile definitions: %v", err)
	}
	got := map[string]columnDefinition{}
	for _, name := range []string{"email", "name", "mobile", "address", "pincode"} {
		col, ok := snapshot.Columns["user"][name]
		if !ok {
			t.Fatalf("user.%s missing from profile definitions", name)
		}
		got[name] = col
	}
	return got
}

func assertCredentialProfilesUnchanged(t *testing.T, want, got []legacyProfileRow) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("credential upgrade changed profiles\nbefore=%+v\nafter=%+v", want, got)
	}
}

func assertUserProfileDefinitionsUnchanged(t *testing.T, want, got map[string]columnDefinition) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("credential upgrade changed profile definitions\nbefore=%+v\nafter=%+v", want, got)
	}
}

func assertLegacyPasswordValuesAbsent(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var count int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `user` WHERE password IS NOT NULL").Scan(&count); err != nil {
		t.Fatalf("count non-null legacy password placeholders: %v", err)
	}
	if count != 0 {
		t.Fatalf("legacy fixture contains %d non-null password placeholders; want absent password values", count)
	}
}

func assertInitialCredentialTargets(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var rows, nullHashes, resetRows int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*), SUM(password_hash IS NULL), SUM(must_reset_password = TRUE) FROM `user`").Scan(&rows, &nullHashes, &resetRows); err != nil {
		t.Fatalf("inspect initialized credential targets: %v", err)
	}
	if rows == 0 {
		t.Fatalf("credential upgrade produced zero account rows; want preserved legacy rows")
	}
	if nullHashes != rows || resetRows != rows {
		t.Fatalf("initialized credential targets rows/null_hashes/reset = %d/%d/%d, want every row NULL/TRUE", rows, nullHashes, resetRows)
	}
}

func snapshotProtectedCredentials(t *testing.T, ctx context.Context, db *sql.DB) []protectedCredentialRow {
	t.Helper()
	rows, err := db.QueryContext(ctx, "SELECT email, password_hash, must_reset_password FROM `user` ORDER BY email")
	if err != nil {
		t.Fatalf("snapshot protected credentials: %v", err)
	}
	defer rows.Close()

	var credentials []protectedCredentialRow
	for rows.Next() {
		var row protectedCredentialRow
		if err := rows.Scan(&row.Email, &row.PasswordHash, &row.MustResetPassword); err != nil {
			t.Fatalf("scan protected credential row: %v", err)
		}
		credentials = append(credentials, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate protected credentials: %v", err)
	}
	return credentials
}

func assertColumn(t *testing.T, ctx context.Context, db *sql.DB, table, column, wantType string, wantNullable bool, wantDefault string) {
	t.Helper()
	snapshot, err := inspectSchemaDefinition(ctx, db)
	if err != nil {
		t.Fatalf("inspect %s.%s: %v", table, column, err)
	}
	got, ok := snapshot.Columns[table][column]
	if !ok {
		t.Fatalf("%s.%s missing", table, column)
	}
	if got.Type != wantType || got.Nullable != wantNullable || got.Default != wantDefault {
		t.Fatalf("%s.%s = type %q nullable %t default %q, want type %q nullable %t default %q",
			table, column, got.Type, got.Nullable, got.Default, wantType, wantNullable, wantDefault)
	}
}

func assertNoColumn(t *testing.T, ctx context.Context, db *sql.DB, table, column string) {
	t.Helper()
	snapshot, err := inspectSchemaDefinition(ctx, db)
	if err != nil {
		t.Fatalf("inspect absence of %s.%s: %v", table, column, err)
	}
	if _, ok := snapshot.Columns[table][column]; ok {
		t.Fatalf("%s.%s is present; want absent", table, column)
	}
}

func assertNoBusinessTables(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	snapshot, err := inspectSchemaDefinition(ctx, db)
	if err != nil {
		t.Fatalf("inspect cleaned credential fixture: %v", err)
	}
	if got := presentBusinessTables(snapshot); len(got) != 0 {
		t.Fatalf("business tables after repeated cleanup = %v, want none", got)
	}
}

func snapshotCredentialSentinel(t *testing.T, ctx context.Context, db *sql.DB) string {
	t.Helper()
	var exists int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'sentinel_owned'`).Scan(&exists); err != nil {
		t.Fatalf("inspect unrelated sentinel existence: %v", err)
	}
	if exists == 0 {
		return ""
	}
	return snapshotSentinel(t, ctx, db)
}

func assertCredentialSentinelUnchanged(t *testing.T, want, got string) {
	t.Helper()
	if got != want {
		t.Fatalf("credential upgrade changed unrelated sentinel\nwant=%q\n got=%q", want, got)
	}
}
