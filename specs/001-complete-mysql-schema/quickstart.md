# Quickstart: Validate the Complete MySQL Schema

**Date**: 2026-11-10 | **Goal**: java to go | **Feature**: [spec.md](spec.md)

This guide specifies the implementation phase's validation workflow for the Java-to-Go persistence foundation. At planning completion, the referenced Go harness, migrations and Compose file do not exist yet. The commands below become runnable after tasks implement them; no database or Go checks have run during this planning step.

## Prerequisites and configuration

- A disposable Linux environment with Docker Compose, Go 1.24-compatible toolchain, goimports and golangci-lint. Implementation pins exact tool/module versions and the MySQL 8.4 image digest.
- An implementation checkout containing the source layout in [plan.md](plan.md), versioned up/down migration pairs and a changelog. Track all migration files in Git; this planning workspace currently has no repository/actual branch.
- Runtime-injected SCHEMA_TEST_MYSQL_PASSWORD for container provisioning and SCHEMA_TEST_DSN for the Go harness. Supply these through the environment/secret mechanism; no sample credential literal belongs in source, fixtures, documentation or logs. The DSN selects the isolated shopping-cart database and the disposable instance.
- compose.schema.yaml provisions shopping-cart through MYSQL_DATABASE, publishes its test-only connection and exposes a health check. The password is injected, not embedded. The harness must refuse a wrong namespace/unsupported environment before mutation.
- Only internal/config reads environment, once at test startup into Config. Helpers receive that config and apply the declared settings to every SQL and migration connection. The harness imports golang-migrate's file/MySQL drivers and the MySQL SQL driver; tests do not shell out to hand-apply schema SQL.

The full [schema contract](contracts/schema-contract.md) defines modes, collation, precision, field boundaries and expected SQL errors. The [migration contract](contracts/migration-contract.md) defines sources/history names and credential entry paths. Avoid duplicating SQL definitions in this guide.

## Start the isolated database

From the working tree root, with runtime configuration already supplied:

```sh
docker compose -f compose.schema.yaml up -d --wait
go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
```

Expected: the container is healthy; MySQL reports the pinned 8.4 version; the selected database is shopping-cart; required strict modes, utf8mb4, selected collation, UTC session time zone and foreign_key_checks=1 are verified. The test must fail, not skip, if prerequisites/configuration are missing. Record tool/server versions and image digest without exposing credentials.

## Demonstrate acceptance failures before implementation

During initial development, T004 supplies only typed configuration loading, read-only connectivity and callable, non-enforcing configuration/environment checks. Before T005 implements their refusals, write and run:

```sh
go test -race ./internal/config -run '^TestConfig$' -count=1 -v
go test -race ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
```

Expected red results identify missing refusals: invalid configuration or an unsupported environment is accepted. Use valid startup configuration and pass controlled Config values/isolated connection settings into the negative cases without rereading environment; retain a successful read-only container connection as the positive control. Startup aborts, connection errors and skips do not count as the expected failure. Record assertions, commands and exit status in validation.md. Implement those checks in T005, then require these tests and TestEnvironment green in T006 before any domain red gate. The earlier TestEnvironment command is the completed baseline check, not a substitute for these development-stage red results.

Keep occupied-bootstrap and incompatible-credential-entry checks non-enforcing in migration scaffolding until T011/T012 author their cases and T015 records their missing-refusal failures. Invoke these entry checks directly on the legacy user fixture (occupied bootstrap) and an empty fixture (incompatible credentials), without migration mutation or source-file lookup. Implement occupied-inventory refusal in T016; before credential-entry implementation in T020, also demonstrate refusal of the now-existing initial target schema red. Require the relevant refusal cases green with unchanged business objects/history before invoking each track. Finalization keeps its own T023/T024 red gate before T025–T027 implement its SQL/helper behavior.

Retain the existing SQL acceptance gates as well. Establish a working environment and executable tests first. With complete-schema DDL absent, run:

```sh
go test -race ./tests/schema -run '^TestSchemaDefinitions$' -count=1 -v
```

Expected red result: assertions identify absent required schema behavior/objects. A compile error, connection refusal or missing external tool does not prove a missing constraint. Repeat this discipline for each new behavior: value/default checks, referential actions, credential replacement and finalization. Retain the expected failure's assertion and exit status before implementing that behavior. Do not run this red-state instruction against a completed schema and claim a new failure demonstration.

## Validate setup, values and referential rules

After implementing the complete-schema migration pair:

```sh
go test -race ./tests/schema -run '^TestSchema(Definitions|Values|References|Lifecycle)$' -count=1 -v
```

Expected: the harness applies the schema track using golang-migrate and schema_migrations_schema. Fresh setup has six business tables and zero business records. Metadata matches [data-model.md](data-model.md). Synthetic writes prove the exact length/null/default/numeric boundaries, quantity checks, current timestamp defaults and fixed-precision money.

Reference cases must show orphan writes and historical product deletion/id change failures with unchanged rows. Include the mixed-reference cases, successful Q1 cascades, last-order-reference removal and permitted ordinary product updates. Inspect orders.prodid length45 while only matching parent ids of length<=30 can succeed. Keep signed BIGINT endpoint fixtures separate from generated-id tests.

The lifecycle case installs an unrelated sentinel through its own test fixture migration, populates synthetic records, performs bounded schema down/up, and proves identical definitions and unchanged sentinel data. It also proves integrity remains enabled after recreation. No original seed/image data is imported. The schema tests clean up only their declared objects and run serially for a shared schema; Go's race check does not provide database isolation.

## Validate the independent credential paths

```sh
go test -race ./tests/schema -run '^TestCredential(Upgrade|Finalization)$' -count=1 -v
```

The upgrade test uses a separate disposable legacy fixture defined by versioned SQL from the verified legacy user columns. Its synthetic accounts have NULL legacy passwords. The credentials track preserves profiles/identities/counts, removes password and initializes NULL hash/reset TRUE. It does not traverse bootstrap or claim full conversion of legacy field definitions. Its deliberate down refusal must preserve all protected fields/data and surface any dirty history.

The finalization test uses separately isolated states. Verify refusal with a missing hash, with a TRUE flag despite a hash, and with both pending. Verify readiness only after every synthetic account has a non-null encoded hash and false reset flag. The test quiesces writers and exercises both the harness preflight and SQL guard. Success changes hash nullability only, retains values and cleans the guard routine; subsequent NULL/omitted hashes fail. Finalization down restores nullability without removing hashes or recreating password.

These tests simulate account state with synthetic DML; they do not implement sign-in, password hashing or reset endpoints. Use fresh instances/fixtures after deliberately failed migration attempts so dirty metadata cannot contaminate later cases.

## Complete the implementation gates

Check formatting with no output from the first two commands:

```sh
gofmt -l internal tests/schema
goimports -l internal tests/schema
go vet ./...
golangci-lint run
go test -race ./... -count=1
```

CI must fail if either formatting command lists files or if any other check exits nonzero. Test every case in [schema-contract.md](contracts/schema-contract.md); retain results grouped by US1/US2/US3 and SC-001–005, including expected SQL error categories, before/after rows and environment characterization of fractional money/collation behavior. Missing or unexecuted cases remain visibly incomplete. Do not interpret a local green run as platform acceptance.

## Failure handling and cleanup

On a migration error, stop and inspect its exact source/version, owned objects and dirty status. MySQL DDL is not assumed to undo previous statements. Never automatically force history, disable FKs or rebuild an existing legacy account database. The [migration contract](contracts/migration-contract.md) defines the refused credential downgrade and scoped forward-repair boundary.

After retaining diagnostics, discard only the disposable validation instance:

```sh
docker compose -f compose.schema.yaml down --volumes
```

The command removes this validation stack's synthetic database volume. It is not a production rollback, legacy data migration, push, merge or deployment.

## Changelog

- 2026-11-10 — Added C1's development-stage Go refusal red/green workflow for the java to go harness, including delayed bootstrap/credential-entry preflights; preserved SQL and finalization gates. Commands and results remain future implementation work.

- 2026-11-10 — Added the future runnable acceptance-first guide for the java to go schema foundation, work unit run-89ffa3642d6a4d44-wu. All database/build/lint results remain unexecuted during planning.
