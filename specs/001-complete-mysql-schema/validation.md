# Validation: Complete MySQL Schema

**Date**: 2026-11-16 | **Goal**: java to go

This file records implementation evidence for the Java-to-Go MySQL schema foundation. T004 adds only callable, non-enforcing configuration and environment scaffolding, so the configuration/environment refusal cases are expected to be red until T005.

## T004 Scaffolding Red Gate

Changelog:

- 2026-11-16 — Added typed configuration scaffolding in `internal/config`, read-only schema test helpers, startup loading in `tests/schema`, and table-driven configuration/environment preflight tests. Configuration and environment refusal entry points intentionally return success in T004.

Commands attempted:

```sh
go get github.com/go-sql-driver/mysql@v1.9.3
```

Exit status: `127`.

Result: the local `go` executable is not installed in this workspace, so the command could not resolve module metadata. The MySQL driver requirement was instead pinned manually to `github.com/go-sql-driver/mysql v1.10.1` with checksums from the public Go checksum database. This tool error does not establish a missing refusal.

Commands still required with a Go toolchain and valid disposable MySQL runtime:

```sh
go test -race ./internal/config -run '^TestConfig$' -count=1 -v
go test -race ./tests/schema -run '^TestEnvironmentReadOnlyConnection$' -count=1 -v
go test -race ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
```

Commands attempted:

```sh
go test -race ./internal/config -run '^TestConfig$' -count=1 -v
go test -race ./tests/schema -run '^TestEnvironmentReadOnlyConnection$' -count=1 -v
go test -race ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
docker compose -f compose.schema.yaml ps
```

Exit status for each command: `127`.

Result: `/bin/bash` reported `go: command not found` for each Go test command and `docker: command not found` for the Compose command. These are parse/tool/runtime errors and do not establish the missing refusal. The tests now contain the assertions that should fail once a Go toolchain and disposable MySQL runtime are available.

Temporary local Go 1.24.0 toolchain setup:

```sh
mkdir -p .tools/go && curl -fsSL https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -xz -C .tools/go --strip-components=1 && .tools/go/bin/go version
```

Exit status: `0`.

Output summary: `go version go1.24.0 linux/amd64`.

Race commands attempted with `GOROOT` pinned to the local toolchain:

```sh
CGO_ENABLED=1 GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test -race ./internal/config -run '^TestConfig$' -count=1 -v
CGO_ENABLED=1 GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test -race ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
CGO_ENABLED=1 GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test -race ./tests/schema -run '^TestEnvironmentReadOnlyConnection$' -count=1 -v
```

Exit status: interrupted after each reported the same tool prerequisite failure.

Output summary: `cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH`. This race-toolchain failure does not establish the missing refusal.

Non-race assertion checks:

```sh
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config -run '^TestConfig$' -count=1 -v
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentReadOnlyConnection$' -count=1 -v
```

Exit status for each command: `1`.

Assertion results:

- `TestConfig/valid_shopping_cart_dsn` passed.
- `TestConfig/missing_dsn_refused` failed with `expected configuration refusal, got nil`.
- `TestConfig/wrong_namespace_refused` failed with `expected configuration refusal, got nil`.
- `TestEnvironmentPreflight/valid_runtime_settings` passed.
- `TestEnvironmentPreflight/wrong_namespace_refused` failed with `expected environment refusal, got nil`.
- `TestEnvironmentPreflight/unsupported_server_refused` failed with `expected environment refusal, got nil`.
- `TestEnvironmentPreflight/missing_strict_settings_refused` failed with `expected environment refusal, got nil`.
- `TestEnvironmentPreflight/disabled_foreign_keys_refused` failed with `expected environment refusal, got nil`.
- `TestEnvironmentReadOnlyConnection` failed with `open read-only connection: dial tcp 127.0.0.1:3306: connect: connection refused` because no disposable MySQL container is available in this runtime. This connection failure does not establish a missing refusal.

Expected T004 outcome:

- `TestEnvironmentReadOnlyConnection` establishes actual read-only container connectivity when `SCHEMA_TEST_DSN` points at the disposable `shopping-cart` database.
- `TestConfig` fails assertions for the missing DSN and wrong namespace cases because `config.Validate` is callable but non-enforcing.
- `TestEnvironmentPreflight` fails assertions for wrong namespace, unsupported server version, missing strict SQL settings and disabled foreign keys because `validateEnvironment` is callable but non-enforcing.

No migration mutation, business DDL, fixture DDL, occupied-bootstrap preflight or credential-entry preflight was invoked for T004.

## T005 Configuration Refusals and Migration Scaffolding

Changelog:

- 2026-11-16 — Implemented required/invalid configuration refusals and environment preflight refusals after the recorded T004 red gate. Added redacted DSN diagnostics, Config-passed connection helpers, golang-migrate v4 file/MySQL scaffolding with explicit schema, credential, finalization, legacy-fixture and sentinel-fixture histories, selected version targets, owned-object cleanup, metadata/row snapshots and SQL code/SQLSTATE assertion support. Occupied-bootstrap and credential-entry checks remain callable and non-enforcing for their later red gates; no finalization preflight was implemented.

Local implementation sanity command:

```sh
SCHEMA_TEST_DSN='schema_test:secret@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config ./tests/schema -run '^(TestConfig|TestEnvironmentPreflight)$' -count=1 -v
```

Exit status: `0`.

Result summary: `TestConfig` passed valid, missing-DSN and wrong-namespace cases. `TestEnvironmentPreflight` passed valid, wrong-namespace, unsupported-server, missing-strict-settings and disabled-foreign-key cases using synthetic snapshots. This is T005 implementation sanity evidence only; T006 still owns the required green evidence record and actual read-only MySQL environment baseline.

Required race command attempted:

```sh
SCHEMA_TEST_DSN='schema_test:secret@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' CGO_ENABLED=1 GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test -race ./internal/config ./tests/schema -run '^(TestConfig|TestEnvironmentPreflight)$' -count=1 -v
```

Exit status: `1`.

Result: `cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH`. This runtime/toolchain prerequisite failure is not a test assertion failure and does not establish or disprove MySQL environment readiness.

## T006 Configuration/Environment Green Gate Attempt

Changelog:

- 2026-11-16 — Added the read-only `TestEnvironment` baseline against the selected runtime database and migration connection. The baseline asserts MySQL 8.4, `shopping-cart`, InnoDB support, utf8mb4/utf8mb4_0900_bin, required SQL modes, UTC and `foreign_key_checks=1` across newly acquired SQL and migration connections. Added `TestEnvironmentRefusesInvalidBeforeMigration` to prove invalid migration configuration is refused before migration invocation and compare business/history snapshots before and after the refusal.

Passing commands:

```sh
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config -run '^TestConfig$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
```

Exit status for each command: `0`.

Result summary: `TestConfig` passed valid, missing-DSN and wrong-namespace cases. `TestEnvironmentPreflight` passed valid runtime settings plus wrong namespace, unsupported server, missing strict settings and disabled foreign keys. These green results are retained beside T004's red evidence.

Actual-container baseline command:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
```

Exit status: `1`.

Result: `TestEnvironment` failed with `open read-only connection: open configured connection schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?...: dial tcp 127.0.0.1:3307: connect: connection refused`. This is the required fail-not-skip behavior for a missing actual MySQL prerequisite; it is not green T006 evidence.

Invalid-before-migration command:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentRefusesInvalidBeforeMigration$' -count=1 -v
```

Exit status: `1`.

Result: `TestEnvironmentRefusesInvalidBeforeMigration` failed at baseline connection setup with `dial tcp 127.0.0.1:3307: connect: connection refused`. Because no actual disposable MySQL container is reachable, the test cannot yet snapshot unchanged business objects/history around the invalid-input refusal.

Prerequisite checks:

```sh
docker compose -f compose.schema.yaml ps
which docker
which podman
which nerdctl
which mysql
ss -ltnp | grep 3307
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' CGO_ENABLED=1 GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test -race ./internal/config ./tests/schema -run '^(TestConfig|TestEnvironmentPreflight|TestEnvironment|TestEnvironmentRefusesInvalidBeforeMigration)$' -count=1 -v
```

Exit status: `docker compose` returned `127`; the runtime-discovery commands found no Docker, Podman, nerdctl, mysql client or listener on port 3307; the race command returned `1`.

Result: this workspace currently lacks the actual MySQL container/runtime and the C compiler required by `go test -race`. T006 is therefore not complete: all three required tests are not green, no domain test failure may be used as evidence, and `tasks.md` remains unchecked for T006.

Fresh T006 rerun in this turn:

```sh
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config -run '^TestConfig$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^(TestEnvironment|TestEnvironmentRefusesInvalidBeforeMigration)$' -count=1 -v
```

Exit status: `0`, `0`, and `1`.

Result summary: `TestConfig` and `TestEnvironmentPreflight` remain green. `TestEnvironment` and `TestEnvironmentRefusesInvalidBeforeMigration` both fail before any migration invocation because `127.0.0.1:3307` refuses the configured MySQL connection. This continues to demonstrate fail-not-skip behavior for the missing actual-container prerequisite, but it is not green T006 evidence.

Fresh T006 rerun after the speckit-implement invocation:

```sh
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config -run '^TestConfig$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^(TestEnvironment|TestEnvironmentRefusesInvalidBeforeMigration)$' -count=1 -v
```

Exit status: `0`, `0`, and `1`.

Result summary: `TestConfig` passed all three cases. `TestEnvironmentPreflight` passed the valid runtime, wrong namespace, unsupported server, missing strict settings and disabled foreign key cases. `TestEnvironment` and `TestEnvironmentRefusesInvalidBeforeMigration` failed before migration invocation with `dial tcp 127.0.0.1:3307: connect: connection refused`, so the actual-container baseline and unchanged business/history proof are still not green.

Runtime prerequisite check:

```sh
command -v go
command -v gcc
command -v docker
command -v podman
command -v nerdctl
command -v mysql
command -v mysqld
command -v mariadbd

## T016 Occupied Bootstrap Preflight and Initial Product/User SQL

Changelog:

- 2026-11-17 — Implemented the occupied-bootstrap inventory preflight in `tests/schema/helpers_test.go`. The check now inspects the six owned business tables (`product`, `user`, `orders`, `usercart`, `user_demand`, `transactions`) and refuses a schema bootstrap when any are already present, without mutating business objects or migration history. Added `migrations/schema/000001_complete_mysql_schema.up.sql` with only the T016-authorized `product` and `user` tables, using `CREATE TABLE` without `IF NOT EXISTS`, explicit InnoDB/utf8mb4/utf8mb4_0900_bin, signed native numerics, protected `password_hash`/`must_reset_password` columns and no seed, legacy `password` column, namespace drop or extra policy.

Static and synthetic checks:

```sh
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" gofmt -w tests/schema/helpers_test.go
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config -run '^TestConfig$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
```

Exit status: `0`, `0` and `0`.

Result summary: `TestConfig` passed valid, missing-DSN and wrong-namespace cases. `TestEnvironmentPreflight` passed valid runtime settings plus wrong namespace, unsupported server, missing strict settings and disabled foreign-key cases. `gofmt -l tests/schema/helpers_test.go` produced no output after formatting.

Required occupied-bootstrap refusal command attempted:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestSchemaLifecycle/(occupied inventory refused before migration|legacy user fixture directly refuses occupied bootstrap)$' -count=1 -v
```

Exit status: `1`.

Result: `TestSchemaLifecycle` failed before either occupied-bootstrap subtest could execute because `127.0.0.1:3307` refused the configured MySQL connection. No schema migration was invoked by this failed local attempt, and no business objects/history snapshot could be established in this workspace. This is the same missing disposable MySQL prerequisite recorded in T006; it is not green T011 refusal evidence.
ss -ltnp
```

Exit status: `0` for the combined shell command.

Result: none of the runtime commands produced a path, and no listening TCP sockets were reported. The workspace still has the local `.tools/go` binary, but no container runtime, MySQL server/client or C compiler was found. T006 remains incomplete and must not be checked off until the actual MySQL 8.4 container baseline and invalid-before-migration snapshot test pass.

Current T006 rerun from the requested speckit-implement pass:

```sh
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config -run '^TestConfig$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^(TestEnvironment|TestEnvironmentRefusesInvalidBeforeMigration)$' -count=1 -v
```

Exit status: `0`, `0`, and `1`.

Result summary: `TestConfig` passed all valid, missing-DSN and wrong-namespace cases. `TestEnvironmentPreflight` passed the valid runtime, wrong namespace, unsupported server, missing strict settings and disabled foreign key cases. `TestEnvironment` and `TestEnvironmentRefusesInvalidBeforeMigration` failed before any migration invocation because `127.0.0.1:3307` refused the configured MySQL connection. This is still fail-not-skip behavior for the missing actual-container prerequisite, not green T006 evidence.

Container/runtime availability check:

```sh
id
command -v go
command -v gcc
command -v cc
command -v docker
command -v podman
command -v nerdctl
command -v mysql
command -v mysqld
command -v mariadbd
apt-get update
```

Exit status: command discovery found only the unprivileged `worker` user context and no runtime paths; `apt-get update` returned `100`.

Result summary: no Docker-compatible container runtime, MySQL server/client, system Go toolchain or C compiler is available in `PATH`. `apt-get update` cannot refresh package metadata because the unprivileged user cannot write `/var/lib/apt/lists/partial`. The actual-container requirement remains unmet, so T006 is intentionally left unchecked.

Latest T006 rerun from the requested speckit-implement pass:

```sh
GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config -run '^TestConfig$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironmentPreflight$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^(TestEnvironment|TestEnvironmentRefusesInvalidBeforeMigration)$' -count=1 -v
command -v docker
command -v podman
command -v nerdctl
command -v mysql
command -v mysqld
command -v mariadbd
command -v gcc
ss -ltnp
```

Exit status: `0`, `0`, `1`, and `0` for the runtime discovery command.

Result summary: `TestConfig` passed all valid, missing-DSN and wrong-namespace cases. `TestEnvironmentPreflight` passed the valid runtime, wrong namespace, unsupported server, missing strict settings and disabled foreign key cases. `TestEnvironment` and `TestEnvironmentRefusesInvalidBeforeMigration` failed before any migration invocation with `dial tcp 127.0.0.1:3307: connect: connection refused`. Runtime discovery produced no Docker-compatible container runtime, MySQL server/client, C compiler or listener on port 3307. This preserves fail-not-skip evidence for missing prerequisites, but it is not green T006 evidence; T006 remains unchecked and no domain failure may be used as evidence.

Completed T006 green evidence from this pass:

Changelog:

- 2026-11-16 — Completed the T006 configuration/environment green gate for the Java-to-Go schema foundation. The environment baseline runs against a disposable MySQL 8.4.2 server on the configured `127.0.0.1:3307` test port, with the `shopping-cart` namespace, InnoDB support, utf8mb4/utf8mb4_0900_bin, required strict modes, UTC and `foreign_key_checks=1` verified on newly acquired SQL and migration-style connections. The invalid-input case refuses before migration invocation and preserves business/history snapshots.

Local runtime setup used for the baseline:

```sh
.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqld --version
```

Output summary: `mysqld  Ver 8.4.2 for Linux on x86_64 (MySQL Community Server - GPL)`.

Environment characterization query after applying the same session settings as the harness:

```sh
SET SESSION sql_mode='STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO', time_zone='+00:00', foreign_key_checks=1;
SET NAMES utf8mb4 COLLATE utf8mb4_0900_bin;
SELECT VERSION(), DATABASE(), @@session.sql_mode, @@session.time_zone, @@session.foreign_key_checks, @@character_set_connection, @@collation_connection;
SELECT SUPPORT FROM INFORMATION_SCHEMA.ENGINES WHERE ENGINE='InnoDB';
```

Output summary: `8.4.2`, `shopping-cart`, `NO_AUTO_VALUE_ON_ZERO,STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION`, `+00:00`, `1`, `utf8mb4`, `utf8mb4_0900_bin`, and InnoDB support `DEFAULT`.

Code adjustment made for the green baseline:

- `tests/schema/helpers_test.go` sets the required `utf8mb4_0900_bin` collation through `SET NAMES` instead of the Go driver's DSN collation field, because `go-sql-driver/mysql` rejects that server collation name before the session can be inspected.
- `tests/schema/environment_test.go` inspects a newly acquired migration-style database connection without requiring unauthored migration source files.
- Snapshot helpers collect metadata rows before issuing row-count queries, avoiding self-blocking on the single configured connection.

Passing T006 command:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./internal/config ./tests/schema -run '^(TestConfig|TestEnvironmentPreflight|TestEnvironment|TestEnvironmentRefusesInvalidBeforeMigration)$' -count=1 -v
```

Exit status: `0`.

Result summary:

- `TestConfig` passed valid, missing-DSN and wrong-namespace cases.
- `TestEnvironmentPreflight` passed valid runtime settings plus wrong namespace, unsupported server, missing strict settings and disabled foreign key refusal cases.
- `TestEnvironment` passed against the live MySQL 8.4.2 runtime, asserting the selected namespace, InnoDB support, utf8mb4/utf8mb4_0900_bin, required SQL modes, UTC and enabled foreign keys across newly acquired SQL and migration-style connections.
- `TestEnvironmentRefusesInvalidBeforeMigration` passed, proving invalid migration configuration is refused before migration invocation and leaves business objects/history unchanged.

No domain test failure has been used as evidence before these three T006 checks turned green.

## T010 Shared Schema Contract Red Gate

Changelog:

- 2026-11-17 — Ran the shared definition, value and reference suites for the Java-to-Go MySQL foundation before any complete-schema migration file exists. The harness inspected the empty provisioned `shopping-cart` baseline because `migrations/schema/` contains only `.gitkeep`, so the failures identify absent required schema objects/behavior rather than missing source files. Preserve this red evidence beside the later green runs.

Precondition checks:

```sh
find migrations -maxdepth 3 -type f -print | sort
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" .tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
```

Exit status: first command `0`; second command `0`.

Result summary: `migrations/schema/` contains no complete-schema migration source, only `.gitkeep`. The local disposable MySQL 8.4.2 baseline is reachable on `127.0.0.1:3307`; the password shown above is redacted.

Combined red-gate command:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestSchema(Definitions|Values|References)$' -count=1 -v
```

Exit status: `1`.

Assertion failures:

- `TestSchemaDefinitions` failed with `business tables = [], want exactly [orders product transactions user user_demand usercart]`.
- `TestSchemaReferences` failed with `reset schema values with "DELETE FROM transactions": Error 1146 (42S02): Table 'shopping-cart.transactions' doesn't exist`.
- `TestSchemaValues` failed with `reset schema values with "DELETE FROM transactions": Error 1146 (42S02): Table 'shopping-cart.transactions' doesn't exist`.

Per-suite commands:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestSchemaDefinitions$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestSchemaValues$' -count=1 -v
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestSchemaReferences$' -count=1 -v
```

Exit status for each command: `1`.

Per-suite assertion failures:

- `TestSchemaDefinitions`: `definitions_test.go:43: business tables = [], want exactly [orders product transactions user user_demand usercart]`.
- `TestSchemaValues`: `values_test.go:33: reset schema values with "DELETE FROM transactions": Error 1146 (42S02): Table 'shopping-cart.transactions' doesn't exist`.
- `TestSchemaReferences`: `references_test.go:33: reset schema values with "DELETE FROM transactions": Error 1146 (42S02): Table 'shopping-cart.transactions' doesn't exist`.

Coverage by story:

| Story | Shared suite coverage | Red-gate status |
| --- | --- | --- |
| US1 | Complete six-table inventory, empty business tables, metadata, keys, indexes, FKs, checks, engine/charset/collation and absence of obsolete/extra business objects through `TestSchemaDefinitions`. | Expected red: all required business tables are absent from the empty provisioned baseline. |
| US2 | Required/optional values, length limits, binary image, money, signed numeric boundaries, generated ids, quantity checks, timestamps, demand duplicates and hash capacity through `TestSchemaValues`. | Expected red: the first required target table is absent, so value behavior cannot exist yet. |
| US3 | Duplicate identities/pairs, orphan refusals, historical product protection, Q1 cascade/restrict actions, mixed-reference atomicity and reconstruction enforcement through `TestSchemaReferences`. | Expected red: the first required target table is absent, so reference behavior cannot exist yet. |

This T010 evidence is not a compilation, connection, missing-tool or missing-source-file failure. The tests compiled, connected to the disposable MySQL baseline and reached assertions against absent required objects/behavior before the complete-schema DDL was authored.

## T015 Lifecycle/Credential Entry Red Gate

Changelog:

- 2026-11-17 — Applied the legacy user fixture and sentinel version 1 through golang-migrate while target schema and credential migrations remain unauthored. Recorded the direct occupied-bootstrap and credential-entry preflight reds before implementing those checks. The direct reds do not invoke target migration mutation or inspect unauthored target sources; they fail only because `checkOccupiedBootstrap` and `checkCredentialEntry` still accept invalid entry states.

Working fixture/environment checks:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" .tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
find migrations tests/schema/testdata -maxdepth 3 -type f -print | sort
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
```

Exit status: `0`, `0`, and `0`.

Result summary: the disposable MySQL 8.4.2 runtime was reachable on `127.0.0.1:3307`; fixture files exist for legacy version 1 and sentinel versions 1 and 2; `migrations/schema`, `migrations/credentials` and `migrations/credentials-finalize` still contain only `.gitkeep`; `TestEnvironment` passed against `shopping-cart` with the required session settings. Sentinel version 2 was not applied because its external reference depends on product, which is not authored yet.

Direct occupied-bootstrap preflight red:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestSchemaLifecycle/legacy_user_fixture_directly_refuses_occupied_bootstrap$' -count=1 -v
```

Exit status: `1`.

Result summary: the legacy fixture applied through golang-migrate using `schema_migrations_fixture_legacy`, creating the synthetic legacy `user` table with two rows. The direct occupied-bootstrap check did not invoke the schema migration source and left the preflight snapshot unchanged. The expected red assertion is `expected legacy user fixture to refuse occupied bootstrap before migration invocation`, because `checkOccupiedBootstrap` is still non-enforcing.

Direct credential-entry preflight red:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestCredentialUpgrade/empty_fixture_refuses_credential_entry_before_migration$' -count=1 -v
```

Exit status: `1`.

Result summary: the empty fixture state had no business tables or credential history before the check and remained unchanged afterward. The direct credential-entry check did not invoke the credentials migration source. The expected red assertion is `expected missing legacy user to refuse credential entry before migration invocation`, because `checkCredentialEntry` is still non-enforcing.

Lifecycle SQL red retained:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior$' -count=1 -v
```

Exit status: `1`.

Result summary: sentinel version 1 applied through golang-migrate before the schema lifecycle case, proving the unrelated fixture path and history work. The case then failed at `apply schema before populated lifecycle` with `no migration found for version 1` because the complete-schema migration is not authored until T016-T019. This retains the lifecycle SQL red gate as blocked by absent target SQL, but this missing source lookup is not counted as the expected entry-preflight refusal.

Credential-upgrade SQL red retained:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' GOROOT="$PWD/.tools/go" PATH="$PWD/.tools/go/bin:$PATH" go test ./tests/schema -run '^TestCredentialUpgrade/legacy_only_upgrade_preserves_profile_and_initializes_protected_target_columns$' -count=1 -v
```

Exit status: `1`.

Result summary: legacy fixture version 1 and sentinel version 1 applied through golang-migrate before the credential-upgrade case. The still-non-enforcing credential-entry check accepted the legacy-only entry state as expected for T015, and the case then failed at `upgrade legacy credentials` with `no migration found for version 1` because the credential migration is not authored until T020-T021. This retains the upgrade SQL red gate, but this missing source lookup is not counted as the expected entry-preflight refusal.

Harness adjustments made only to establish T015 evidence:

- `tests/schema/helpers_test.go` resolves migration source directories from the repository root when Go runs package tests from `tests/schema`, and enables multi-statement migration execution so versioned fixture SQL can create tables and insert synthetic rows through golang-migrate.
- `tests/schema/lifecycle_test.go` and `tests/schema/credentials_test.go` keep the two direct preflight red cases from using migration version lookups after the preflight call, because those lookups create empty history tables and would mutate the state being compared.

T016 subsequently implements the occupied-bootstrap preflight and product/user SQL; see the T016 evidence above. The local T016 rerun could not establish green occupied-bootstrap refusal evidence because this workspace no longer has a reachable disposable MySQL listener. T020 remains blocked until the credential-entry red above is retained beside its later green refusal and the additional initial-target red required after T016-T019. No migration mutation, unauthored source lookup, or business-dependent external-reference fixture application is used as the expected direct preflight failure.

## T020 Credential-Entry Preflight and Legacy Password Replacement

Changelog:

- 2026-11-23 — Implemented the legacy-only credential-entry preflight in `tests/schema/helpers_test.go`: `checkCredentialEntry` now inspects the `user` table through `INFORMATION_SCHEMA` and refuses credential entry when the `user` table is absent, when it already carries `password_hash` or `must_reset_password` target columns, or when the legacy `password` column is missing. The preflight never invokes migration mutation and leaves business objects and credential history unchanged. Added `migrations/credentials/000001_replace_legacy_password.up.sql` as a single `ALTER TABLE user` that drops the legacy `password` column and adds `password_hash VARCHAR(255) NULL DEFAULT NULL` plus `must_reset_password BOOLEAN NOT NULL DEFAULT TRUE`; existing rows are initialized exclusively through those column defaults, legacy profile columns (`email`, `name`, `mobile`, `address`, `pincode`) and their types are retained verbatim, and no SQL in the migration reads, copies, hashes or otherwise migrates the discarded plaintext password contents.

Initial-target credential-entry red gate rerun on the schema authored by T016–T019, recorded before implementing the preflight:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestCredentialUpgrade/initial_target_schema_refuses_credential_entry_before_migration$' -count=1 -v
```

Exit status: `1`.

Result summary: the complete-schema migration file `migrations/schema/000001_complete_mysql_schema.up.sql` now exists, so `schemaMigrationAuthored` returns true and the `initial target schema refuses credential entry before migration` subtest is no longer skipped. In this workspace the runtime disposable MySQL listener on `127.0.0.1:3307` is unavailable, so the test harness opened the shared credential-upgrade connection and failed fast with `dial tcp 127.0.0.1:3307: connect: connection refused` before invoking any migration mutation. This is the same fail-not-skip boundary recorded in T006/T010/T015, not the T020 missing-refusal assertion itself; the still-non-enforcing `checkCredentialEntry` return path was the underlying defect being recorded. No credential, schema or fixture migration source was invoked by this run.

Preflight implementation and migration file:

- `tests/schema/helpers_test.go` — `checkCredentialEntry` returns a refusal for: missing `user` table (`credential entry refused: legacy user table is absent`), present-target columns (`credential entry refused: user table already carries target credential columns: must_reset_password, password_hash`), and missing legacy column (`credential entry refused: user table missing legacy password column`). The accepting path only returns `nil` when the `user` table carries the legacy `password` column and neither target column.
- `migrations/credentials/000001_replace_legacy_password.up.sql` — one statement, no `DELIMITER` directives, no seed, no secondary table, no read of the discarded column before `DROP COLUMN password`.

Static verification of the changed files:

```sh
gofmt -l tests/schema/helpers_test.go
go vet ./...
go test -c ./tests/schema -o /tmp/schematest.bin
```

Exit status for each command: `0`.

Result summary: `gofmt -l` produced no output for the updated helper file; `go vet ./...` reported no findings against the whole module; the credentials package test binary compiled without errors using the local `go1.25.14` toolchain (the Go 1.24 baseline pin is still the required execution environment per T001). No MySQL runtime was exercised by these static checks.

Follow-up credential-entry refusal rerun after implementing the preflight, intended to retain green evidence beside T015:

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestCredentialUpgrade$' -count=1 -v
```

Exit status: `1`.

Result summary: the shared credential-upgrade connection could not be opened in this workspace (`dial tcp 127.0.0.1:3307: connect: connection refused`). The `empty fixture refuses credential entry before migration`, `incompatible entry refuses before migration` and `initial target schema refuses credential entry before migration` subtests therefore never executed, so the green refusal evidence required beside T015's red cannot be recorded in this workspace. This is the same missing-container prerequisite captured in T006/T010/T015/T016; it is not a successful green run, no business objects or credential history were snapshotted, and no credentials or schema migration mutation was invoked. The implementation changes required by T020 are complete, but local green evidence must be re-recorded on an environment that has the disposable MySQL 8.4 listener available. No migration mutation, forced history or plaintext password access is used by `checkCredentialEntry` or by `migrations/credentials/000001_replace_legacy_password.up.sql`.
