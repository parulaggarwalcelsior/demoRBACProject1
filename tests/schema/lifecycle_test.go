package schema_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

func TestSchemaLifecycle(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	db, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open schema lifecycle connection: %v", err)
	}
	defer db.Close()

	t.Run("initial up and already applied no change", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)

		before, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot before initial schema up: %v", err)
		}
		if err := checkOccupiedBootstrap(ctx, db); err != nil {
			t.Fatalf("empty bootstrap inventory refused: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("initial schema up: %v", err)
		}
		assertTrackVersion(t, ctx, schemaTrack, schemaTrack.Target, false)
		assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)

		firstDefinition := inspectLifecycleDefinition(t, ctx, db)
		assertExactBusinessInventory(t, firstDefinition)
		assertBusinessTablesEmpty(t, firstDefinition)

		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("already-applied schema up: %v", err)
		}
		assertTrackVersion(t, ctx, schemaTrack, schemaTrack.Target, false)
		secondDefinition := inspectLifecycleDefinition(t, ctx, db)
		if !reflect.DeepEqual(firstDefinition, secondDefinition) {
			t.Fatalf("already-applied schema up changed metadata\nbefore=%#v\nafter=%#v", firstDefinition, secondDefinition)
		}

		after, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot after already-applied schema up: %v", err)
		}
		if len(before.Tables) != 0 && reflect.DeepEqual(before, after) {
			t.Fatalf("initial up left pre-existing populated snapshot unchanged: before=%+v after=%+v", before, after)
		}
	})

	t.Run("occupied inventory refused before migration", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)
		assertExecOK(t, ctx, db, "CREATE TABLE product (pid VARCHAR(30) NOT NULL PRIMARY KEY) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin")
		before, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot occupied inventory before refusal: %v", err)
		}
		if err := checkOccupiedBootstrap(ctx, db); err == nil {
			t.Fatalf("expected occupied bootstrap inventory refusal before migration invocation")
		}
		assertTrackVersion(t, ctx, schemaTrack, 0, false)
		after, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot occupied inventory after refusal: %v", err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("occupied bootstrap refusal changed objects/history\nbefore=%+v\nafter=%+v", before, after)
		}
	})

	t.Run("legacy user fixture directly refuses occupied bootstrap", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, legacyFixtureTrack, legacyFixtureTrack.Target); err != nil {
			t.Fatalf("apply legacy user fixture before target migrations exist: %v", err)
		}
		before, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot legacy fixture before occupied bootstrap check: %v", err)
		}
		acceptedOccupiedBootstrap := checkOccupiedBootstrap(ctx, db) == nil
		after, err := snapshotBusinessAndHistory(ctx, db, startupConfig)
		if err != nil {
			t.Fatalf("snapshot legacy fixture after occupied bootstrap check: %v", err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("legacy occupied bootstrap check changed objects/history\nbefore=%+v\nafter=%+v", before, after)
		}
		if acceptedOccupiedBootstrap {
			t.Fatalf("expected legacy user fixture to refuse occupied bootstrap before migration invocation")
		}
	})

	t.Run("populated bounded down up keeps sentinel and behavior", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, sentinelFixtureTrack, 1); err != nil {
			t.Fatalf("apply unrelated sentinel fixture version 1: %v", err)
		}
		seedSentinel(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema before populated lifecycle: %v", err)
		}
		populateLifecycleBusinessRows(t, ctx, db)

		beforeDefinition := inspectLifecycleDefinition(t, ctx, db)
		beforeSentinel := snapshotSentinel(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, 0); err != nil {
			t.Fatalf("bounded schema down: %v", err)
		}
		assertTrackVersion(t, ctx, schemaTrack, 0, false)
		assertSentinelUnchanged(t, beforeSentinel, snapshotSentinel(t, ctx, db))

		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("schema up after bounded down: %v", err)
		}
		assertTrackVersion(t, ctx, schemaTrack, schemaTrack.Target, false)
		assertTrackVersion(t, ctx, sentinelFixtureTrack, 1, false)
		assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)
		assertSentinelUnchanged(t, beforeSentinel, snapshotSentinel(t, ctx, db))
		afterDefinition := inspectLifecycleDefinition(t, ctx, db)
		beforeStructure := beforeDefinition
		beforeStructure.Rows = nil
		afterStructure := afterDefinition
		afterStructure.Rows = nil
		if !reflect.DeepEqual(beforeStructure, afterStructure) {
			t.Fatalf("schema definition after down/up differs\nbefore=%#v\nafter=%#v", beforeStructure, afterStructure)
		}
		assertBusinessTablesEmpty(t, afterDefinition)
		assertForeignKeyChecksEnabled(t, ctx, db)
		assertLifecycleBehavior(t, ctx, db)
	})

	t.Run("external reference teardown refusal leaves partial state for discard", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, sentinelFixtureTrack, 1); err != nil {
			t.Fatalf("apply unrelated sentinel fixture before external reference case: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema before external reference case: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, sentinelFixtureTrack, 2); err != nil {
			t.Fatalf("apply external-reference sentinel after product exists: %v", err)
		}
		populateLifecycleBusinessRows(t, ctx, db)
		beforeSentinel := snapshotSentinel(t, ctx, db)

		err := migrateToVersion(ctx, startupConfig, schemaTrack, 0)
		if err == nil {
			t.Fatalf("expected external reference to refuse bounded schema down")
		}
		assertTrackDirty(t, ctx, schemaTrack)
		partial, err := snapshotOwnedObjects(ctx, db, startupConfig, schemaTrack)
		if err != nil {
			t.Fatalf("inspect partial DDL state after teardown refusal: %v", err)
		}
		if len(partial.Tables) == 0 {
			t.Fatalf("expected partial state inspection after teardown refusal, got no tracked tables")
		}
		assertSentinelUnchanged(t, beforeSentinel, snapshotSentinel(t, ctx, db))
	})

	t.Run("schema up does not implicitly finalize credentials", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema for implicit finalization check: %v", err)
		}
		assertTrackVersion(t, ctx, schemaTrack, schemaTrack.Target, false)
		assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)

		snapshot := inspectLifecycleDefinition(t, ctx, db)
		hash, ok := snapshot.Columns["user"]["password_hash"]
		if !ok {
			t.Fatalf("user.password_hash missing after implicit-finalization check")
		}
		if !hash.Nullable || hash.Default != "null" {
			t.Fatalf("user.password_hash after schema up = nullable %t default %q, want nullable true default null (no implicit finalization)",
				hash.Nullable, hash.Default)
		}
		reset, ok := snapshot.Columns["user"]["must_reset_password"]
		if !ok {
			t.Fatalf("user.must_reset_password missing after implicit-finalization check")
		}
		if reset.Nullable || reset.Default != "1" {
			t.Fatalf("user.must_reset_password after schema up = nullable %t default %q, want nullable false default 1 (no implicit finalization)",
				reset.Nullable, reset.Default)
		}
		assertNoFinalizationGuardRoutine(t, ctx, db)
	})

	t.Run("finalization down before bootstrap reconstruction", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema before finalization-first reconstruction: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, finalizeCredentialsTrack.Target); err != nil {
			t.Fatalf("apply finalization on empty table before reconstruction: %v", err)
		}
		assertTrackVersion(t, ctx, finalizeCredentialsTrack, finalizeCredentialsTrack.Target, false)
		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, 0); err != nil {
			t.Fatalf("finalization down before bootstrap reconstruction: %v", err)
		}
		assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)
		postFinalize := inspectLifecycleDefinition(t, ctx, db)
		if hash := postFinalize.Columns["user"]["password_hash"]; !hash.Nullable || hash.Default != "null" {
			t.Fatalf("user.password_hash after finalization down = nullable %t default %q, want nullable true default null",
				hash.Nullable, hash.Default)
		}

		if err := migrateToVersion(ctx, startupConfig, schemaTrack, 0); err != nil {
			t.Fatalf("schema down after finalization down: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("schema up after bootstrap reconstruction: %v", err)
		}
		assertTrackVersion(t, ctx, schemaTrack, schemaTrack.Target, false)
		assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)
		rebuilt := inspectLifecycleDefinition(t, ctx, db)
		if hash := rebuilt.Columns["user"]["password_hash"]; !hash.Nullable || hash.Default != "null" {
			t.Fatalf("user.password_hash after reconstruction = nullable %t default %q, want nullable true default null",
				hash.Nullable, hash.Default)
		}
		assertNoFinalizationGuardRoutine(t, ctx, db)
	})

	t.Run("recreated nullable hashes do not retain stale finalized history", func(t *testing.T) {
		resetSchemaLifecycle(t, ctx, db)
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema before stale finalization history check: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, finalizeCredentialsTrack.Target); err != nil {
			t.Fatalf("apply finalization for stale history check: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, finalizeCredentialsTrack, 0); err != nil {
			t.Fatalf("finalization down for stale history check: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, 0); err != nil {
			t.Fatalf("schema down for stale history check: %v", err)
		}
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("schema up for stale history check: %v", err)
		}
		assertTrackVersion(t, ctx, finalizeCredentialsTrack, 0, false)
		rebuilt := inspectLifecycleDefinition(t, ctx, db)
		hash, ok := rebuilt.Columns["user"]["password_hash"]
		if !ok {
			t.Fatalf("user.password_hash missing after recreated schema")
		}
		if !hash.Nullable || hash.Default != "null" {
			t.Fatalf("stale finalized history leaked onto recreated nullable hashes: nullable=%t default=%q, want nullable true default null",
				hash.Nullable, hash.Default)
		}
		assertNoFinalizationGuardRoutine(t, ctx, db)
	})
}

func resetSchemaLifecycle(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	if err := cleanupOwnedObjects(ctx, db); err != nil {
		t.Fatalf("reset lifecycle owned objects: %v", err)
	}
	for _, track := range []MigrationTrack{schemaTrack, credentialsTrack, finalizeCredentialsTrack, legacyFixtureTrack, sentinelFixtureTrack} {
		_, _ = db.ExecContext(ctx, "DROP TABLE IF EXISTS `"+track.HistoryTable+"`")
	}
}

func inspectLifecycleDefinition(t *testing.T, ctx context.Context, db *sql.DB) schemaDefinitionSnapshot {
	t.Helper()
	snapshot, err := inspectSchemaDefinition(ctx, db)
	if err != nil {
		t.Fatalf("inspect lifecycle schema definition: %v", err)
	}
	return snapshot
}

func assertTrackVersion(t *testing.T, ctx context.Context, track MigrationTrack, wantVersion uint, wantDirty bool) {
	t.Helper()
	version, dirty, err := migrationVersion(ctx, startupConfig, track)
	if err != nil {
		t.Fatalf("inspect %s migration version: %v", track.Name, err)
	}
	if version != wantVersion || dirty != wantDirty {
		t.Fatalf("%s version/dirty = %d/%t, want %d/%t", track.Name, version, dirty, wantVersion, wantDirty)
	}
}

func assertTrackDirty(t *testing.T, ctx context.Context, track MigrationTrack) {
	t.Helper()
	_, dirty, err := migrationVersion(ctx, startupConfig, track)
	if errors.Is(err, migrate.ErrNilVersion) {
		t.Fatalf("%s has nil version after failed migration; want explicit dirty state", track.Name)
	}
	if err != nil {
		t.Fatalf("inspect %s dirty state: %v", track.Name, err)
	}
	if !dirty {
		t.Fatalf("%s dirty = false after failed migration; want dirty=true", track.Name)
	}
}

func populateLifecycleBusinessRows(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	insertProduct(t, ctx, db, "life-product")
	insertUser(t, ctx, db, "life@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid, quantity, amount, shipped) VALUES (?, ?, ?, ?, ?)", "life-order", "life-product", -2, "12.34", 2)
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid, quantity) VALUES (?, ?, ?)", "life@example.test", "life-product", 1)
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name, author) VALUES (?, ?, NULL)", "life@example.test", "life-book")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity, amount) VALUES (?, ?, ?, ?, ?)", "life-tx", "life@example.test", "life-product", 1, "9.99")
}

func assertLifecycleBehavior(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	insertProduct(t, ctx, db, "life-behavior-product")
	insertUser(t, ctx, db, "life-behavior@example.test")
	assertExecError(t, ctx, db, fkChildErr, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "life-missing-order", "missing-product")
	assertExecError(t, ctx, db, fkChildErr, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "missing@example.test", "life-behavior-product")
	assertExecError(t, ctx, db, checkErr, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?)", "life-bad-quantity", "life-behavior@example.test", "life-behavior-product", 0)
}

func seedSentinel(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	assertExecOK(t, ctx, db, "INSERT INTO sentinel_owned (label) VALUES (?)", "unrelated-sentinel")
}

func snapshotSentinel(t *testing.T, ctx context.Context, db *sql.DB) string {
	t.Helper()
	var value sql.NullString
	err := db.QueryRowContext(ctx, "SELECT COALESCE(GROUP_CONCAT(label ORDER BY label SEPARATOR '\\n'), '') FROM sentinel_owned").Scan(&value)
	if err != nil {
		t.Fatalf("snapshot unrelated sentinel: %v", err)
	}
	return value.String
}

func assertSentinelUnchanged(t *testing.T, want, got string) {
	t.Helper()
	if got != want {
		t.Fatalf("sentinel changed\nwant=%q\n got=%q", want, got)
	}
}
