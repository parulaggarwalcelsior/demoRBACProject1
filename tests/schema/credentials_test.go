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

func TestCredentialFinalization(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	db, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open credential finalization connection: %v", err)
	}
	defer db.Close()

	pendingStates := []struct {
		name  string
		hash  any
		reset any
	}{
		{"null_hash_reset_true", nil, 1},
		{"null_hash_reset_false", nil, 0},
		{"non_null_hash_reset_true", "synthetic-pending-hash-true", 1},
		{"non_null_hash_nonzero_reset_flag_two", "synthetic-pending-hash-two", 2},
	}

	for _, state := range pendingStates {
		state := state
		t.Run("harness preflight refuses pending "+state.name, func(t *testing.T) {
			resetCredentialUpgrade(t, ctx, db)
			applyFinalizationInitialSchema(t, ctx)
			insertFinalizationAccount(t, ctx, db, "preflight-"+state.name+"@example.test", state.hash, state.reset)
			beforeState := snapshotFinalizationState(t, ctx, db)
			beforeDef := snapshotUserHashDefinition(t, ctx, db)

			if err := checkFinalizationReadiness(ctx, db); err == nil {
				t.Fatalf("expected harness preflight to refuse pending state %q before migration invocation", state.name)
			}

			assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)
			if got := snapshotFinalizationState(t, ctx, db); !reflect.DeepEqual(beforeState, got) {
				t.Fatalf("preflight refusal changed account state for %q\nbefore=%+v\nafter=%+v", state.name, beforeState, got)
			}
			if got := snapshotUserHashDefinition(t, ctx, db); !reflect.DeepEqual(beforeDef, got) {
				t.Fatalf("preflight refusal changed password_hash definition for %q\nbefore=%+v\nafter=%+v", state.name, beforeDef, got)
			}
			assertNoFinalizationGuardRoutine(t, ctx, db)
		})

		t.Run("direct migration guard refuses pending "+state.name, func(t *testing.T) {
			resetCredentialUpgrade(t, ctx, db)
			applyFinalizationInitialSchema(t, ctx)
			assertFinalizationWriterQuiescence(t, db)
			insertFinalizationAccount(t, ctx, db, "guard-"+state.name+"@example.test", state.hash, state.reset)
			beforeState := snapshotFinalizationState(t, ctx, db)
			beforeDef := snapshotUserHashDefinition(t, ctx, db)

			if err := migrateFinalizationDirect(ctx, startupConfig); err == nil {
				t.Fatalf("expected direct finalization migration to refuse pending state %q", state.name)
			}

			if got := snapshotFinalizationState(t, ctx, db); !reflect.DeepEqual(beforeState, got) {
				t.Fatalf("direct migration guard refusal changed account state for %q\nbefore=%+v\nafter=%+v", state.name, beforeState, got)
			}
			if got := snapshotUserHashDefinition(t, ctx, db); !reflect.DeepEqual(beforeDef, got) {
				t.Fatalf("direct migration guard refusal changed password_hash definition for %q\nbefore=%+v\nafter=%+v", state.name, beforeDef, got)
			}

			ownership, err := inspectFinalizationOwnership(ctx, db, startupConfig)
			if err != nil {
				t.Fatalf("inspect finalization ownership after direct guard refusal for %q: %v", state.name, err)
			}
			if !ownership.Dirty {
				t.Fatalf("direct guard refusal for %q left finalization history clean; want dirty=true after SQL guard SIGNAL", state.name)
			}
			if err := discardFinalizationFixture(ctx, db); err != nil {
				t.Fatalf("discard disposable finalization fixture after direct guard refusal for %q: %v", state.name, err)
			}
		})
	}

	t.Run("writers quiesced precondition single writer pool", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyFinalizationInitialSchema(t, ctx)
		assertFinalizationWriterQuiescence(t, db)
		if _, err := db.ExecContext(ctx, "SELECT 1"); err != nil {
			t.Fatalf("exercise single-writer connection before finalization: %v", err)
		}
	})

	t.Run("empty table allows finalization and cleans helper routine", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyFinalizationInitialSchema(t, ctx)

		if err := checkFinalizationReadiness(ctx, db); err != nil {
			t.Fatalf("empty user table preflight refused: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, finalizeCredentialsTrack.Target); err != nil {
			t.Fatalf("apply finalization on empty user table: %v", err)
		}

		assertTrackVersion(t, ctx, finalizeCredentialsTrack, finalizeCredentialsTrack.Target, false)
		assertColumn(t, ctx, db, "user", "password_hash", "varchar(255)", false, "")
		assertNoFinalizationGuardRoutine(t, ctx, db)
	})

	t.Run("all ready accounts allow finalization preserving hashes flags and profiles", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyFinalizationInitialSchema(t, ctx)
		seedReadyAccounts(t, ctx, db)

		beforeProfiles := snapshotFinalizationProfiles(t, ctx, db)
		beforeCreds := snapshotProtectedCredentials(t, ctx, db)
		beforeState := snapshotFinalizationState(t, ctx, db)

		if err := checkFinalizationReadiness(ctx, db); err != nil {
			t.Fatalf("all-ready preflight refused: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, finalizeCredentialsTrack.Target); err != nil {
			t.Fatalf("apply finalization on all-ready user table: %v", err)
		}

		assertTrackVersion(t, ctx, finalizeCredentialsTrack, finalizeCredentialsTrack.Target, false)
		assertColumn(t, ctx, db, "user", "password_hash", "varchar(255)", false, "")
		assertNoColumn(t, ctx, db, "user", "password")
		assertNoFinalizationGuardRoutine(t, ctx, db)

		if got := snapshotFinalizationProfiles(t, ctx, db); !reflect.DeepEqual(beforeProfiles, got) {
			t.Fatalf("finalization changed profiles\nbefore=%+v\nafter=%+v", beforeProfiles, got)
		}
		if got := snapshotProtectedCredentials(t, ctx, db); !reflect.DeepEqual(beforeCreds, got) {
			t.Fatalf("finalization changed hashes/flags\nbefore=%+v\nafter=%+v", beforeCreds, got)
		}
		if got := snapshotFinalizationState(t, ctx, db); !reflect.DeepEqual(beforeState, got) {
			t.Fatalf("finalization changed account state\nbefore=%+v\nafter=%+v", beforeState, got)
		}
	})

	t.Run("finalized schema refuses subsequent null and omitted hash writes", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyFinalizationInitialSchema(t, ctx)

		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, finalizeCredentialsTrack.Target); err != nil {
			t.Fatalf("apply finalization before refusal probes: %v", err)
		}
		assertColumn(t, ctx, db, "user", "password_hash", "varchar(255)", false, "")

		assertExecError(t, ctx, db, nullErr, "INSERT INTO `user` (email, password_hash) VALUES (?, NULL)", "null-hash@example.test")
		assertExecError(t, ctx, db, requiredDefaultErr, "INSERT INTO `user` (email) VALUES (?)", "omitted-hash@example.test")
	})

	t.Run("finalization down restores only nullability and refuses plaintext restoration", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyFinalizationInitialSchema(t, ctx)
		seedReadyAccounts(t, ctx, db)

		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, finalizeCredentialsTrack.Target); err != nil {
			t.Fatalf("apply finalization before down: %v", err)
		}
		beforeProfiles := snapshotFinalizationProfiles(t, ctx, db)
		beforeCreds := snapshotProtectedCredentials(t, ctx, db)

		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, 0); err != nil {
			t.Fatalf("finalization down: %v", err)
		}

		assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)
		assertColumn(t, ctx, db, "user", "password_hash", "varchar(255)", true, "null")
		assertColumn(t, ctx, db, "user", "must_reset_password", "tinyint(1)", false, "1")
		assertNoColumn(t, ctx, db, "user", "password")
		assertNoFinalizationGuardRoutine(t, ctx, db)

		if got := snapshotFinalizationProfiles(t, ctx, db); !reflect.DeepEqual(beforeProfiles, got) {
			t.Fatalf("finalization down changed profiles\nbefore=%+v\nafter=%+v", beforeProfiles, got)
		}
		if got := snapshotProtectedCredentials(t, ctx, db); !reflect.DeepEqual(beforeCreds, got) {
			t.Fatalf("finalization down changed hashes/flags\nbefore=%+v\nafter=%+v", beforeCreds, got)
		}
	})

	t.Run("failed finalization surfaces truthful administrative state", func(t *testing.T) {
		resetCredentialUpgrade(t, ctx, db)
		applyFinalizationInitialSchema(t, ctx)
		insertFinalizationAccount(t, ctx, db, "dirty-pending@example.test", nil, 1)
		beforeDef := snapshotUserHashDefinition(t, ctx, db)
		beforeState := snapshotFinalizationState(t, ctx, db)

		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, finalizeCredentialsTrack.Target); err == nil {
			t.Fatalf("expected finalization to fail against pending state")
		}

		assertTrackDirty(t, ctx, finalizeCredentialsTrack)
		if got := snapshotUserHashDefinition(t, ctx, db); !reflect.DeepEqual(beforeDef, got) {
			t.Fatalf("failed finalization changed password_hash definition\nbefore=%+v\nafter=%+v", beforeDef, got)
		}
		if got := snapshotFinalizationState(t, ctx, db); !reflect.DeepEqual(beforeState, got) {
			t.Fatalf("failed finalization changed account state\nbefore=%+v\nafter=%+v", beforeState, got)
		}
		assertNoColumn(t, ctx, db, "user", "password")
	})
}

type finalizationAccountRow struct {
	Email             string
	PasswordHash      sql.NullString
	MustResetPassword int64
	Name              sql.NullString
	Mobile            sql.NullInt64
	Address           sql.NullString
	Pincode           sql.NullInt64
}

type userHashDefinition struct {
	Type     string
	Nullable bool
	Default  string
}

func applyFinalizationInitialSchema(t *testing.T, ctx context.Context) {
	t.Helper()
	if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
		t.Fatalf("apply initial target schema before finalization: %v", err)
	}
}

func insertFinalizationAccount(t *testing.T, ctx context.Context, db *sql.DB, email string, hash any, reset any) {
	t.Helper()
	assertExecOK(t, ctx, db,
		"INSERT INTO `user` (email, password_hash, must_reset_password) VALUES (?, ?, ?)",
		email, hash, reset,
	)
}

func seedReadyAccounts(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	ready := []struct {
		email string
		hash  string
	}{
		{"ready-one@example.test", "$argon2id$v=19$m=65536,t=3,p=4$readyonereadyonereadyonereadyoneready$readyonereadyonereadyonereadyoneready"},
		{"ready-two@example.test", "$argon2id$v=19$m=65536,t=3,p=4$readytworeadytworeadytworeadytworeadyt$readytworeadytworeadytworeadytworeadyt"},
		{"ready-three@example.test", "synthetic-ready-three-hash"},
	}
	for _, account := range ready {
		assertExecOK(t, ctx, db,
			"INSERT INTO `user` (email, password_hash, must_reset_password, name, mobile, address, pincode) VALUES (?, ?, 0, ?, ?, ?, ?)",
			account.email, account.hash, account.email+"-name", int64(len(account.email)), account.email+"-address", len(account.email),
		)
	}
}

func snapshotFinalizationState(t *testing.T, ctx context.Context, db *sql.DB) []finalizationAccountRow {
	t.Helper()
	rows, err := db.QueryContext(ctx, "SELECT email, password_hash, must_reset_password, name, mobile, address, pincode FROM `user` ORDER BY email")
	if err != nil {
		t.Fatalf("snapshot finalization state: %v", err)
	}
	defer rows.Close()

	var accounts []finalizationAccountRow
	for rows.Next() {
		var row finalizationAccountRow
		if err := rows.Scan(&row.Email, &row.PasswordHash, &row.MustResetPassword, &row.Name, &row.Mobile, &row.Address, &row.Pincode); err != nil {
			t.Fatalf("scan finalization account: %v", err)
		}
		accounts = append(accounts, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate finalization state: %v", err)
	}
	return accounts
}

func snapshotFinalizationProfiles(t *testing.T, ctx context.Context, db *sql.DB) []legacyProfileRow {
	t.Helper()
	return snapshotLegacyProfiles(t, ctx, db)
}

func snapshotUserHashDefinition(t *testing.T, ctx context.Context, db *sql.DB) userHashDefinition {
	t.Helper()
	snapshot, err := inspectSchemaDefinition(ctx, db)
	if err != nil {
		t.Fatalf("inspect user hash definition: %v", err)
	}
	col, ok := snapshot.Columns["user"]["password_hash"]
	if !ok {
		t.Fatalf("user.password_hash missing from snapshot")
	}
	return userHashDefinition{Type: col.Type, Nullable: col.Nullable, Default: col.Default}
}

func assertNoFinalizationGuardRoutine(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var count int
	err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM INFORMATION_SCHEMA.ROUTINES
WHERE ROUTINE_SCHEMA = DATABASE()
  AND ROUTINE_NAME = 'schema_guard_finalize_credentials'`).Scan(&count)
	if err != nil {
		t.Fatalf("inspect finalization guard routine presence: %v", err)
	}
	if count != 0 {
		t.Fatalf("finalization guard routine schema_guard_finalize_credentials is present; want absent after completion")
	}
}

// checkFinalizationReadiness is the Go harness preflight for the finalization track; it
// lives in tests/schema/helpers_test.go (added in T027) alongside selectFinalizationTrack,
// migrateFinalizationDirect, assertFinalizationWriterQuiescence, inspectFinalizationOwnership
// and discardFinalizationFixture.
