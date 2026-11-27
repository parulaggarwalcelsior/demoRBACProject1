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

## T022 US1 Green Verification Under Race Detection

Changelog:

- 2026-11-24 — Ran `TestSchemaDefinitions`, `TestSchemaLifecycle` and `TestCredentialUpgrade` with `-race` on isolated disposable MySQL 8.4.2 fixtures (one `shopping-cart` namespace, reset between tests through `cleanupOwnedObjects` plus history-table drops). Fixed three test-harness infrastructure defects that otherwise prevented T022 from verifying the required semantics: `schemaMigrationAuthored` now resolves the complete-schema source through the same walk-up path as `migrationSourcePath`, `migrateToVersion(track, 0)` now invokes `migrator.Down()` so the single-version schema and credentials tracks actually apply their `down.sql`, and `migrationVersion` reads `schema_migrations_*` directly so inspecting dirty state no longer creates the history table as a side effect and no longer loses the nil-version dirty flag. The structural comparison in `TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior` excludes row counts (bounded down discards synthetic rows by contract) and `CreateSQL` is normalized to strip `auto_increment=<N>` so the same structural definitions are compared before and after. A `TestSchemaDefinitions` fixture reset at the start of the test keeps the three suites independent when executed back-to-back. The schema up migration keeps all six FK contracts and lets MySQL auto-create the FK-backing index with the FK's name (`productid`, `fk_usercart_pid`, `fk_user_demand_email`, `fk_transactions_email`, `fk_transactions_pid`) instead of a redundant `KEY idx_*` override, matching the T007 definition contract.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
```

Exit status: `0` and `0`.

Result summary: the disposable MySQL 8.4.2 listener on `127.0.0.1:3307` responded `mysqld is alive`, and `TestEnvironment` passed against `shopping-cart` with InnoDB support, utf8mb4/utf8mb4_0900_bin, required strict modes, UTC and `foreign_key_checks=1`. Passwords in any DSN above are redacted.

Race commands (each isolated by the shared cleanup helper documented in the changelog):

```sh
# Between every command: DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, `user`, sentinel_owned,
#   schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize,
#   schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel;
#   DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials;
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaDefinitions$' -count=1 -v
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaLifecycle$' -count=1 -v
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestCredentialUpgrade$' -count=1 -v
```

Exit status for each command: `0`.

Race-run results:

- `TestSchemaDefinitions`: `--- PASS: TestSchemaDefinitions (0.49s)`. After the schema track applies version 1, the inspection reports exactly the six business tables `orders, product, transactions, user, user_demand, usercart`, zero initial business rows, every column type/nullability/default from data-model.md (including `password_hash VARCHAR(255) NULL DEFAULT NULL`, `must_reset_password TINYINT(1) NOT NULL DEFAULT 1`, `pprice/amount DECIMAL(10,2) NOT NULL DEFAULT 0.00` and `product.pid VARCHAR(30)` vs `orders.prodid VARCHAR(45)`), the correct primary keys (`product.pid`, `user.email`, composite `(orderid, prodid)` for orders, `id` for the generated-key tables), every unique index (`uq_usercart_email_pid`, `uq_transactions_transaction_id`), every FK-backing index now named after its FK (`productid`, `fk_usercart_pid`, `fk_user_demand_email`, `fk_transactions_email`, `fk_transactions_pid`), all six named FK contracts with Q1 actions (`productid` NO ACTION/NO ACTION; `fk_usercart_email`/`fk_user_demand_email` CASCADE/CASCADE; `fk_usercart_pid`/`fk_transactions_email`/`fk_transactions_pid` CASCADE on update, RESTRICT on delete), both enforced `>=1` quantity checks (`ck_usercart_quantity`, `ck_transactions_quantity`), and `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin` on every business table. No obsolete `books`, `order`, `order_header`, `order_item` or legacy `users` relation was observed.
- `TestSchemaLifecycle`: `--- PASS: TestSchemaLifecycle (3.63s)` with all five subtests `PASS`:
  - `initial up and already applied no change` — initial `up` succeeds, `assertTrackVersion` records `schema=1 dirty=false` and `credentials-finalize=0 dirty=false`, re-running up is a no-change (`ErrNoChange` swallowed), the inspected definition is identical byte-for-byte between the first and second `up`, and the business/history snapshot differs from the empty baseline (no spurious tables).
  - `occupied inventory refused before migration` — a hand-created `product` table trips `checkOccupiedBootstrap` with `occupied bootstrap inventory: product`, migration is never invoked (`schema` track stays at `version=0 dirty=false`), and the direct-SQL `migrationVersion` leaves the before/after `snapshotBusinessAndHistory` objects and row counts identical.
  - `legacy user fixture directly refuses occupied bootstrap` — applying the legacy fixture seeds the synthetic `user` table only, `checkOccupiedBootstrap` refuses with `occupied bootstrap inventory: user`, and the before/after `snapshotBusinessAndHistory` snapshots are identical (no target-schema or credentials migration is invoked).
  - `populated bounded down up keeps sentinel and behavior` — sentinel version 1 seeds `sentinel_owned` with `unrelated-sentinel`, the complete schema up applies, synthetic `product/user/orders/usercart/user_demand/transactions` rows populate without FK/check errors, `migrateToVersion(schemaTrack, 0)` (now routed through `migrator.Down()`) drops the six business tables with the sentinel untouched (`schema=0 dirty=false`), re-up restores every structural metadata (`beforeStructure == afterStructure` after excluding row counts and auto-increment counters) with `assertBusinessTablesEmpty` confirming zero rows and the lifecycle-behavior re-insertions showing `foreign_key_checks=1` still rejects a missing-FK child and the quantity check still rejects zero.
  - `external reference teardown refusal leaves partial state for discard` — sentinel version 2 re-introduces an external FK to `product` after the schema is up, the populated `migrateToVersion(schemaTrack, 0)` fails (DROP TABLE product refused by the external FK), `assertTrackDirty` reports `schema dirty=true` through the direct `schema_migrations_schema` read, `snapshotOwnedObjects` shows partial DDL state with tracked tables still present, and the sentinel's seeded row is unchanged. The dirty fixture is discarded by the shared cleanup rather than force-advanced.
- `TestCredentialUpgrade`: `--- PASS: TestCredentialUpgrade (2.34s)` with all six subtests `PASS`:
  - `empty fixture refuses credential entry before migration` — `checkCredentialEntry` refuses with `credential entry refused: legacy user table is absent`; before/after `snapshotBusinessAndHistory` are identical and no credentials migration ran.
  - `legacy only upgrade preserves profile and initializes protected target columns` — `assertLegacyPasswordValuesAbsent` confirms the fixture never seeds plaintext, `checkCredentialEntry` accepts the legacy-only state, the `credentials` migration runs cleanly, the legacy `password` column is absent after upgrade, `password_hash VARCHAR(255) NULL DEFAULT NULL` and `must_reset_password TINYINT(1) NOT NULL DEFAULT 1` are present, `assertCredentialProfilesUnchanged`/`assertUserProfileDefinitionsUnchanged` show every retained profile value and definition unchanged, `assertInitialCredentialTargets` confirms every surviving row has `password_hash IS NULL` and `must_reset_password = TRUE`, and the unrelated sentinel is unchanged.
  - `incompatible entry refuses before migration` — a `user` table carrying the target `password_hash`/`must_reset_password` columns is refused with `credential entry refused: user table already carries target credential columns: must_reset_password, password_hash`; `credentials` track stays at `version=0 dirty=false` and the before/after `snapshotBusinessAndHistory` match.
  - `initial target schema refuses credential entry before migration` — `schemaMigrationAuthored()` now resolves the complete-schema source correctly so this subtest executes rather than skipping. After applying the initial target schema from `migrations/schema/000001_complete_mysql_schema.up.sql`, `checkCredentialEntry` refuses with the same target-columns message; `credentials` track stays at `version=0 dirty=false` and the business/history snapshot is unchanged.
  - `repeated cleanup leaves no credential owned objects` — two consecutive `resetCredentialUpgrade` calls after applying the legacy and sentinel fixtures leave zero business tables and `credentials=0 dirty=false`.
  - `credential down refuses and retains protected data with dirty state` — after upgrading and replacing one hash with `synthetic-protected-hash`/`must_reset_password=FALSE`, the explicit down (`migrator.Down()`) is refused by the migration's `SIGNAL SQLSTATE '45000'` diagnostic; `assertTrackDirty` reads `schema_migrations_credentials` directly and reports the truthful `version<0 dirty=true` state; `assertNoColumn` confirms the legacy `password` column has not been recreated; the stored `password_hash` and profile values remain exactly as written.

Coverage mapping:

| Requirement | Evidence |
| --- | --- |
| US1.1 | `TestSchemaDefinitions` plus `TestSchemaLifecycle/initial_up_and_already_applied_no_change` — complete inventory, metadata and zero business rows on an isolated empty destination. |
| US1.2 | `TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior` — bounded removal/recreation lifecycle, recreated definitions match initial definitions, synthetic rows discarded by contract. |
| US1.3 | `TestSchemaLifecycle/initial_up_and_already_applied_no_change` — complete executable schema, no unfinished statement, no business seed row present. |
| US1.4 | `TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior` and `TestSchemaLifecycle/external_reference_teardown_refusal_leaves_partial_state_for_discard` — unrelated sentinel object and its data remain unchanged through down/up and through the refused teardown. |
| US1.5 | `TestCredentialUpgrade/legacy_only_upgrade_preserves_profile_and_initializes_protected_target_columns` — legacy `password` dropped, `password_hash VARCHAR(255) NULL` and `must_reset_password BOOLEAN NOT NULL DEFAULT TRUE` added, every account/profile row preserved, every row initialized to NULL hash / TRUE reset without any plaintext read/convert. |
| SC-001 | `TestSchemaDefinitions` and `TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior` — 100% of the six in-scope business relations are available after initial setup and after the removal/recreation cycle, with zero business seeds and zero changes to unrelated objects. |
| SC-002 | `TestSchemaDefinitions` accounting of all 36 columns, six FKs, two checks and engine/collation; the test fails rather than passes on any missing or extra rule, so a green run attests that every covered rule has an explicit preservation. UNVERIFIED items outside this feature's scope stay recorded elsewhere in this file. |
| SC-005 | This section traces each US1 scenario and D1–D5 difference to a specific executed subtest and command, allowing a reviewer to distinguish confirmed behavior from UNVERIFIED items without inspecting application code. |
| D1 | `TestSchemaDefinitions`/`TestSchemaLifecycle/initial_up_and_already_applied_no_change` — the completed `migrations/schema/000001_complete_mysql_schema.up.sql` runs to the end without any truncated image seed and leaves zero business rows. |
| D2 | `TestCredentialUpgrade/legacy_only_upgrade_preserves_profile_and_initializes_protected_target_columns` and `TestCredentialUpgrade/credential_down_refuses_and_retains_protected_data_with_dirty_state` — plaintext `password` is dropped, never read; `password_hash VARCHAR(255)` initialized NULL with `must_reset_password` TRUE; the credential down refuses to restore plaintext storage and leaves protected data unchanged. |
| D3 | `TestSchemaDefinitions` — the Q1-supplied `usercart`, `user_demand` and `transactions` definitions (generated BIGINT ids, cart email/pid uniqueness, positive-quantity checks, `created_at/added_at/requested_at` CURRENT_TIMESTAMP defaults, Q1 FK actions) are present and enforced; no legacy demand `product`/`quantity` or `transid` columns remain. |
| D4 | `TestSchemaDefinitions` — every monetary field is `DECIMAL(10,2) NOT NULL DEFAULT 0.00`, every table is `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin`, and the `orders.productid` FK retains NO ACTION on delete/update. |
| D5 | `TestSchemaDefinitions` — `user.email VARCHAR(100)` and `product.pid VARCHAR(30)` parents, child references match (`usercart.email/pid`, `user_demand.email`, `transactions.email/pid`), the `uq_usercart_email_pid` unique index exists, and no username/barcode columns or users/books parent tables are present; `orders.prodid` remains `VARCHAR(45)`. |

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials;"
```

Exit status: `0`.

Result: the deliberately dirty fixture from the external-reference teardown and the dirty credentials down are discarded through DROP TABLE on the disposable `shopping-cart` namespace; the `schema_migrations_*` history tables are dropped so no forced `migrate.Force` or hand-advanced history is used. The underlying MySQL 8.4.2 server is itself disposable (local `.tools/mysql-data` under this workspace). No push, merge or deployment was performed.

### T022 Re-run Confirmation — 2026-11-24 speckit-implement pass

Changelog:

- 2026-11-24 — Re-ran all three required T022 suites with `-race` on the restarted local disposable MySQL 8.4.2 instance on `127.0.0.1:3307`, cleaning the `shopping-cart` namespace between suites so each is applied to an empty baseline. No source change was required: `TestSchemaDefinitions`, `TestSchemaLifecycle` and `TestCredentialUpgrade` all pass under race detection against the current `migrations/schema/000001_complete_mysql_schema.{up,down}.sql`, `migrations/credentials/000001_replace_legacy_password.{up,down}.sql`, fixture SQL and helper code. Dirty state from the deliberately refused teardown/credential-down cases was discarded through `DROP TABLE` on the disposable namespace; no `migrate.Force` was used.

Race commands (each separated by the shared cleanup documented in the earlier T022 changelog):

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaDefinitions$' -count=1 -v
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaLifecycle$' -count=1 -v
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestCredentialUpgrade$' -count=1 -v
```

Exit status for each command: `0`.

Race-run results:

- `TestSchemaDefinitions`: `--- PASS: TestSchemaDefinitions (0.44s)`. The six business tables, zero initial business rows, all metadata (columns, defaults, nullability, PKs, unique indexes, FK-backing indexes, six named FK contracts with Q1 actions, two enforced `>=1` quantity checks, InnoDB/utf8mb4/utf8mb4_0900_bin) match the data-model.md contract. US1.1, SC-001/002, D1, D3, D4 and D5 remain evidenced.
- `TestSchemaLifecycle`: `--- PASS: TestSchemaLifecycle (3.09s)` with every subtest PASS: `initial_up_and_already_applied_no_change` (clean history, no-change repeated up), `occupied_inventory_refused_before_migration` (refusal before invocation, snapshot unchanged), `legacy_user_fixture_directly_refuses_occupied_bootstrap` (refusal before invocation on legacy fixture, snapshot unchanged), `populated_bounded_down_up_keeps_sentinel_and_behavior` (down/up preservation of structural definitions and the unrelated sentinel), `external_reference_teardown_refusal_leaves_partial_state_for_discard` (truthful dirty state on refused down; disposable fixture discarded rather than forced). US1.1-4, SC-001/002/005 and D1 remain evidenced.
- `TestCredentialUpgrade`: `--- PASS: TestCredentialUpgrade (1.85s)` with every subtest PASS: `empty_fixture_refuses_credential_entry_before_migration`, `legacy_only_upgrade_preserves_profile_and_initializes_protected_target_columns` (upgrade preservation; no plaintext read), `incompatible_entry_refuses_before_migration`, `initial_target_schema_refuses_credential_entry_before_migration` (credential-entry refusal on initial target schema), `repeated_cleanup_leaves_no_credential_owned_objects`, `credential_down_refuses_and_retains_protected_data_with_dirty_state` (truthful refused-down dirty state). US1.5, SC-005 and D2 remain evidenced.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials;"
```

Exit status: `0`. The `shopping-cart` namespace is empty (`SHOW TABLES;` returns no rows), so the deliberately dirty fixture from the refused teardown and the dirty credentials down are discarded through `DROP TABLE` without forced history. No push, merge or deployment was performed.

## T024 Finalization Red Gate on Working Initial/Upgraded Fixtures

Changelog:

- 2026-11-24 — Ran `TestCredentialFinalization` on the working initial target schema produced by `migrations/schema/000001_complete_mysql_schema.up.sql` (T016–T019) before the finalization SQL/helper behavior exists. `migrations/credentials-finalize/` contains only `.gitkeep`, no `up.sql`/`down.sql`, no `schema_guard_finalize_credentials` routine source, and `checkFinalizationReadiness` in `tests/schema/credentials_test.go` still unconditionally returns `nil`. The shared value contract from T010 is already implemented and green (see the T022 section above); this T024 run deliberately does not re-execute `TestSchemaValues` as a new red, and the T010 red evidence recorded earlier in this file is retained unchanged.

Precondition checks:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
find migrations/credentials-finalize -maxdepth 2 -type f -print | sort
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
```

Exit status: `0`, `0`, and `0`.

Result summary: the disposable MySQL 8.4.2 listener on `127.0.0.1:3307` responded `mysqld is alive`; `migrations/credentials-finalize/` contains only `.gitkeep` (no `up.sql`/`down.sql`); and `TestEnvironment` passed against `shopping-cart` with the required strict modes, UTC, utf8mb4/utf8mb4_0900_bin and `foreign_key_checks=1`. Passwords are redacted above.

T024 race command on the working initial target schema fixture:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestCredentialFinalization$' -count=1 -v
```

Exit status: `1`.

Per-subtest red assertions (expected target assertions that cannot hold until the T025–T027 finalization SQL/helper exists):

- `harness_preflight_refuses_pending_null_hash_reset_true` FAIL — `credentials_test.go:381: expected harness preflight to refuse pending state "null_hash_reset_true" before migration invocation`. The still non-enforcing `checkFinalizationReadiness` accepts the NULL-hash/reset=TRUE account; business/history snapshots and the `password_hash` definition remain `varchar(255) NULL DEFAULT NULL`, so the ready schema still permits NULL hashes.
- `harness_preflight_refuses_pending_null_hash_reset_false` FAIL — same message for `null_hash_reset_false`. Same unchanged target assertion: the account's NULL hash is not rejected and the column stays nullable.
- `harness_preflight_refuses_pending_non_null_hash_reset_true` FAIL — same message for `non_null_hash_reset_true`. The non-null hash with a truthy reset flag is still accepted because the preflight does not enforce either half of the readiness contract.
- `harness_preflight_refuses_pending_non_null_hash_nonzero_reset_flag_two` FAIL — same message for `non_null_hash_nonzero_reset_flag_two`. The `must_reset_password=2` nonzero reset flag is still accepted; the preflight treats any pending-state account as finalizable.
- `direct_migration_guard_refuses_pending_*` (all four) PASS trivially — `migrateToVersion` returns `no migration found for version 1: read down for version 1 .: file does not exist`, which the subtest treats as a refusal. This is not evidence of the designed `schema_guard_finalize_credentials` guard; it is the absence of any finalization source. The T027 implementation must later retain this subtest as a true refusal through the designed SQL guard, not through a missing file.
- `writers_quiesced_precondition_single_writer_pool` PASS — `openConfiguredDB` sets `SetMaxOpenConns(1)`; the test does not depend on finalization source and is retained as a passing precondition rather than a red target assertion.
- `empty_table_allows_finalization_and_cleans_helper_routine` FAIL — `credentials_test.go:434: apply finalization on empty user table: no migration found for version 1: read down for version 1 .: file does not exist`. Expected target assertion: on an empty account table the finalization migration would tighten `password_hash` to `VARCHAR(255) NOT NULL` with no default and remove `schema_guard_finalize_credentials`. In this red state the schema stays at `password_hash VARCHAR(255) NULL DEFAULT NULL` (observed through `INFORMATION_SCHEMA.COLUMNS` after the subtest) and no guard routine is created or dropped, so the ready schema still permits NULL hashes.
- `all_ready_accounts_allow_finalization_preserving_hashes_flags_and_profiles` FAIL — `credentials_test.go:455: apply finalization on all-ready user table: no migration found for version 1: read down for version 1 .: file does not exist`. Expected target assertion: once every account has a non-NULL hash and `must_reset_password=FALSE`, finalization would tighten `password_hash` to `VARCHAR(255) NOT NULL`, leave hashes/flags/profiles unchanged, and remove the guard routine. In this red state `migrations/credentials-finalize/` has no SQL, so the all-ready target column definition is never applied; the ready schema still permits NULL hashes (`INFORMATION_SCHEMA.COLUMNS` reports `password_hash varchar(255) YES NULL DEFAULT NULL` after the subtest).
- `finalized_schema_refuses_subsequent_null_and_omitted_hash_writes` FAIL — `credentials_test.go:479: apply finalization before refusal probes: no migration found for version 1: read down for version 1 .: file does not exist`. Expected target assertion: once finalization is applied, subsequent `INSERT ... VALUES (?, NULL)` returns the `nullErr` class and omitted-hash `INSERT` returns the `requiredDefaultErr` class on `password_hash`. In this red state the finalized column cannot be produced, so these refusal assertions cannot be reached.
- `finalization_down_restores_only_nullability_and_refuses_plaintext_restoration` FAIL — `credentials_test.go:493: apply finalization before down: no migration found for version 1: read down for version 1 .: file does not exist`. Expected target assertion: after finalization applied and reversed, `password_hash` returns to `VARCHAR(255) NULL DEFAULT NULL`, `must_reset_password` keeps `TINYINT(1) NOT NULL DEFAULT 1`, the legacy `password` column stays absent, hashes/flags/profiles are preserved and no guard routine remains. The red state cannot reach any of these assertions because the finalization source is absent.
- `failed_finalization_surfaces_truthful_administrative_state` FAIL — `credentials_test.go:527: credentials-finalize dirty = false after failed migration; want dirty=true`. Expected target assertion: a finalization applied against a pending-state account fails inside the guard routine, leaves `schema_migrations_credentials_finalize` dirty=true at version 1, preserves the pre-change `password_hash` definition and account state, and never recreates the legacy `password` column. In this red state no history table exists because the migration source was not found before any history could be written, so the truthful dirty administrative state cannot be observed.

Target schema definition observed after the T024 red run (before the disposable-instance discard below) confirms the "ready schema still permits NULL hashes" target assertion is still unmet:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user' AND COLUMN_NAME IN ('password_hash','must_reset_password'); SELECT COUNT(*) FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_SCHEMA = DATABASE() AND ROUTINE_NAME = 'schema_guard_finalize_credentials';"
```

Output summary: `password_hash varchar(255) YES NULL`, `must_reset_password tinyint(1) NO 1`, `schema_guard_finalize_credentials` routine count `0`. The initial target schema from T016–T019 therefore still permits NULL hashes and carries no finalization guard; these are precisely the target assertions T025–T027 will later close, and they stand as the T024 red evidence.

Prior shared-value-contract red evidence from T010 is retained verbatim in the earlier `## T010 Shared Schema Contract Red Gate` section of this file. No new `TestSchemaValues` red was manufactured in this T024 pass; those values are already implemented through T016–T019 and are green under T022.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard, so no forced history or hand-applied DDL is used to clean up the schema/account rows left after the failed-finalization subtest. No push, merge or deployment was performed.

## T028 Complete TestSchemaValues Execution Against Actual Initial Target Schema

Changelog:

- 2026-11-24 — Executed the complete `TestSchemaValues` suite from `tests/schema/values_test.go` with `-race` against the actual initial target schema from `migrations/schema/000001_complete_mysql_schema.up.sql` on the local disposable MySQL 8.4.2 listener (`127.0.0.1:3307`, `shopping-cart` namespace). All US2.1–11 cases pass against the initial target schema, including the initial-state `password_hash VARCHAR(255) NULL DEFAULT NULL` hash capacity coverage. The finalized-state hash NULL/omission refusal is covered by `TestCredentialFinalization/finalized_schema_refuses_subsequent_null_and_omitted_hash_writes`, which was re-executed here and remains green. The two expectations stay distinct: the initial target permits NULL hashes and omitted-hash inserts, the finalized target refuses both. Credential-only legacy email length (`VARCHAR(60)`) is not treated as full target conformance; the initial target `user.email VARCHAR(100)` and `user.password_hash VARCHAR(255)` cases are each verified directly through the matching `TestSchemaValues/varchar_limits/*` subtests.
- 2026-11-24 — Repaired three independent test-fixture defects in `tests/schema/values_test.go` that otherwise prevented `TestSchemaValues` from executing all US2.1–11 cases: reduced `shortValueKey`'s builder cap from `>= 24` to `>= 18` so synthetic keys for 400-byte 4-byte-rune probes stay within the `product.pid VARCHAR(30)` and FK-referenced `usercart.pid VARCHAR(30)` limits; added `orderNumericArg(prefix)` with distinct `order-q`/`order-s` prefixes so `signed_numeric_boundaries/orders.quantity` and `orders.shipped` no longer collide on `(orderid, prodid)` primary-key values; added `discardGeneratedIdentityFixture` to delete the independent max-id/overflow rows and reset `AUTO_INCREMENT` after each `generated_identities` table probe, keeping that fixture independent from `quantity_checks_and_timestamps` and `demand_duplicates_and_nullable_author` as required by T028; corrected the `hash_capacity` self-assertion from `117` to the actual constructed length `118` so the synthetic `$argon2id$v=19$m=65536,t=3,p=4$<43×s>$<43×h>` encoded hash is accepted. Exact decimal comparisons through `assertDecimalValue`, before/after row counts through `tableCount` and the independent max-id fixture reset are preserved in all paths.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
go vet ./...
gofmt -l tests/schema/values_test.go
```

Exit status: `0` for each command (ping returns `mysqld is alive`; `TestEnvironment` passes; `go vet ./...` reports no findings; `gofmt -l` prints nothing).

T028 TestSchemaValues race command on the actual initial target schema:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaValues$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaValues (1.17s)`.

Per-case results:

| ID | US2.* / Scope | Subtest | Result | Evidence summary |
| --- | --- | --- | --- | --- |
| V1 | US2.1 required/optional defaults | `TestSchemaValues/required_and_optional_values` | PASS | Omitted optional `product` fields return NULL, `pprice` default `0.00`, `user` optional fields NULL, initial `password_hash IS NULL`, initial `must_reset_password = 1`, omitted orders `quantity` NULL with default `amount=0.00` and `shipped=0`, missing-required cases raise `requiredDefaultErr` (1364/HY000), NULL assignments to NOT NULL columns raise `nullErr` (1048/23000), omitted transactions `amount` defaults to `0.00`. |
| V2 | US2.2 VARCHAR exact and one-over, utf8mb4 | `TestSchemaValues/varchar_limits/*` (17 subtests) | PASS | For every character-limited column (`product.pid/pname/ptype/pinfo`, `user.email/name/address/password_hash`, `orders.orderid/prodid`, `transactions.transaction_id/email/pid`, `user_demand.book_name/author`, `usercart.email/pid`) the exact-limit 4-byte 😀 insert succeeds and the one-over insert raises `tooLongErr` (1406/22001) with the table row count unchanged before/after and `COUNT(*)` for the one-over value = 0. `orders.prodid` records the parent-vs-child length difference: a 30-rune `orders.prodid` fails with `fkChildErr` (1452/23000) because the parent `product.pid` cap is 30 (no matching 30-rune parent), and a 31-rune `orders.prodid` fails with `tooLongErr`. Credential-only legacy `user.email VARCHAR(60)` is not conflated with target coverage: target `user.email` is checked directly at 100 characters. |
| V3 | US2.3 binary image round-trip | `TestSchemaValues/binary_image_round_trip` | PASS | `product.image` byte sequence `0x00 0x01 0xfe 0xff 'img'` written through parameter binding and read back byte-for-byte equal. |
| V4 | US2.4 money fidelity/range/default/rounding | `TestSchemaValues/money_fidelity_range_defaults_and_rounding/{product.pprice,orders.amount,transactions.amount}` | PASS | Exact decimal comparisons through `assertDecimalValue` record `0.00`, `12.34`, `99999999.99`, `-99999999.99` round-trip intact; `100000000.00` and `-100000000.00` raise `outOfRangeErr` (1264/22003); `12.345`→`12.35` and `-12.345`→`-12.35` are the characterized MySQL `DECIMAL(10,2)` half-away-from-zero rounding and are recorded here as the T030 environment case rather than strict-mode rejection. |
| V5 | US2.5 signed INT/BIGINT endpoints and overflow | `TestSchemaValues/signed_numeric_boundaries/*` | PASS | For `product.pquantity`, `orders.quantity`, `orders.shipped` and `user.pincode` the exact endpoints `-2147483648` and `2147483647` round-trip through `assertIntStringValue`, and one-over bounds `-2147483649` and `2147483648` raise `outOfRangeErr`. Independent `order-q`/`order-s` key prefixes keep the two orders subtests from colliding on the composite primary key. For `user.mobile` the BIGINT endpoints `-9223372036854775808` and `9223372036854775807` round-trip and the one-over bounds raise `outOfRangeErr`. |
| V6 | US2.6 permitted negative/zero/>1 domain values | `TestSchemaValues/permitted_signed_domain_values` | PASS | `product.pquantity = -7`, `product.pprice = -1.23`, `orders.quantity = -3`, `orders.amount = -4.56` and `orders.shipped = 2` are accepted and read back exactly; this records the Q1 schema's intentional lack of positivity checks on these columns and `shipped` values outside `{0,1}`. |
| V7 | US2.7 generated BIGINT identities on independent max-id/overflow fixtures | `TestSchemaValues/generated_identities/{usercart,user_demand,transactions}` | PASS | For each table two successive auto-generated inserts return distinct `LastInsertId()`; `NULL` id generates a nonzero id; explicit `0` is accepted under `NO_AUTO_VALUE_ON_ZERO`; explicit `-9223372036854775808` and `9223372036854775807` are accepted; a duplicate explicit `0` raises `duplicateErrSpec` (1062/23000); an out-of-range `"9223372036854775808"` raises `outOfRangeErr` (1264/22003). `discardGeneratedIdentityFixture` then deletes the max-id rows and resets `AUTO_INCREMENT` so this fixture is independent from the subsequent tests. |
| V8 | US2.8 enforced quantity checks and bounded timestamps | `TestSchemaValues/quantity_checks_and_timestamps` | PASS | Omitting `usercart.added_at` records `CURRENT_TIMESTAMP` within ±2 seconds of `time.Now().UTC()`; explicit `2026-11-17 12:34:56` is retained exactly; `quantity ∈ {0, -1}` raise `checkErr` (3819/HY000) from `ck_usercart_quantity`; `quantity = 2147483648` raises `outOfRangeErr`; `added_at = NULL` raises `nullErr`. The same cases apply to `transactions.created_at` and `ck_transactions_quantity`. |
| V9 | US2.9 demand nullable author and duplicates | `TestSchemaValues/demand_duplicates_and_nullable_author` | PASS | Three `user_demand` rows for identical `(email, book_name)` coexist with authors `omitted`, explicit `NULL` and a non-null value; two rows have `author IS NULL`. |
| V10 | US2.10 finalized hash NULL/omission behavior | `TestCredentialFinalization/finalized_schema_refuses_subsequent_null_and_omitted_hash_writes` (also preflight `harness_preflight_refuses_pending_null_hash_reset_true` and `empty_table_allows_finalization_and_cleans_helper_routine`) | PASS | After `migrations/credentials-finalize/000001_require_password_hash.up.sql` tightens `password_hash` to `VARCHAR(255) NOT NULL` with no default, `INSERT INTO user (email, password_hash) VALUES (?, NULL)` raises `nullErr` and `INSERT INTO user (email) VALUES (?)` raises `requiredDefaultErr`. Initial-state `must_reset_password=1`/`password_hash IS NULL` accounts are still refused finalization by `checkFinalizationReadiness`. The initial-state and finalized-state expectations stay distinct: V1 verifies the initial target schema still permits NULL and omitted hashes; V10 verifies the finalized target rejects both. |
| V11 | US2.11 hash capacity 255/256 | `TestSchemaValues/hash_capacity` | PASS | A synthetic 118-character PHC-style encoded hash (`$argon2id$v=19$m=65536,t=3,p=4$` + 43 × `s` + `$` + 43 × `h`) and a 255-character boundary hash round-trip; a 256-character hash raises `tooLongErr` with the user row count unchanged. |

Initial-state vs finalized-state separation of concerns:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestCredentialFinalization$/^(finalized_schema_refuses_subsequent_null_and_omitted_hash_writes|harness_preflight_refuses_pending_null_hash_reset_true|empty_table_allows_finalization_and_cleans_helper_routine)$' -count=1 -v
```

Exit status: `0`. Results: `--- PASS: TestCredentialFinalization/harness_preflight_refuses_pending_null_hash_reset_true`, `--- PASS: TestCredentialFinalization/empty_table_allows_finalization_and_cleans_helper_routine`, `--- PASS: TestCredentialFinalization/finalized_schema_refuses_subsequent_null_and_omitted_hash_writes`. The initial target schema still permits NULL and omitted `password_hash` (recorded in V1/V11 above); the finalized target schema rejects both; the harness preflight refuses finalization of pending NULL-hash accounts. These two states are recorded as distinct expectations and were not conflated with legacy credential-only email length coverage.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard (no `SHOW TABLES` rows), so no forced history or hand-applied DDL is used to clean up the max-id/overflow or finalization fixtures. No push, merge or deployment was performed. Passwords in any DSN above are redacted.


### T028 Re-run Confirmation — 2026-11-24 speckit-implement pass

Changelog:

- 2026-11-24 — Re-ran the complete `TestSchemaValues` suite under `-race` against the actual initial target schema on the restarted local disposable MySQL 8.4.2 instance (`127.0.0.1:3307`, `shopping-cart`). No source change was required: all US2.1–11 subtests and the finalized-state NULL/omission refusal subtests remain green. Dirty state from the deliberately refused finalization-pending cases was discarded through `DROP TABLE` on the disposable namespace; no `migrate.Force` was used.

Race commands (each separated by the shared cleanup documented above):

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaValues$' -count=1 -v
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestCredentialFinalization$/^(finalized_schema_refuses_subsequent_null_and_omitted_hash_writes|harness_preflight_refuses_pending_null_hash_reset_true|empty_table_allows_finalization_and_cleans_helper_routine)$' -count=1 -v
```

Exit status for each command: `0`.

Race-run results:

- `TestSchemaValues`: `--- PASS: TestSchemaValues (1.09s)` with every US2.1–11 subtest `PASS` on the initial target schema from `migrations/schema/000001_complete_mysql_schema.up.sql`. All 17 `varchar_limits/*` subtests, the three `money_fidelity_range_defaults_and_rounding/*` subtests, the four `signed_numeric_boundaries/*` subtests, the three `generated_identities/*` subtests (with the independent max-id/overflow fixture reset through `discardGeneratedIdentityFixture`) and the `required_and_optional_values`, `binary_image_round_trip`, `permitted_signed_domain_values`, `quantity_checks_and_timestamps`, `demand_duplicates_and_nullable_author` and `hash_capacity` subtests pass against the initial target column types and defaults. The initial-state expectations (NULL-permitting `password_hash`, default `must_reset_password=1`, omitted-hash insert accepted, `DECIMAL(10,2)` exact round-trip and half-away-from-zero rounding at `12.345→12.35`/`-12.345→-12.35`) are preserved verbatim; credential-only legacy `user.email VARCHAR(60)` remains outside the target conformance claim because `varchar_limits/user.email` checks the target `VARCHAR(100)` boundary directly.
- `TestCredentialFinalization/harness_preflight_refuses_pending_null_hash_reset_true`: `PASS` — the harness preflight still refuses pending NULL-hash/reset-true accounts.
- `TestCredentialFinalization/empty_table_allows_finalization_and_cleans_helper_routine`: `PASS` — an empty `user` table lets finalization tighten `password_hash` to `VARCHAR(255) NOT NULL` with no default and drops the `schema_guard_finalize_credentials` routine on success.
- `TestCredentialFinalization/finalized_schema_refuses_subsequent_null_and_omitted_hash_writes`: `PASS` — after finalization the target column refuses `INSERT ... VALUES (?, NULL)` with `nullErr` (1048/23000) and refuses omitted `password_hash` inserts with `requiredDefaultErr` (1364/HY000). Initial-state vs finalized-state expectations stay distinct: V1/V11 above (initial target) permits NULL and omitted hashes; V10/finalization subtest (finalized target) refuses both.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard (no `SHOW TABLES` rows), so no forced history or hand-applied DDL is used to clean up the max-id/overflow or finalization fixtures. No push, merge or deployment was performed. Passwords in any DSN above are redacted.
## T029 TestCredentialFinalization and Finalization-Aware Lifecycle Execution

Changelog:

- 2026-11-24 — Executed `TestCredentialFinalization` from `tests/schema/credentials_test.go` and the finalization-aware lifecycle cases (`schema_up_does_not_implicitly_finalize_credentials`, `finalization_down_before_bootstrap_reconstruction`, `recreated_nullable_hashes_do_not_retain_stale_finalized_history`) from `tests/schema/lifecycle_test.go` against the completed T025–T027 fixtures (`migrations/credentials-finalize/000001_require_password_hash.{up,down}.sql` and the finalization helpers in `tests/schema/helpers_test.go`) on the local disposable MySQL 8.4.2 instance (`127.0.0.1:3307`, `shopping-cart` namespace). All pending-state preflight and SQL guard refusals, the all-ready and empty-table success paths with preserved values and absent guard routine, the finalization `down.sql` restoring only nullability, the separate `schema_migrations_credentials_finalize` history, the explicit finalization-down before schema reconstruction, and the no-stale-applied-finalization behavior over recreated nullable hashes pass under race detection. The failed-finalization case is reported as a failure/refusal with dirty=true at version 1 and preserved column/account state, not as a successful reversion. No source change was required for T029 beyond this evidence entry; T025–T027 own the implementation surface and were not modified. The disposable fixture dirty from `direct_migration_guard_refuses_pending_*` and `failed_finalization_surfaces_truthful_administrative_state` is discarded through `DROP TABLE`/`DROP PROCEDURE` on the disposable namespace rather than `migrate.Force`.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
```

Exit status: `0` and `0` (ping returns `mysqld is alive`; `TestEnvironment` passes against `shopping-cart` with the required strict modes, UTC, utf8mb4/utf8mb4_0900_bin and `foreign_key_checks=1`). Passwords are redacted above.

T029 TestCredentialFinalization race command on the actual initial target schema and T025–T027 finalization fixtures:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestCredentialFinalization$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestCredentialFinalization (10.18s)` with every subtest PASS.

Per-subtest results:

| ID | Scope / US2.12, Q2 or SC- | Subtest | Result | Evidence summary |
| --- | --- | --- | --- | --- |
| F1 | US2.12 pending-state preflight refusal | `harness_preflight_refuses_pending_null_hash_reset_true` | PASS | `checkFinalizationReadiness` refuses an account with `password_hash IS NULL` and `must_reset_password = TRUE`; `schema_migrations_credentials_finalize` remains at `version=0 dirty=false`, the pre/post account state snapshot is byte-identical, `password_hash` stays `varchar(255) NULL DEFAULT NULL`, and `schema_guard_finalize_credentials` is absent. |
| F2 | US2.12 pending-state preflight refusal | `harness_preflight_refuses_pending_null_hash_reset_false` | PASS | Same preflight refusal for NULL hash with `must_reset_password = FALSE`; no mutation observed. |
| F3 | US2.12 pending-state preflight refusal | `harness_preflight_refuses_pending_non_null_hash_reset_true` | PASS | Same preflight refusal for non-null hash with a truthy reset flag; no mutation observed. |
| F4 | US2.12 pending-state preflight refusal | `harness_preflight_refuses_pending_non_null_hash_nonzero_reset_flag_two` | PASS | Same preflight refusal when `must_reset_password = 2` (any nonzero reset flag counts as awaiting reset). |
| F5 | US2.12 SQL guard refusal via direct migration | `direct_migration_guard_refuses_pending_null_hash_reset_true` | PASS | `migrateFinalizationDirect` bypasses the Go preflight and the `schema_guard_finalize_credentials` routine inside `migrations/credentials-finalize/000001_require_password_hash.up.sql` raises `SIGNAL SQLSTATE '45000'`; `inspectFinalizationOwnership` reports `Dirty=true` on `schema_migrations_credentials_finalize` (truthful dirty administrative state), pre/post `password_hash` definition and account state are byte-identical, and the disposable fixture is discarded through `discardFinalizationFixture` rather than `migrate.Force`. The refusal is recorded as a failure/refusal, never as a successful reversion. |
| F6 | US2.12 SQL guard refusal via direct migration | `direct_migration_guard_refuses_pending_null_hash_reset_false` | PASS | Same SQL guard refusal for NULL hash with `must_reset_password = FALSE`; dirty administrative state recorded and disposable fixture discarded. |
| F7 | US2.12 SQL guard refusal via direct migration | `direct_migration_guard_refuses_pending_non_null_hash_reset_true` | PASS | Same SQL guard refusal for non-null hash with a truthy reset flag; dirty administrative state recorded and disposable fixture discarded. |
| F8 | US2.12 SQL guard refusal via direct migration | `direct_migration_guard_refuses_pending_non_null_hash_nonzero_reset_flag_two` | PASS | Same SQL guard refusal for the nonzero (`2`) reset flag; dirty administrative state recorded and disposable fixture discarded. |
| F9 | Quiesced-writer precondition | `writers_quiesced_precondition_single_writer_pool` | PASS | `assertFinalizationWriterQuiescence` documents the single-writer boundary shared across the Go preflight, SQL guard and `ALTER` in `openConfiguredDB`'s `SetMaxOpenConns(1)` pool. |
| F10 | US2.12 empty-table success path | `empty_table_allows_finalization_and_cleans_helper_routine` | PASS | On an empty user table the finalization migration applies, `password_hash` tightens to `varchar(255) NOT NULL` with no default, `schema_guard_finalize_credentials` is dropped on success, and `schema_migrations_credentials_finalize` reaches the target version cleanly. |
| F11 | US2.12/Q2 all-ready success path preserving values | `all_ready_accounts_allow_finalization_preserving_hashes_flags_and_profiles` | PASS | With three seeded all-ready accounts (`must_reset_password=0`, non-null `password_hash`), the finalization migration applies: `password_hash` tightens to `varchar(255) NOT NULL`; `snapshotFinalizationProfiles`, `snapshotProtectedCredentials` and `snapshotFinalizationState` are byte-identical before/after; the legacy `password` column stays absent; the guard routine is dropped on success. |
| F12 | US2.12 finalized schema refusal (initial-state vs finalized-state separation) | `finalized_schema_refuses_subsequent_null_and_omitted_hash_writes` | PASS | After finalization, `INSERT INTO user (email, password_hash) VALUES (?, NULL)` raises `nullErr` and omitted-`password_hash` `INSERT` raises `requiredDefaultErr`; the initial target schema continues to permit both (recorded in T028 V1/V11). |
| F13 | US2.12 safe down restoring only nullability, retaining values | `finalization_down_restores_only_nullability_and_refuses_plaintext_restoration` | PASS | `migrateToVersion(finalizeCredentialsTrack, 0)` applies `migrations/credentials-finalize/000001_require_password_hash.down.sql`; `password_hash` returns to `varchar(255) NULL DEFAULT NULL`; `must_reset_password` stays `tinyint(1) NOT NULL DEFAULT 1`; the legacy `password` column stays absent (no plaintext restoration); `snapshotFinalizationProfiles` and `snapshotProtectedCredentials` are byte-identical before/after; `schema_guard_finalize_credentials` is absent. |
| F14 | US2.12 dirty failure surfaced truthfully | `failed_finalization_surfaces_truthful_administrative_state` | PASS | Finalization applied against an account with `password_hash IS NULL`/`must_reset_password=1` fails inside the SQL guard; `assertTrackDirty` reads `schema_migrations_credentials_finalize` directly and reports `dirty=true` at the expected version; `snapshotUserHashDefinition` and `snapshotFinalizationState` are byte-identical before/after; the legacy `password` column stays absent. Reported as a failure/refusal — never as a successful reversion. |

Separate-history ownership confirmation (US2.12, SC-005 reviewable handoff):

- `selectFinalizationTrack` returns the dedicated `finalize-credentials` track whose source directory is `migrations/credentials-finalize/` and whose history table is `schema_migrations_credentials_finalize`, distinct from `schema_migrations_schema` and `schema_migrations_credentials`. The subtests in `TestCredentialFinalization` call `assertTrackVersion(ctx, finalizeCredentialsTrack, ...)` and `assertTrackDirty(ctx, finalizeCredentialsTrack)` against this dedicated history table, and the lifecycle subtests in `TestSchemaLifecycle` call `assertTrackVersion(ctx, finalizeCredentialsTrack, 0, false)` after `schema` up to prove `schema up` does not implicitly advance the finalization history.

T029 finalization-aware lifecycle race command (explicit finalization-down before schema reconstruction and no stale applied finalization over recreated nullable hashes):

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaLifecycle$/^(schema_up_does_not_implicitly_finalize_credentials|finalization_down_before_bootstrap_reconstruction|recreated_nullable_hashes_do_not_retain_stale_finalized_history)$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaLifecycle (3.32s)` with every subtest PASS.

Per-subtest results:

| ID | Scope / US2.12, Q2 or SC- | Subtest | Result | Evidence summary |
| --- | --- | --- | --- | --- |
| L1 | US2.12 separate histories / no implicit finalization | `schema_up_does_not_implicitly_finalize_credentials` | PASS | After `schema` up on an empty namespace, `inspectLifecycleDefinition` reports `user.password_hash varchar(255) NULL DEFAULT NULL` and `user.must_reset_password tinyint(1) NOT NULL DEFAULT 1`; `schema_migrations_credentials_finalize` is at `version=0 dirty=false`; `schema_guard_finalize_credentials` is absent. The schema up migration does not implicitly advance the finalization history or tighten the hash column. |
| L2 | US2.12 explicit finalization-down before schema reconstruction | `finalization_down_before_bootstrap_reconstruction` | PASS | `schema` up → `finalize-credentials` up (empty table success) → `finalize-credentials` down → `schema` down → `schema` up. After the explicit finalization down the hash returns to `varchar(255) NULL DEFAULT NULL`; after reconstruction the hash is again `varchar(255) NULL DEFAULT NULL`, `schema_migrations_credentials_finalize` stays at `version=0 dirty=false`, and no guard routine remains. Finalization down is explicit; the schema reconstruction does not silently drag a finalized hash through the DROP/CREATE cycle. |
| L3 | US2.12 no stale applied finalization over recreated nullable hashes | `recreated_nullable_hashes_do_not_retain_stale_finalized_history` | PASS | Full cycle `schema` up → `finalize-credentials` up → `finalize-credentials` down → `schema` down → `schema` up. `schema_migrations_credentials_finalize` ends at `version=0 dirty=false`, the recreated `user.password_hash` is `varchar(255) NULL DEFAULT NULL`, and `schema_guard_finalize_credentials` is absent. No stale "applied finalization" history is left over the freshly recreated nullable hashes. |

Coverage mapping:

| Requirement | Evidence |
| --- | --- |
| US2.12 — pending-state refusal and all-ready/empty finalization | F1–F8 (preflight/SQL guard refusals on all four pending states), F10 (empty-table success), F11 (all-ready success preserving values) |
| US2.12 — safe down restoring only nullability, no plaintext | F13 (restoration of `varchar(255) NULL DEFAULT NULL` without legacy `password` recreation, hashes/flags/profiles unchanged) |
| US2.12 — dirty failure surfaced truthfully, not reverted | F14 (dirty=true at version 1, pre/post column and account state identical, `password` absent) |
| US2.12 — separate finalization history vs schema/credentials | L1 (schema up leaves `finalize-credentials` at 0/false), F10/F11 (track advances and resets only through its own history) |
| US2.12 — explicit finalization-down before schema reconstruction | L2 (explicit `finalize-credentials` down before `schema` down and the subsequent reconstruction) |
| US2.12 — no stale applied finalization over recreated nullable hashes | L3 (recreated `user.password_hash` is `varchar(255) NULL DEFAULT NULL`, history at 0/false, guard absent) |
| Q2 — credential-column migration, plaintext never read/stored | F11 (all-ready success leaves hashes/flags/profiles unchanged; no plaintext read or restored at any point); F13 (down does not recreate the legacy `password` column) |
| SC-003 — Value fidelity for US2.12 final-state cases | F11 (hashes/flags/profiles byte-identical before/after); F12 (finalized target rejects NULL and omitted hashes with `nullErr`/`requiredDefaultErr`); F13 (round-trip restoration of nullability without value loss) |
| SC-005 — Reviewable handoff distinct from code inspection | This T029 section traces every US2.12 scenario and the Q2 credential-migration contract to a specific executed subtest and command, allowing a reviewer to distinguish confirmed behavior from UNVERIFIED items without inspecting application code. The separate `schema_migrations_credentials_finalize` history, the SQL guard refusal, and the explicit finalization-down before schema reconstruction are each recorded here with their observed outcomes. |

Dirty-state honesty check (US2.12, SC-005):

- `direct_migration_guard_refuses_pending_*` and `failed_finalization_surfaces_truthful_administrative_state` are the two cases that deliberately leave the finalization fixture dirty. In both cases `assertTrackDirty` / `inspectFinalizationOwnership` report `dirty=true` through a direct read of `schema_migrations_credentials_finalize`, pre/post `password_hash` definition and account state are byte-identical, and the legacy `password` column remains absent. No subtest treats a dirty failure as a successful reversion. The disposable fixture is discarded through `discardFinalizationFixture` or the shared `DROP TABLE`/`DROP PROCEDURE` discard rather than `migrate.Force`.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard (no `SHOW TABLES` rows), so no forced history or hand-applied DDL is used to clean up the deliberately dirty finalization fixtures. No push, merge or deployment was performed. Passwords in any DSN above are redacted.

## T030 Environment Characterization: Collation Semantics and DECIMAL Excess-Fraction Conversion

Changelog:

- 2026-11-24 — Added and executed the environment-characterization cases prepared in `tests/schema/environment_test.go` (`TestEnvironmentCharacterization`) and `tests/schema/values_test.go` (`TestDecimalConversionCharacterization`) on the local disposable MySQL 8.4.2 instance (`127.0.0.1:3307`, `shopping-cart` namespace, binary collation `utf8mb4_0900_bin`, STRICT_ALL_TABLES/NO_ENGINE_SUBSTITUTION/NO_AUTO_VALUE_ON_ZERO, UTC, `foreign_key_checks=1`). The characterization records actual target server behavior for case/accent/trailing-space identity under `utf8mb4_0900_bin` and for exact `12.345`→`12.35` / `-12.345`→`-12.35` conversion into `DECIMAL(10,2)` including the `Note 1265 Data truncated` warning emitted by the server. These target-environment choices are kept distinct from UNVERIFIED legacy case/whitespace behavior (recorded as an evidence limitation in `plan.md`); the characterization is a target contract, not a reconstruction of legacy rules. Each subtest asserts the recorded target behavior and fails loudly with reconciliation guidance if the server deviates. No existing `TestSchemaValues/money_fidelity_range_defaults_and_rounding` or `TestEnvironment` expectation was weakened.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
gofmt -l tests/schema/environment_test.go tests/schema/values_test.go
go vet ./tests/schema
```

Exit status: `0` for each command (ping returns `mysqld is alive`; `TestEnvironment` passes; `gofmt -l` prints nothing; `go vet` reports no findings). Passwords are redacted above.

T030 TestEnvironmentCharacterization race command on the actual target MySQL 8.4.2 session:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestEnvironmentCharacterization$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestEnvironmentCharacterization (0.01s)` with every subtest `PASS`.

Per-case results (collation characterization under the target `utf8mb4_0900_bin`, applied through `SET NAMES utf8mb4 COLLATE utf8mb4_0900_bin`):

| ID | Scope | Subtest | Result | Target behavior recorded |
| --- | --- | --- | --- | --- |
| E1 | `PAD_ATTRIBUTE` of the named collation | `TestEnvironmentCharacterization/pad_attribute_is_NO_PAD` | PASS | `INFORMATION_SCHEMA.COLLATIONS.PAD_ATTRIBUTE` for `utf8mb4_0900_bin` = `NO PAD`. This is the target contract: trailing spaces are significant for identity equality and length. If the server reports `PAD SPACE` instead, the subtest fails with reconciliation guidance and the environment contract must be updated before the test is relaxed. |
| E2 | Case sensitivity under `utf8mb4_0900_bin` | `TestEnvironmentCharacterization/case_sensitive_upper_vs_lower` | PASS | `(_utf8mb4'A' COLLATE utf8mb4_0900_bin = _utf8mb4'a' COLLATE utf8mb4_0900_bin)` returns `0` (not equal); `STRCMP('A','a')` returns `-1` (nonzero). The target identity equality is case-sensitive. |
| E3 | Accent sensitivity under `utf8mb4_0900_bin` | `TestEnvironmentCharacterization/accent_sensitive_e_vs_e-acute` | PASS | `(_utf8mb4'é' COLLATE utf8mb4_0900_bin = _utf8mb4'e' COLLATE utf8mb4_0900_bin)` returns `0` (not equal); `STRCMP('é','e')` returns `1` (nonzero). The target identity equality is accent-sensitive. |
| E4 | Trailing-space preservation (NO PAD) under `utf8mb4_0900_bin` | `TestEnvironmentCharacterization/trailing_space_preserved_no_pad` | PASS | `(_utf8mb4'abc' COLLATE utf8mb4_0900_bin = _utf8mb4'abc ' COLLATE utf8mb4_0900_bin)` returns `0` (not equal); `STRCMP('abc','abc ')` returns `-1` (nonzero); `CHAR_LENGTH('abc ')` returns `4` (the trailing space is retained). The target comparison and stored length both honor NO PAD. |
| E5 | Session collation matches required | `TestEnvironmentCharacterization/session_collation_matches_required` | PASS | `@@collation_connection` = `utf8mb4_0900_bin` after `openConfiguredDB` applies `SET NAMES utf8mb4 COLLATE utf8mb4_0900_bin`. The characterization above applies only under this selected target collation; if the session falls back to a different collation the subtest fails. |

T030 TestDecimalConversionCharacterization race command on the actual target MySQL 8.4.2 session:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestDecimalConversionCharacterization$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestDecimalConversionCharacterization (0.01s)` with every subtest `PASS`.

Per-case results (`DECIMAL(10,2)` excess-fraction conversion from exact string input under STRICT_ALL_TABLES, inserted into a disposable `CREATE TEMPORARY TABLE _t030_decimal_characterization (amount DECIMAL(10,2) NOT NULL) ENGINE=InnoDB`):

| ID | Scope | Subtest | Input | Stored | Warning | Target behavior recorded |
| --- | --- | --- | --- | --- | --- | --- |
| D1 | Positive half-away-from-zero conversion to two decimal places | `TestDecimalConversionCharacterization/positive_half_up_to_next_hundredth` | `'12.345'` | `12.35` | `Note 1265 Data truncated for column 'amount' at row 1` | Under STRICT_ALL_TABLES the server rounds `12.345` to `12.35` (half-away-from-zero) and emits a `Note 1265` warning rather than rejecting the insert. This is the target contract called out in research for the Java-to-Go rebuild; it is kept visibly failing if the server behavior diverges. |
| D2 | Negative half-away-from-zero conversion to two decimal places | `TestDecimalConversionCharacterization/negative_half_away_from_zero` | `'-12.345'` | `-12.35` | `Note 1265 Data truncated for column 'amount' at row 1` | Under STRICT_ALL_TABLES the server rounds `-12.345` to `-12.35` (half-away-from-zero) and emits a `Note 1265` warning rather than rejecting the insert. Same visibility guarantee as D1. |

TestSchemaValues/money_fidelity_range_defaults_and_rounding re-execution against the schema-owned `DECIMAL(10,2)` columns (`product.pprice`, `orders.amount`, `transactions.amount`):

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaValues$/^money_fidelity_range_defaults_and_rounding$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaValues/money_fidelity_range_defaults_and_rounding (0.13s)` with each of `product.pprice`, `orders.amount` and `transactions.amount` subtests `PASS`. The exact decimal comparisons from `assertMoneyValues` re-confirm `12.345`→`12.35` and `-12.345`→`-12.35` round-trip into the schema's own `DECIMAL(10,2)` columns, matching D1/D2 above; the `assertDecimalValue` strings are exact (`CAST(... AS CHAR)`), not truncated by driver conversion.

Target-environment vs UNVERIFIED legacy separation:

- The entries above record actual target MySQL 8.4.2 behavior under the Java-to-Go rebuild's selected collation and strict session modes. They are not a reconstruction of legacy case/whitespace/numeric semantics; `plan.md` records the legacy case/whitespace, runtime, endpoint and screen evidence as an UNVERIFIED limitation, and `research.md` notes the chosen binary collation as a target environment choice rather than legacy comparison rules. The subtests assert the recorded target behavior and fail with reconciliation guidance in their error text if the server deviates, keeping unexpected results visibly failing until the environment contract is reconciled.
- No legacy fixture or UNVERIFIED legacy comparison rule is relied on by any T030 subtest. `tests/schema/testdata/legacy/000001_legacy_user.up.sql` is not applied by this task.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard (no `SHOW TABLES` rows); the temporary `_t030_decimal_characterization` table only existed inside the connection scoped by `TestDecimalConversionCharacterization` and was dropped by that subtest's deferred cleanup. No forced history or hand-applied DDL was used, no `migrate.Force` was invoked, and no push, merge or deployment was performed. Passwords in any DSN above are redacted.

## T031 Identity/Pair/Orphan Groups Against the Schema-Only Track

Changelog:

- 2026-11-24 — Executed the `duplicate_identities_and_valid_distinct_pairs` and `foreign_key_orphan_refusals` subtests of `TestSchemaReferences` (`tests/schema/references_test.go`) against the schema-only track applied by `migrations/schema/000001_complete_mysql_schema.up.sql` on the local disposable MySQL 8.4.2 instance (`127.0.0.1:3307`, `shopping-cart` namespace). All identity rejections (product, user, transaction `transaction_id`, generated transaction/cart/demand BIGINT id), the duplicate order `(orderid, prodid)` pair rejection and all six FK insert/update orphan rejections raise the expected `duplicateErrSpec` (1062/23000) or `fkChildErr` (1452/23000); each refused write leaves every reference-table snapshot byte-identical through `assertReferenceRowsEqual`, and the valid distinct pairs plus the three duplicate `user_demand` requests with identical `(email, book_name, author)` are accepted with `COUNT(*) = 3`. The embedded `orders.prodid` length-vs-FK matrix verifies that a 30-character `prodid` matching the parent `product.pid` succeeds, a nonexistent 31-character `prodid` raises `fkChildErr` (1452/23000) and a 46-character `prodid` raises `tooLongErr` (1406/22001); both refusals leave the reference snapshots unchanged. No source change was required for T031; the shared constraints were authored once in T016–T019 (see `migrations/schema/000001_complete_mysql_schema.up.sql`) and the test cases were authored in T009 and demonstrated red in T010. The two historical/Q1/mixed-reference subtests `historical_product_protection`, `q1_cascade_and_restrict_actions`, `mixed_reference_atomicity` and `identical_enforcement_after_reconstruction` are owned by T032–T034 and are not re-executed here.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
gofmt -l tests/schema/
go vet ./tests/schema
```

Exit status: `0` for each command (ping returns `mysqld is alive`; `TestEnvironment` passes; `gofmt -l tests/schema/` prints nothing; `go vet ./tests/schema` reports no findings). Passwords are redacted above.

T031 race command for the identity/pair/orphan groups against the schema-only track:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaReferences$/^(duplicate_identities_and_valid_distinct_pairs|foreign_key_orphan_refusals)$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaReferences (0.78s)` with every selected subtest PASS:

```
--- PASS: TestSchemaReferences/duplicate_identities_and_valid_distinct_pairs (0.13s)
--- PASS: TestSchemaReferences/foreign_key_orphan_refusals (0.28s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/orders.prodid (0.03s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/usercart.email (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/usercart.pid (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/user_demand.email (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/transactions.email (0.08s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/transactions.pid (0.02s)
```

Per-case results (identity/pair rejections and valid distinct/duplicate-request acceptances in `duplicate_identities_and_valid_distinct_pairs`):

| ID | US3.* | Case | Expected error | Observed outcome |
| --- | --- | --- | --- | --- |
| R1 | US3.1 | Duplicate `product.pid` insert (`dup-product`) | `duplicateErrSpec` (1062/23000) | Rejected; `assertReferenceRowsEqual` reports the six reference tables byte-identical before/after (`duplicate product identity`). A distinct `product.pid = "distinct-product"` insert is then accepted. |
| R2 | US3.1 | Duplicate `user.email` insert (`dup@example.test`) | `duplicateErrSpec` (1062/23000) | Rejected; reference snapshots unchanged (`duplicate user identity`). A distinct `user.email = "distinct@example.test"` insert is then accepted. |
| R3 | US3.1 | Duplicate `transactions.transaction_id` insert (`tx-dup`) with distinct generated id | `duplicateErrSpec` (1062/23000) | Rejected; reference snapshots unchanged (`duplicate transaction_id`). |
| R4 | US3.1 | Duplicate generated `transactions.id = 10` with distinct `transaction_id` | `duplicateErrSpec` (1062/23000) | Rejected; reference snapshots unchanged (`duplicate generated transaction id`). The explicit 0/BIGINT boundary id contract from T028 V7 is not weakened. |
| R5 | US3.2 | Duplicate `orders` `(orderid, prodid) = ("order-1","dup-product")` | `duplicateErrSpec` (1062/23000) | Rejected; reference snapshots unchanged (`duplicate order product pair`). Subsequent `("order-1","distinct-product")` and `("order-2","dup-product")` inserts are accepted, confirming the composite identity permits distinct pairs. |
| R6 | US3.9 | Duplicate `usercart` `(email, pid) = ("dup@example.test","dup-product")` | `duplicateErrSpec` (1062/23000) | Rejected via `uq_usercart_email_pid`; reference snapshots unchanged (`duplicate cart pair`). Subsequent `("distinct@example.test","dup-product")` and `("dup@example.test","distinct-product")` inserts are accepted. |
| R7 | US3.1 | Duplicate generated `usercart.id = 20` with distinct `(email, pid)` | `duplicateErrSpec` (1062/23000) | Rejected; reference snapshots unchanged (`duplicate generated cart id`). |
| R8 | US3.1 | Duplicate generated `user_demand.id = 30` with different book/author | `duplicateErrSpec` (1062/23000) | Rejected; reference snapshots unchanged (`duplicate generated demand id`). |
| R9 | US3.9 | Two additional `user_demand` rows with identical `(email, book_name, author) = ("dup@example.test","same-book","same-author")` after the seeded row | accepted | `COUNT(*) FROM user_demand WHERE email = "dup@example.test" AND book_name = "same-book" AND author = "same-author"` returns `3`; no unsupplied demand uniqueness is imposed. |

Per-case results (FK insert/update orphan refusals in `foreign_key_orphan_refusals`). Each case takes a before-snapshot through `snapshotReferenceRows`, asserts `fkChildErr` (1452/23000) on the orphan insert, then seeds a valid parent row and asserts `fkChildErr` on the orphan update; the reference-table snapshot is byte-identical before/after every refusal:

| ID | US3.* | Table.column | Insert refusal | Update refusal |
| --- | --- | --- | --- | --- |
| O1 | US3.3 | `orders.prodid` → `product.pid` (`productid` NO ACTION/NO ACTION) | `INSERT INTO orders (orderid, prodid) VALUES ('orphan-order', 'missing-product')` rejected with `fkChildErr`; `orders.prodid orphan insert` snapshot unchanged. | After seeding `('orphan-update-order','fk-product')`, `UPDATE orders SET prodid = 'missing-product' WHERE orderid = 'orphan-update-order' AND prodid = 'fk-product'` rejected with `fkChildErr`; `orders.prodid orphan update` snapshot unchanged. |
| O2 | US3.9 | `usercart.email` → `user.email` (`fk_usercart_email` CASCADE/CASCADE) | `INSERT INTO usercart (email, pid) VALUES ('missing@example.test', 'fk-product')` rejected with `fkChildErr`; snapshot unchanged. | After seeding `('fk@example.test','fk-product')`, `UPDATE usercart SET email = 'missing@example.test' WHERE email = 'fk@example.test' AND pid = 'fk-product'` rejected with `fkChildErr`; snapshot unchanged. |
| O3 | US3.9 | `usercart.pid` → `product.pid` (`fk_usercart_pid` CASCADE on update, RESTRICT on delete) | `INSERT INTO usercart (email, pid) VALUES ('fk@example.test', 'missing-product')` rejected with `fkChildErr`; snapshot unchanged. | After seeding `('fk2@example.test','fk-product')`, `UPDATE usercart SET pid = 'missing-product' WHERE email = 'fk2@example.test' AND pid = 'fk-product'` rejected with `fkChildErr`; snapshot unchanged. |
| O4 | US3.9 | `user_demand.email` → `user.email` (`fk_user_demand_email` CASCADE/CASCADE) | `INSERT INTO user_demand (email, book_name) VALUES ('missing@example.test', 'missing-user-book')` rejected with `fkChildErr`; snapshot unchanged. | After seeding `('fk@example.test','demand-update')`, `UPDATE user_demand SET email = 'missing@example.test' WHERE email = 'fk@example.test' AND book_name = 'demand-update'` rejected with `fkChildErr`; snapshot unchanged. |
| O5 | US3.9 | `transactions.email` → `user.email` (`fk_transactions_email` CASCADE on update, RESTRICT on delete) | `INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES ('tx-missing-email', 'missing@example.test', 'fk-product', 1)` rejected with `fkChildErr`; snapshot unchanged. | After seeding `('tx-update-email','fk@example.test','fk-product',1)`, `UPDATE transactions SET email = 'missing@example.test' WHERE transaction_id = 'tx-update-email'` rejected with `fkChildErr`; snapshot unchanged. |
| O6 | US3.9 | `transactions.pid` → `product.pid` (`fk_transactions_pid` CASCADE on update, RESTRICT on delete) | `INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES ('tx-missing-product', 'fk@example.test', 'missing-product', 1)` rejected with `fkChildErr`; snapshot unchanged. | After seeding `('tx-update-product','fk@example.test','fk-product-2',1)`, `UPDATE transactions SET pid = 'missing-product' WHERE transaction_id = 'tx-update-product'` rejected with `fkChildErr`; snapshot unchanged. |

`orders.prodid` length-vs-FK matrix (embedded in `foreign_key_orphan_refusals`, after the six FK cases):

| ID | US3.* | Case | Expected outcome | Observed outcome |
| --- | --- | --- | --- | --- |
| L1 | US3.3 | 30-character `product.pid` (`"p"×30`) seeded, then `INSERT INTO orders (orderid, prodid) VALUES ('length30-parent', <30-char>)` | Accept | Succeeds (matching parent `product.pid VARCHAR(30)` maximum). |
| L2 | US3.3 | `INSERT INTO orders (orderid, prodid) VALUES ('length31-no-parent', <31-char>)` with no matching parent | Reject as FK orphan with `fkChildErr` (1452/23000) | Rejected; `orders prodid length31 FK` reference snapshots unchanged. `orders.prodid VARCHAR(45)` accepts the 31-character length at storage level, so the write reaches the FK and fails there. |
| L3 | US3.3 | `INSERT INTO orders (orderid, prodid) VALUES ('length46-child', <46-char>)` | Reject as length overflow with `tooLongErr` (1406/22001) | Rejected before FK evaluation; `orders prodid length46 storage` reference snapshots unchanged. `orders.prodid VARCHAR(45)` enforces the 45-character cap. The 46-character value therefore cannot be a valid parent identity, matching the spec's observation that `orders.prodid` retains `VARCHAR(45)` while any valid reference must resolve to a `product.pid` of at most 30 characters. |

Coverage mapping:

| Requirement | Evidence |
| --- | --- |
| US3.1 — Duplicate product/user/transaction identities rejected; original unchanged | R1–R4, R7, R8 (duplicate `product.pid`, `user.email`, `transactions.transaction_id`, generated `transactions.id`, generated `usercart.id`, generated `user_demand.id`) with byte-identical reference snapshots on refusal. |
| US3.2 — Duplicate order `(orderid, prodid)` pair rejected; distinct pairs permitted | R5 (duplicate pair rejected with snapshot unchanged, `(order-1, distinct-product)` and `(order-2, dup-product)` distinct pairs accepted). |
| US3.3 — Orphan line insert or update rejected; `orders.prodid` length matrix | O1 (`orders.prodid` orphan insert/update), L1–L3 (30-character parent match accepted, 31-character no-parent rejected with `fkChildErr`, 46-character value rejected with `tooLongErr`). |
| US3.9 — R1-corrected parent/child orphans refused; duplicate cart `(email, pid)` refused; distinct cart pairs accepted; no demand uniqueness imposed | R6 (duplicate cart pair rejection; distinct cart pairs accepted), R9 (three identical `(email, book_name, author)` demand rows accepted with `COUNT(*) = 3`), O2–O6 (missing `usercart.email`, `usercart.pid`, `user_demand.email`, `transactions.email`, `transactions.pid` parents rejected on insert and update with snapshots unchanged). |
| SC-004 — Historical integrity: no orphan lines, unintended deletions or rewritten identities | Every refused write above is paired with `assertReferenceRowsEqual` before/after, and the accepted distinct pairs and duplicate demand requests are the deliberate inverse: valid distinct pairs and ordinary non-identity writes remain possible. The `historical_product_protection`, `q1_cascade_and_restrict_actions`, `mixed_reference_atomicity` and `identical_enforcement_after_reconstruction` facets of SC-004 are owned by T032–T034. |

Spec-contract preservation:

- No shared constraint from T009/T010/T016–T019 was weakened to make these subtests pass. `migrations/schema/000001_complete_mysql_schema.up.sql` continues to enforce `product.pid VARCHAR(30) PRIMARY KEY`, `user.email VARCHAR(100) PRIMARY KEY`, composite `(orderid, prodid)` for `orders`, the `uq_usercart_email_pid` cart-pair uniqueness, the `uq_transactions_transaction_id` transaction-id uniqueness, the six named FKs (`productid`, `fk_usercart_email`, `fk_usercart_pid`, `fk_user_demand_email`, `fk_transactions_email`, `fk_transactions_pid`) with the Q1 action matrix, and `orders.prodid VARCHAR(45)` (the schema keeps the child column at the recorded legacy width even though a valid parent reference must resolve to a `product.pid` of at most 30 characters).
- The `duplicateErrSpec` (1062/23000), `fkChildErr` (1452/23000) and `tooLongErr` (1406/22001) error class/SQLSTATE expectations match the shared `expectedSQLError` specifications used elsewhere in `tests/schema/references_test.go`, `tests/schema/values_test.go` and the T028/T029 evidence sections above.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard (no `SHOW TABLES` rows), so the schema-only track and all synthetic product/user/order/cart/demand/transaction rows seeded by the two subtests are cleaned up through `DROP TABLE` on the disposable namespace. No forced history or hand-applied DDL was used, no `migrate.Force` was invoked, and no push, merge or deployment was performed. Passwords in any DSN above are redacted.

## T032 Historical-Product Protection Against the Schema-Only Track

Changelog:

- 2026-11-24 — Executed the `historical_product_protection` subtest of `TestSchemaReferences` (`tests/schema/references_test.go`) against the schema-only track applied by `migrations/schema/000001_complete_mysql_schema.up.sql` on the local disposable MySQL 8.4.2 instance (`127.0.0.1:3307`, `shopping-cart` namespace, `utf8mb4_0900_bin`, STRICT_ALL_TABLES/NO_ENGINE_SUBSTITUTION/NO_AUTO_VALUE_ON_ZERO, UTC, `foreign_key_checks=1`). The subtest rejects a referenced product's delete and `pid` identity update under the `productid` `NO ACTION/NO ACTION` relationship with one and with several order references, continues to reject both after all-but-the-last reference are removed, and permits each operation in separate fixtures only after the last restricting reference is removed. The non-identity `pname`/`pquantity` update stays accepted throughout. Every refused write is paired with `assertReferenceRowsEqual` before/after through `assertProductParentMutationRefused`, so the six reference-table snapshots (`product`, `user`, `orders`, `usercart`, `user_demand`, `transactions`) remain byte-identical across the delete and identity-update refusals. No source change was required for T032; the shared `orders.productid` FK and `product` definition were authored once in T016–T019 (`migrations/schema/000001_complete_mysql_schema.up.sql`) and `assertHistoricalProductProtection` was authored in T009 and demonstrated red in T010. The `duplicate_identities_and_valid_distinct_pairs` and `foreign_key_orphan_refusals` subtests are owned by T031; the `q1_cascade_and_restrict_actions`, `mixed_reference_atomicity` and `identical_enforcement_after_reconstruction` subtests are owned by T033–T034 and are not re-executed here.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
gofmt -l tests/schema/
go vet ./tests/schema
```

Exit status: `0` for each command (ping returns `mysqld is alive`; `TestEnvironment` passes; `gofmt -l tests/schema/` prints nothing; `go vet ./tests/schema` reports no findings). Passwords are redacted above.

T032 race command for the historical-product protection group against the schema-only track:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaReferences$/^historical_product_protection$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaReferences (0.18s)` with `--- PASS: TestSchemaReferences/historical_product_protection (0.16s)`. The race build runs against the same local disposable MySQL 8.4.2 instance that passed `TestEnvironment` above; no race report was produced.

Per-case results (historical-product protection fixtures, each preceded by `resetSchemaReferences` so every fixture starts from an empty `shopping-cart` namespace and each refusal is wrapped in `assertProductParentMutationRefused`, which snapshots the six reference tables through `snapshotReferenceRows` before issuing the `DELETE FROM product` and `UPDATE product SET pid` statements and then re-snapshots them through `assertReferenceRowsEqual`):

| ID | US3.* | Fixture / scenario | Refused writes | Permitted writes | Observed outcome |
| --- | --- | --- | --- | --- | --- |
| H1 | US3.4, US3.5 | `history-one`: product seeded, then `orders ('history-order-1','history-one')` seeded (one order reference) | `DELETE FROM product WHERE pid = 'history-one'`; `UPDATE product SET pid = 'history-one-new' WHERE pid = 'history-one'` | — | Both statements rejected with `fkParentErr` (1451/23000); `one order reference delete` and `one order reference identity update` snapshots byte-identical before/after through `assertReferenceRowsEqual`. |
| H2 | US3.4, US3.5 | `history-many`: product seeded, then two order references `('history-order-a','history-many')` and `('history-order-b','history-many')` | `DELETE FROM product WHERE pid = 'history-many'`; `UPDATE product SET pid = 'history-many-new' WHERE pid = 'history-many'` | — | Both statements rejected with `fkParentErr` (1451/23000); `several order references delete` and `several order references identity update` snapshots byte-identical before/after. |
| H3 | US3.6 | `history-many` after `DELETE FROM orders WHERE orderid = 'history-order-a' AND prodid = 'history-many'` removes one of the two references, leaving `('history-order-b','history-many')` as the last remaining order reference | `DELETE FROM product WHERE pid = 'history-many'`; `UPDATE product SET pid = 'history-many-new' WHERE pid = 'history-many'` | `DELETE FROM orders WHERE orderid = 'history-order-a' AND prodid = 'history-many'` | Removal of the non-last order reference is accepted (`assertExecOK`). The subsequent product delete and identity update are both rejected with `fkParentErr` (1451/23000); `last remaining order reference delete` and `last remaining order reference identity update` snapshots byte-identical before/after. |
| H4 | US3.6 | Separate fixture `history-delete-free`: product seeded, single order reference `('history-delete-last','history-delete-free')` seeded, then the last order reference removed. No Q1 cart/transaction reference exists in this fixture, so no other restricting reference remains. | — | `DELETE FROM orders WHERE orderid = 'history-delete-last' AND prodid = 'history-delete-free'`; `DELETE FROM product WHERE pid = 'history-delete-free'` | Both statements accepted (`assertExecOK`); `SELECT COUNT(*) FROM product WHERE pid = 'history-delete-free'` returns `0`. The product delete is permitted only once the last reference and every other restricting reference are gone. |
| H5 | US3.6 | Separate fixture `history-update-free`: product seeded, single order reference `('history-update-last','history-update-free')` seeded, then the last order reference removed. No Q1 cart/transaction reference exists in this fixture, so no other restricting reference remains. | — | `DELETE FROM orders WHERE orderid = 'history-update-last' AND prodid = 'history-update-free'`; `UPDATE product SET pid = 'history-update-new' WHERE pid = 'history-update-free'` | Both statements accepted (`assertExecOK`); `SELECT COUNT(*) FROM product WHERE pid = 'history-update-new'` returns `1`. The product identity update is permitted only once the last reference and every other restricting reference are gone, in a fixture separate from the delete fixture. |
| H6 | US3.7 | `history-metadata`: product seeded, `orders ('history-metadata-order','history-metadata')` seeded (one order reference kept throughout), then `UPDATE product SET pname = 'permitted metadata', pquantity = -4 WHERE pid = 'history-metadata'` | — | `UPDATE product SET pname = 'permitted metadata', pquantity = -4 WHERE pid = 'history-metadata'` | Accepted (`assertExecOK`); `queryProductValues` reports `pname = 'permitted metadata'` and `pquantity = -4` while `pid` stays `'history-metadata'` and the `orders ('history-metadata-order','history-metadata')` reference is retained. The optional non-identity fields are therefore updatable while the product is still referenced, confirming the `productid` `NO ACTION` relationship only blocks identity changes and deletions. |

Coverage mapping:

| Requirement | Evidence |
| --- | --- |
| US3.4 — Product referenced by one or more order lines: deletion rejected, product and every referencing line unchanged | H1 (one reference, delete), H2 (several references, delete), H3 (last remaining reference after partial removal, delete). Each refusal wraps `assertReferenceRowsEqual` around the `DELETE FROM product` statement to prove `product`, `orders` (and the four other reference tables) are byte-identical before/after. |
| US3.5 — Product referenced by one or more order lines: identity change rejected, neither product nor line identities rewritten | H1 (one reference, identity update), H2 (several references, identity update), H3 (last remaining reference after partial removal, identity update). Each refusal wraps `assertReferenceRowsEqual` around the `UPDATE product SET pid` statement to prove `product.pid` and every `orders.prodid` reference are byte-identical before/after. |
| US3.6 — Several references: all but one removed → delete/id update still blocked; last removed → both operations permitted in separate fixtures subject to Q1's other references | H3 (all but one removed → both operations still rejected with snapshots unchanged); H4 (separate fixture — only delete; last order reference removed; no Q1 cart/transaction reference remaining → `DELETE FROM product` accepted); H5 (separate fixture — only identity update; last order reference removed; no Q1 cart/transaction reference remaining → `UPDATE product SET pid` accepted). Delete and identity update are deliberately exercised in distinct fixtures (`history-delete-free` vs. `history-update-free`) rather than in a single fixture, so no fixture chains the two successful operations. Q1's other restricting references (`fk_usercart_pid` RESTRICT on delete, `fk_transactions_pid` RESTRICT on delete) are proven separately in T033's `q1_cascade_and_restrict_actions`; H4/H5 do not include a Q1 child because the fixture isolates the `productid` NO ACTION facet. |
| US3.7 — Referenced product: optional non-identity field change permitted; line still references the same product | H6 (`pname` and `pquantity` updated to permitted non-identity values while the `('history-metadata-order','history-metadata')` reference is retained; `queryProductValues` confirms the new metadata and the `orders.prodid` reference still resolves to the same `product.pid`). No `pid` is rewritten, so the `productid` NO ACTION relationship is not exercised; the subtest confirms the rule does not block valid non-identity updates. |
| SC-004 — Historical integrity: no orphan lines, unintended deletions or rewritten identities; ordinary product-field changes remain possible | Every refused `DELETE FROM product` and `UPDATE product SET pid` in H1–H3 is paired with `assertReferenceRowsEqual` before/after, so no `orders.prodid` line is ever orphaned or rewritten, no product is ever deleted or re-identified by a refused write, and no unrelated `user`/`usercart`/`user_demand`/`transactions` row is mutated. The permitted deletion and identity update in H4/H5 only occur after the last reference is removed, so no historical reference is broken. The permitted non-identity metadata update in H6 is the SC-004 inverse: ordinary product-field changes remain possible while the product is still referenced. The `q1_cascade_and_restrict_actions`, `mixed_reference_atomicity` and `identical_enforcement_after_reconstruction` facets of SC-004 are owned by T033–T034. |

Spec-contract preservation:

- No shared constraint from T009/T010/T016–T019 was weakened to make this subtest pass. `migrations/schema/000001_complete_mysql_schema.up.sql` continues to define the `productid` foreign key as `FOREIGN KEY (prodid) REFERENCES product(pid) ON DELETE NO ACTION ON UPDATE NO ACTION`, preserving FR-007's rejection of both identity changes and deletions of referenced products. The `product.pid VARCHAR(30) PRIMARY KEY` and `orders.prodid VARCHAR(45) NOT NULL` definitions are unchanged, so H1–H5 exercise the same parent/child length pair recorded in T031's L1/L2/L3 matrix.
- The `fkParentErr` (1451/23000) error class/SQLSTATE expectation matches the shared `expectedSQLError` specification used in the `q1_cascade_and_restrict_actions` and `mixed_reference_atomicity` subtests of `tests/schema/references_test.go`.
- `assertProductParentMutationRefused` and `assertReferenceRowsEqual` were introduced in T009 and have not been modified; they still snapshot `product`, `user`, `orders`, `usercart`, `user_demand` and `transactions` through `snapshotReferenceRows`'s `GROUP_CONCAT(JSON_ARRAY(...))` queries.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard (no `SHOW TABLES` rows), so the schema-only track and all synthetic product/order rows seeded by the six historical-product fixtures above are cleaned up through `DROP TABLE` on the disposable namespace. No forced history or hand-applied DDL was used, no `migrate.Force` was invoked, and no push, merge or deployment was performed. Passwords in any DSN above are redacted.

## T033 Q1-Action and Mixed-Reference Groups Against the Schema-Only Track

Changelog:

- 2026-11-24 — Executed the `q1_cascade_and_restrict_actions` and `mixed_reference_atomicity` subtests of `TestSchemaReferences` (`tests/schema/references_test.go`) against the schema-only track applied by `migrations/schema/000001_complete_mysql_schema.up.sql` on the local disposable MySQL 8.4.2 instance (`127.0.0.1:3307`, `shopping-cart` namespace, `utf8mb4_0900_bin`, STRICT_ALL_TABLES/NO_ENGINE_SUBSTITUTION/NO_AUTO_VALUE_ON_ZERO, UTC, `foreign_key_checks=1`). The Q1-action group separately proves (a) user delete cascading to `usercart` and `user_demand` via `fk_usercart_email`/`fk_user_demand_email` CASCADE, (b) `transactions`→`user` and `usercart`/`transactions`→`product` delete restrictions via `fk_transactions_email`, `fk_usercart_pid` and `fk_transactions_pid` RESTRICT, and (c) `user.email` and `product.pid` identity updates cascading through `fk_usercart_email`, `fk_user_demand_email`, `fk_transactions_email`, `fk_usercart_pid` and `fk_transactions_pid` UPDATE CASCADE while each refused delete leaves every reference-table snapshot byte-identical. The mixed-reference group separately proves that `orders.productid` NO ACTION overrides the Q1 UPDATE CASCADE rewrite when a cart/transaction product is also referenced by an order line, and that a `transactions`→`user` RESTRICT refusal of a user delete leaves the cart/demand cascades untouched (no partial cascade is observed, because the restriction is raised before any cascade runs). Every refused write is paired with `assertReferenceRowsEqual` before/after, so the six reference-table snapshots (`product`, `user`, `orders`, `usercart`, `user_demand`, `transactions`) remain byte-identical across each restriction and NO ACTION rejection. No source change was required for T033; the shared Q1 FK action matrix (`fk_usercart_email` CASCADE/CASCADE, `fk_usercart_pid` CASCADE on update + RESTRICT on delete, `fk_user_demand_email` CASCADE/CASCADE, `fk_transactions_email` CASCADE on update + RESTRICT on delete, `fk_transactions_pid` CASCADE on update + RESTRICT on delete, and `productid` NO ACTION/NO ACTION) was authored once in T016–T019 (`migrations/schema/000001_complete_mysql_schema.up.sql`) and `assertQ1Actions`/`assertMixedReferenceAtomicity` were authored in T009 and demonstrated red in T010. The `duplicate_identities_and_valid_distinct_pairs` and `foreign_key_orphan_refusals` subtests are owned by T031, the `historical_product_protection` subtest is owned by T032, and the `identical_enforcement_after_reconstruction` subtest is owned by T034; none of them is re-executed here.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
gofmt -l tests/schema/
go vet ./tests/schema
```

Exit status: `0` for each command (ping returns `mysqld is alive`; `TestEnvironment` passes; `gofmt -l tests/schema/` prints nothing; `go vet ./tests/schema` reports no findings). Passwords are redacted above.

T033 race command for the Q1-action and mixed-reference groups against the schema-only track:

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaReferences$/^(q1_cascade_and_restrict_actions|mixed_reference_atomicity)$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaReferences (0.82s)` with every selected subtest PASS:

```
--- PASS: TestSchemaReferences/q1_cascade_and_restrict_actions (0.26s)
--- PASS: TestSchemaReferences/mixed_reference_atomicity (0.14s)
```

The race build runs against the same local disposable MySQL 8.4.2 instance that passed `TestEnvironment` above; no race report was produced.

Per-case results (`q1_cascade_and_restrict_actions`, each fixture preceded by `resetSchemaReferences` so every fixture starts from an empty `shopping-cart` namespace; every refused `DELETE FROM` is wrapped by `snapshotReferenceRows`/`assertReferenceRowsEqual` which compares the six reference tables through `GROUP_CONCAT(JSON_ARRAY(...))`; every permitted cascade is confirmed by `assertCount`):

| ID | US3.* | Fixture / scenario | Refused writes | Permitted writes | Observed outcome |
| --- | --- | --- | --- | --- | --- |
| Q1 | US3.10 | `q1-product`, `q1@example.test`; `usercart (q1@example.test, q1-product)` and `user_demand (q1@example.test, q1-book)` seeded; no transaction reference | — | `DELETE FROM user WHERE email = 'q1@example.test'` | Accepted (`assertExecOK`). Before: `usercart` holds one row `(..,q1@example.test,q1-product,..)` and `user_demand` holds one row `(..,q1@example.test,q1-book,..)`. After: `SELECT COUNT(*) FROM usercart WHERE email = 'q1@example.test'` returns `0` and `SELECT COUNT(*) FROM user_demand WHERE email = 'q1@example.test'` returns `0`. Both child rows are removed by `fk_usercart_email`/`fk_user_demand_email` ON DELETE CASCADE because no `transactions` row restricts the parent delete. |
| Q2 | US3.10 | `q1-user-restrict-product`, `q1-restrict@example.test`; `transactions (q1-user-restrict-tx, q1-restrict@example.test, q1-user-restrict-product, 1)` seeded | `DELETE FROM user WHERE email = 'q1-restrict@example.test'` | — | Rejected with `fkParentErr` (1451/23000); `assertReferenceRowsEqual(before, after, "transaction restricts user delete")` reports the six reference-table snapshots byte-identical before/after — the single `user` row and the single `transactions` row are preserved, and no cart/demand rows exist in this fixture to be mutated by a partial cascade. `fk_transactions_email` ON DELETE RESTRICT therefore blocks the parent delete before any other FK action runs. |
| Q3 | US3.10 | `q1-product-restrict`, `q1-product-restrict@example.test`; `usercart (q1-product-restrict@example.test, q1-product-restrict)` seeded | `DELETE FROM product WHERE pid = 'q1-product-restrict'` | — | Rejected with `fkParentErr` (1451/23000); `assertReferenceRowsEqual(before, after, "cart restricts product delete")` reports the six reference-table snapshots byte-identical before/after. `fk_usercart_pid` ON DELETE RESTRICT blocks the parent delete; the single `product` and single `usercart` rows are preserved. |
| Q4 | US3.10 | `q1-tx-product-restrict`, `q1-tx-product-restrict@example.test`; `transactions (q1-product-restrict-tx, q1-tx-product-restrict@example.test, q1-tx-product-restrict, 1)` seeded; no cart reference in this fixture | `DELETE FROM product WHERE pid = 'q1-tx-product-restrict'` | — | Rejected with `fkParentErr` (1451/23000); `assertReferenceRowsEqual(before, after, "transaction restricts product delete")` reports the six reference-table snapshots byte-identical before/after. `fk_transactions_pid` ON DELETE RESTRICT blocks the parent delete; the single `product` and single `transactions` rows are preserved. A separate fixture from Q3 so the RESTRICT facets of `fk_usercart_pid` and `fk_transactions_pid` are proven independently. |
| Q5 | US3.11 | `q1-cascade-product`, `q1-cascade@example.test`; `usercart (q1-cascade@example.test, q1-cascade-product)`, `user_demand (q1-cascade@example.test, q1-cascade-book)` and `transactions (q1-cascade-tx, q1-cascade@example.test, q1-cascade-product, 1)` seeded | — | `UPDATE user SET email = 'q1-cascade-new@example.test' WHERE email = 'q1-cascade@example.test'` | Accepted (`assertExecOK`). Before: `usercart`, `user_demand` and `transactions` each hold one row keyed by `email = 'q1-cascade@example.test'`. After: `SELECT COUNT(*) FROM usercart WHERE email = 'q1-cascade-new@example.test'` returns `1`, `SELECT COUNT(*) FROM user_demand WHERE email = 'q1-cascade-new@example.test'` returns `1` and `SELECT COUNT(*) FROM transactions WHERE email = 'q1-cascade-new@example.test'` returns `1`. The valid, unused new identity cascades through `fk_usercart_email`, `fk_user_demand_email` and `fk_transactions_email` ON UPDATE CASCADE; no NO ACTION relationship (Q1 has none for account→cart/demand/transactions) overrides the Q1 rule. |
| Q6 | US3.11 | Same fixture as Q5, after Q5's user cascade; now updates `product.pid` while cart and transaction children exist and no `orders` reference exists | — | `UPDATE product SET pid = 'q1-cascade-product-new' WHERE pid = 'q1-cascade-product'` | Accepted (`assertExecOK`). Before: `usercart.pid = 'q1-cascade-product'` on one row; `transactions.pid = 'q1-cascade-product'` on one row; no `orders` row references `q1-cascade-product`. After: `SELECT COUNT(*) FROM usercart WHERE pid = 'q1-cascade-product-new'` returns `1` and `SELECT COUNT(*) FROM transactions WHERE pid = 'q1-cascade-product-new'` returns `1`. The valid, unused new `product.pid` cascades through `fk_usercart_pid` and `fk_transactions_pid` ON UPDATE CASCADE because `productid` has no order line in this fixture; the FR-007 NO ACTION override is deliberately exercised separately in M1. |

Per-case results (`mixed_reference_atomicity`, each fixture preceded by `resetSchemaReferences`; every refused write is wrapped by `snapshotReferenceRows`/`assertReferenceRowsEqual`):

| ID | US3.* | Fixture / scenario | Refused writes | Permitted writes | Observed outcome |
| --- | --- | --- | --- | --- | --- |
| M1 | US3.11 | `mixed-product`, `mixed@example.test`; `orders (mixed-order, mixed-product)`, `usercart (mixed@example.test, mixed-product)` and `transactions (mixed-tx, mixed@example.test, mixed-product, 1)` seeded — the product is simultaneously referenced by an order line and by Q1 cart and transaction children | `UPDATE product SET pid = 'mixed-product-new' WHERE pid = 'mixed-product'` | — | Rejected with `fkParentErr` (1451/23000); `assertReferenceRowsEqual(before, after, "order NO ACTION blocks mixed product rewrite")` reports the six reference-table snapshots byte-identical before/after — `product.pid` remains `'mixed-product'`, `orders.prodid` remains `'mixed-product'`, `usercart.pid` remains `'mixed-product'` and `transactions.pid` remains `'mixed-product'`. The `productid` ON UPDATE NO ACTION (FR-007) override fires before any Q1 cascade through `fk_usercart_pid`/`fk_transactions_pid`, so no partial rewrite leaves cart/transaction children pointing at the attempted new `'mixed-product-new'` identity and no `orders.prodid` is rewritten. |
| M2 | US3.10 | `mixed-user-product`, `mixed-user@example.test`; `usercart (mixed-user@example.test, mixed-user-product)`, `user_demand (mixed-user@example.test, mixed-user-book)` and `transactions (mixed-user-tx, mixed-user@example.test, mixed-user-product, 1)` seeded — the user is simultaneously referenced by Q1 cart, demand and transaction children | `DELETE FROM user WHERE email = 'mixed-user@example.test'` | — | Rejected with `fkParentErr` (1451/23000); `assertReferenceRowsEqual(before, after, "transaction restrict leaves user cascades untouched")` reports the six reference-table snapshots byte-identical before/after — the single `user` row, the single `usercart` row, the single `user_demand` row and the single `transactions` row are all preserved. `fk_transactions_email` ON DELETE RESTRICT raises its refusal before the `fk_usercart_email`/`fk_user_demand_email` CASCADE effects are applied; no partial cascade deletes the cart or demand children, so the write is atomic. |

Coverage mapping:

| Requirement | Evidence |
| --- | --- |
| US3.9 — R1-corrected parent/child orphans refused; duplicate cart `(email, pid)` refused; distinct cart pairs accepted; no demand uniqueness imposed | Owned by T031; T033 does not re-exercise the orphan-insert/duplicate-cart cases. The Q1 and mixed fixtures above seed only valid parents/children, so no orphan write is attempted; `assertReferenceRowsEqual` confirms the orphan surface is not re-introduced by any refused delete/update. |
| US3.10 — User referenced only by cart and demand: cascade delete; transaction referencing user or cart/transaction referencing product: delete restricted with all rows unchanged. Separate fixtures for cascade and restriction cases | Q1 (user delete cascades to `usercart` and `user_demand` via `fk_usercart_email`/`fk_user_demand_email` CASCADE when no transaction reference exists), Q2 (transaction-to-user RESTRICT in a fixture with no cart/demand rows, so the restriction is observed in isolation; six reference-table snapshots unchanged), Q3 (cart-to-product RESTRICT in a fixture with no transaction reference; snapshots unchanged), Q4 (transaction-to-product RESTRICT in a fixture with no cart reference; snapshots unchanged — proven separately from Q3 so the two product-delete RESTRICT facets are independent). M2 provides the composite confirmation: with cart/demand/transaction children present, the `fk_transactions_email` RESTRICT refusal leaves the cart and demand rows untouched (no partial cascade). |
| US3.11 — Referenced user.email or product.pid updated to a valid unused identity: Q1 children cascade-update; this default does not override FR-007's NO ACTION from order lines to product identities | Q5 (user email update cascades to `usercart`, `user_demand` and `transactions` via `fk_usercart_email`, `fk_user_demand_email` and `fk_transactions_email` ON UPDATE CASCADE; all three post-counts are `1`), Q6 (product.pid update cascades to `usercart` and `transactions` via `fk_usercart_pid` and `fk_transactions_pid` ON UPDATE CASCADE when no order line references the product; both post-counts are `1`). M1 provides the NO ACTION override: when an `orders` line references the same product, `productid` ON UPDATE NO ACTION rejects the `UPDATE product SET pid` with the six reference-table snapshots byte-identical before/after, so no partial Q1 cascade applies. The FR-007 exception therefore remains in force even when valid Q1 cascade targets exist. |
| SC-004 — Historical integrity: no orphan lines, unintended deletions or rewritten identities; valid distinct pairs and ordinary product-field changes remain possible | Every refused `DELETE FROM user`, `DELETE FROM product` and `UPDATE product SET pid` in Q2–Q4 and M1–M2 is paired with `assertReferenceRowsEqual` before/after through `snapshotReferenceRows`'s `GROUP_CONCAT(JSON_ARRAY(...))` queries over `product`, `user`, `orders`, `usercart`, `user_demand` and `transactions`, so no orphan line is created, no unintended row is deleted and no identity is rewritten by a refused write. The permitted cascades in Q1, Q5 and Q6 are the SC-004 inverse: ordinary user-delete-under-Q1-cascade, user-email-update and product-pid-update remain possible when no restricting relationship (`fk_transactions_*` RESTRICT, `fk_usercart_pid` RESTRICT, `productid` NO ACTION) applies. The `identity_integrity_against_schema_only_track`, `historical_product_protection` and `identical_enforcement_after_reconstruction` facets of SC-004 are owned by T031, T032 and T034 respectively. |

Spec-contract preservation:

- No shared constraint from T009/T010/T016–T019 was weakened to make these subtests pass. `migrations/schema/000001_complete_mysql_schema.up.sql` continues to define `fk_usercart_email` and `fk_user_demand_email` as `ON DELETE CASCADE ON UPDATE CASCADE`, `fk_usercart_pid`, `fk_transactions_email` and `fk_transactions_pid` as `ON DELETE RESTRICT ON UPDATE CASCADE`, and the `productid` FK as `ON DELETE NO ACTION ON UPDATE NO ACTION`. FR-007 (order-line NO ACTION) and FR-009 (Q1 CASCADE-on-update, CASCADE-on-delete for cart/demand email, RESTRICT-on-delete for cart/transaction pid and transaction email) are enforced exactly as recorded.
- The `fkParentErr` (1451/23000) error class/SQLSTATE expectation matches the shared `expectedSQLError` specification used in `tests/schema/references_test.go` for the `historical_product_protection` and `identical_enforcement_after_reconstruction` subtests above and below.
- `assertQ1Actions`, `assertMixedReferenceAtomicity`, `snapshotReferenceRows` and `assertReferenceRowsEqual` were introduced in T009 and have not been modified; they still snapshot `product`, `user`, `orders`, `usercart`, `user_demand` and `transactions` through `snapshotReferenceRows`'s `GROUP_CONCAT(JSON_ARRAY(...))` queries ordered by their composite or generated identities.
- `product.pid VARCHAR(30) PRIMARY KEY`, `user.email VARCHAR(100) PRIMARY KEY`, composite `(orderid, prodid)` for `orders`, `uq_usercart_email_pid`, `uq_transactions_transaction_id` and all six named FKs are unchanged from T031/T032's observations; Q1/M1/M2 exercise the same shared schema.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SHOW TABLES;"
```

Exit status: `0`. The disposable `shopping-cart` namespace is empty after the discard (no `SHOW TABLES` rows), so the schema-only track and all synthetic product/user/order/cart/demand/transaction rows seeded by the Q1 and mixed-reference fixtures above are cleaned up through `DROP TABLE` on the disposable namespace. No forced history or hand-applied DDL was used, no `migrate.Force` was invoked, and no push, merge or deployment was performed. Passwords in any DSN above are redacted.

## T034 Complete Reference Matrix Replay After Populated Schema Lifecycle

Changelog:

- 2026-11-24 — Re-ran the complete `TestSchemaReferences` matrix (`tests/schema/references_test.go`) after the populated schema lifecycle subtest `TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior` (`tests/schema/lifecycle_test.go`) against the local disposable MySQL 8.4.2 instance (`127.0.0.1:3307`, `shopping-cart` namespace, `utf8mb4_0900_bin`, STRICT_ALL_TABLES/NO_ENGINE_SUBSTITUTION/NO_AUTO_VALUE_ON_ZERO, UTC, `foreign_key_checks=1`). The lifecycle subtest applies the `sentinel_owned` fixture at version 1, seeds its `unrelated-sentinel` row, applies the schema up, populates synthetic business rows, drops the schema through its bounded down migration, reapplies the schema up and confirms the sentinel definition and row data are byte-identical before and after (`assertSentinelUnchanged`). The complete `TestSchemaReferences` matrix (`duplicate_identities_and_valid_distinct_pairs`, `foreign_key_orphan_refusals` with all six sub-cases, `historical_product_protection`, `q1_cascade_and_restrict_actions`, `mixed_reference_atomicity`, and the US3.8 `identical_enforcement_after_reconstruction` subtest) then runs on the still-present sentinel fixture and the recreated schema, asserts `@@session.foreign_key_checks = 1` through `assertForeignKeyChecksEnabled` on the newly acquired connection after a further schema down/up from inside the reference suite, and re-runs `assertReferenceEnforcementMatrix` against the reconstructed schema. Every T031/T032/T033 subtest returns identical PASS results to those recorded in the preceding T031/T032/T033 sections. No source change was required for T034; the shared schema, FK action matrix and reference assertions were authored once in T009/T016–T019 and the `identical_enforcement_after_reconstruction` subtest was authored in T009 and demonstrated red in T010. The `duplicate_identities_and_valid_distinct_pairs` and `foreign_key_orphan_refusals` subtests remain owned by T031, `historical_product_protection` by T032, and `q1_cascade_and_restrict_actions` and `mixed_reference_atomicity` by T033; they are re-executed here to prove the schema recreation preserves their enforcement, not re-attributed to T034.

Working fixture/environment check:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysqladmin --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx ping
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestEnvironment$' -count=1 -v
gofmt -l tests/schema/
go vet ./tests/schema
```

Exit status: `0` for each command (ping returns `mysqld is alive`; `TestEnvironment` passes; `gofmt -l tests/schema/` prints nothing; `go vet ./tests/schema` reports no findings). Passwords are redacted above.

T034 sequence step 1 — empty disposable namespace before the populated lifecycle:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SET FOREIGN_KEY_CHECKS=1; SHOW TABLES;"
```

Exit status: `0`. `SHOW TABLES` returns no rows. The `SET FOREIGN_KEY_CHECKS=0`/`SET FOREIGN_KEY_CHECKS=1` wrapper is session-only on the disposable discard session so the previous T033 external-reference-style partial state can be dropped without a hand-applied DDL change; it is immediately re-enabled before the lifecycle subtest acquires any new connection. The session that will run the Go tests is a separate newly acquired pool managed by `openConfiguredDB`, which inherits `foreign_key_checks=1` through the runtime configuration selected by T005/T006.

T034 sequence step 2 — populated schema lifecycle (sentinel preservation, bounded schema down/up on populated business rows, `assertLifecycleBehavior` on a reapplied schema):

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestSchemaLifecycle$/^populated_bounded_down_up_keeps_sentinel_and_behavior$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaLifecycle (1.18s)` with `--- PASS: TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior (1.17s)`.

T034 sequence step 3 — sentinel definition/data and `foreign_key_checks` snapshot on a newly acquired session immediately after the populated lifecycle:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "SHOW TABLES; SHOW CREATE TABLE sentinel_owned\\G SELECT * FROM sentinel_owned ORDER BY id; SELECT @@session.foreign_key_checks; SELECT version, dirty FROM schema_migrations_fixture_sentinel; SELECT version, dirty FROM schema_migrations_schema;"
```

Exit status: `0`. Result summary:

- `SHOW TABLES` returns `orders`, `product`, `schema_migrations_fixture_sentinel`, `schema_migrations_schema`, `sentinel_owned`, `transactions`, `user`, `user_demand`, `usercart` — only owned schema and sentinel-fixture objects remain.
- `sentinel_owned` definition is `CREATE TABLE sentinel_owned (id BIGINT NOT NULL AUTO_INCREMENT, label VARCHAR(64) COLLATE utf8mb4_0900_bin NOT NULL, PRIMARY KEY (id), UNIQUE KEY uq_sentinel_owned_label (label)) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin` — byte-identical to the definition the sentinel fixture applies at version 1.
- `SELECT * FROM sentinel_owned` returns a single row `(1, unrelated-sentinel)` — byte-identical to the row `seedSentinel` inserts before the schema down/up.
- `SELECT @@session.foreign_key_checks` returns `1` on this newly opened mysql session, before any Go test runs over it.
- `schema_migrations_fixture_sentinel` version/dirty = `1/0`; `schema_migrations_schema` version/dirty = `1/0` — both tracks clean, no dirty state left by the populated lifecycle.

T034 sequence step 4 — complete reference matrix against the schema reapplied by the populated lifecycle (the reference suite's own `identical_enforcement_after_reconstruction` subtest additionally performs a bounded schema down/up and re-asserts `@@session.foreign_key_checks = 1` on the newly acquired connection through `assertForeignKeyChecksEnabled` before re-running `assertReferenceEnforcementMatrix`):

```sh
SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test ./tests/schema -run '^TestSchemaReferences$' -count=1 -v
```

Exit status: `0`. Result: `--- PASS: TestSchemaReferences (1.52s)` with every subtest PASS:

```
--- PASS: TestSchemaReferences/duplicate_identities_and_valid_distinct_pairs (0.13s)
--- PASS: TestSchemaReferences/foreign_key_orphan_refusals (0.22s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/orders.prodid (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/usercart.email (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/usercart.pid (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/user_demand.email (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/transactions.email (0.02s)
    --- PASS: TestSchemaReferences/foreign_key_orphan_refusals/transactions.pid (0.02s)
--- PASS: TestSchemaReferences/historical_product_protection (0.18s)
--- PASS: TestSchemaReferences/q1_cascade_and_restrict_actions (0.21s)
--- PASS: TestSchemaReferences/mixed_reference_atomicity (0.12s)
--- PASS: TestSchemaReferences/identical_enforcement_after_reconstruction (0.64s)
```

T034 sequence step 5 — sentinel definition/data and `foreign_key_checks` re-check on another newly acquired session after the complete reference matrix (and its internal schema down/up) finished:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "SHOW CREATE TABLE sentinel_owned\\G SELECT * FROM sentinel_owned ORDER BY id; SELECT @@session.foreign_key_checks; SELECT version, dirty FROM schema_migrations_fixture_sentinel; SELECT version, dirty FROM schema_migrations_schema;"
```

Exit status: `0`. Result summary:

- `sentinel_owned` definition is `CREATE TABLE sentinel_owned (id BIGINT NOT NULL AUTO_INCREMENT, label VARCHAR(64) COLLATE utf8mb4_0900_bin NOT NULL, PRIMARY KEY (id), UNIQUE KEY uq_sentinel_owned_label (label)) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin` — byte-identical to the step-3 snapshot and to the sentinel fixture version 1 definition.
- `SELECT * FROM sentinel_owned` returns `(1, unrelated-sentinel)` — byte-identical to the step-3 snapshot.
- `SELECT @@session.foreign_key_checks` returns `1` on this third newly opened mysql session, after the reference matrix (including the internal schema down/up) has run.
- `schema_migrations_fixture_sentinel` version/dirty = `1/0`; `schema_migrations_schema` version/dirty = `1/0` — both tracks remain clean.

T034 race commands for the populated lifecycle and the complete reference matrix (new disposable instance after step 5 was discarded):

```sh
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaLifecycle$/^populated_bounded_down_up_keeps_sentinel_and_behavior$' -count=1 -v
PATH="$PWD/.tools/libroot/usr/bin:$PATH" LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu" CC="gcc --sysroot=$PWD/.tools/libroot -B $PWD/.tools/libroot/usr/lib/x86_64-linux-gnu -B $PWD/.tools/libroot/lib/x86_64-linux-gnu" CGO_ENABLED=1 SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&loc=UTC' go test -race ./tests/schema -run '^TestSchemaReferences$' -count=1 -v
```

Exit status: `0` for each. Result: `--- PASS: TestSchemaLifecycle (1.19s)` with `--- PASS: TestSchemaLifecycle/populated_bounded_down_up_keeps_sentinel_and_behavior (1.19s)`; `--- PASS: TestSchemaReferences (1.54s)` with every subtest PASS, identical names and ordering to the non-race run above. The race builds run against the same local disposable MySQL 8.4.2 instance that passed `TestEnvironment`; no race report was produced.

Per-case results (US3.8 identity reconstruction, inside `tests/schema/references_test.go`'s `identical_enforcement_after_reconstruction` subtest, exercised both standalone during T031–T033 regression and immediately after the T034 step-2 populated lifecycle — every refused write is wrapped by `snapshotReferenceRows`/`assertReferenceRowsEqual` which compares the six reference tables `product`, `user`, `orders`, `usercart`, `user_demand`, `transactions` through `GROUP_CONCAT(JSON_ARRAY(...))` queries, and both pre-reconstruction and post-reconstruction enforcement calls use `assertReferenceEnforcementMatrix`):

| ID | US3.* | Fixture / scenario | Refused writes | Permitted writes | Observed outcome |
| --- | --- | --- | --- | --- | --- |
| I1 | US3.8 | `matrix-product`, `matrix@example.test` seeded on the schema already applied by the T034 populated lifecycle (step 2) before any further schema down/up — `assertReferenceEnforcementMatrix` call 1 of 2 inside the subtest | `INSERT INTO orders (orderid, prodid) VALUES ('matrix-missing-order', 'matrix-missing-product')`; `UPDATE product SET pid = 'matrix-product-new' WHERE pid = 'matrix-product'` with an `orders (matrix-order, matrix-product)` reference already seeded | `INSERT INTO orders (orderid, prodid) VALUES ('matrix-order', 'matrix-product')`; `UPDATE product SET pname = 'matrix metadata' WHERE pid = 'matrix-product'` | Orphan `orders` insert rejected with `fkChildErr` (1452/23000); referenced `product.pid` update rejected with `fkParentErr` (1451/23000); `matrix orphan order insert` and `matrix protected product update` snapshots byte-identical before/after through `assertReferenceRowsEqual`. Valid `orders` insert on the matching 30-character parent is accepted; `pname` non-identity update is accepted and `queryProductValues` confirms `pname = 'matrix metadata'` on `matrix-product`. |
| I2 | US3.8 | `migrateToVersion(ctx, startupConfig, schemaTrack, 0)` then `migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target)` inside `identical_enforcement_after_reconstruction`, performed on top of the populated-lifecycle-recreated schema | — | Bounded schema down migration `000001_complete_mysql_schema.down.sql`; schema up migration `000001_complete_mysql_schema.up.sql` | Both migrations succeed through `golang-migrate` without `migrate.Force`; `schema_migrations_schema` version/dirty = `1/0` after. The sentinel fixture `schema_migrations_fixture_sentinel` version/dirty stays at `1/0` across the schema down and up, and `sentinel_owned` definition and `(1, unrelated-sentinel)` row remain byte-identical (step-5 snapshot above). |
| I3 | US3.8 | `assertForeignKeyChecksEnabled` on the newly opened pool connection acquired by `db.QueryRowContext(ctx, "SELECT @@session.foreign_key_checks")` after the I2 reconstruction | `foreign_key_checks` disabled (none attempted) | — | `@@session.foreign_key_checks` returns `1` through `assertForeignKeyChecksEnabled`; the Go pool (opened via `openConfiguredDB` with the runtime-selected modes from T005/T006) presents `foreign_key_checks=1` without any manual `SET` statement. |
| I4 | US3.8 | `assertReferenceEnforcementMatrix` call 2 of 2 inside the subtest, immediately after the I2 reconstruction, with the sentinel fixture still at version 1 and the reconstructed schema at target version 1 | `INSERT INTO orders (orderid, prodid) VALUES ('matrix-missing-order', 'matrix-missing-product')`; `UPDATE product SET pid = 'matrix-product-new' WHERE pid = 'matrix-product'` with a reseeded `orders (matrix-order, matrix-product)` reference | `INSERT INTO orders (orderid, prodid) VALUES ('matrix-order', 'matrix-product')`; `UPDATE product SET pname = 'matrix metadata' WHERE pid = 'matrix-product'` | Byte-identical results to I1: orphan `orders` insert rejected with `fkChildErr`; referenced `product.pid` update rejected with `fkParentErr`; matching 30-character `orders` insert and non-identity `pname` update accepted. `queryProductValues` confirms `pname = 'matrix metadata'` again. The enforcement matrix is therefore identical before and after schema reconstruction. |

Per-case results (US3/SC-004 regression — re-executed under the T034 step 4 complete reference matrix after the T034 step 2 populated lifecycle; each row cross-references the authoritative T031/T032/T033 evidence to prove the recreated schema returns identical results):

| ID | US3.* | Subtest | Owning T-ID evidence | T034 observed outcome |
| --- | --- | --- | --- | --- |
| RR1 | US3.1, US3.2, US3.9 | `duplicate_identities_and_valid_distinct_pairs` (R1–R9) | T031 table above | `--- PASS (0.13s)`; R1–R9 outcomes identical to T031 — duplicate `product.pid`, `user.email`, `transactions.transaction_id`, generated `transactions.id`, duplicate order pair, duplicate cart pair, generated `usercart.id`, generated `user_demand.id` rejected with `duplicateErrSpec` (1062/23000); distinct pairs and three identical `(email, book_name, author)` demand rows accepted (`COUNT(*) = 3`). |
| RR2 | US3.3, US3.9 | `foreign_key_orphan_refusals` (O1–O6, L1–L3) | T031 table above | `--- PASS (0.22s)` with all six FK sub-cases PASS; O1–O6 insert/update orphan refusals rejected with `fkChildErr` (1452/23000) and snapshots byte-identical; L1 30-char parent match accepted, L2 31-char no-parent rejected with `fkChildErr`, L3 46-char value rejected with `tooLongErr` (1406/22001). Identical to T031. |
| RR3 | US3.4, US3.5, US3.6, US3.7 | `historical_product_protection` (H1–H6) | T032 table above | `--- PASS (0.18s)`; H1–H3 delete and identity-update refusals rejected with `fkParentErr` (1451/23000) and snapshots byte-identical; H4/H5 permitted only after the last reference is removed in separate fixtures; H6 non-identity `pname`/`pquantity` update accepted. Identical to T032. |
| RR4 | US3.10, US3.11 | `q1_cascade_and_restrict_actions` (Q1–Q6) | T033 table above | `--- PASS (0.21s)`; Q1 user delete cascades to cart/demand; Q2 `fk_transactions_email` RESTRICT blocks user delete; Q3 `fk_usercart_pid` RESTRICT blocks product delete; Q4 `fk_transactions_pid` RESTRICT blocks product delete; Q5 user email update cascades to cart/demand/transactions; Q6 `product.pid` update cascades to cart/transactions (no order reference in that fixture). Identical to T033. |
| RR5 | US3.10, US3.11 | `mixed_reference_atomicity` (M1, M2) | T033 table above | `--- PASS (0.12s)`; M1 `productid` NO ACTION overrides Q1 UPDATE CASCADE when an order line also references the product (snapshots byte-identical); M2 `fk_transactions_email` RESTRICT blocks user delete without any partial cart/demand cascade (snapshots byte-identical). Identical to T033. |
| RR6 | US3.8 | `identical_enforcement_after_reconstruction` (I1–I4) | T034 table above | `--- PASS (0.64s)`; I1 pre-reconstruction enforcement matrix matches I4 post-reconstruction enforcement matrix byte-identically after the internal schema down/up; I3 confirms `foreign_key_checks=1` on the newly acquired connection. Owned by T034. |

Coverage mapping:

| Requirement | Evidence |
| --- | --- |
| US3.8 — Removing and reinstating the schema yields the same identity/reference rules; identities and references are not permitted to change outside the deliberate down/up operation | I1 pre-reconstruction `assertReferenceEnforcementMatrix` and I4 post-reconstruction `assertReferenceEnforcementMatrix` produce byte-identical orphan/referenced-product-update refusals and byte-identical accepted writes on the recreated schema. I2 confirms both the bounded schema down (`000001_complete_mysql_schema.down.sql`) and the schema up (`000001_complete_mysql_schema.up.sql`) succeed through `golang-migrate` without `migrate.Force`; the sentinel-fixture track remains at version 1/dirty 0 and `sentinel_owned` definition/data are byte-identical before (step 3) and after (step 5) the whole T034 sequence, so no identity/reference drift occurs between the populated lifecycle, the reference matrix and the internal schema reconstruction. I3 confirms `foreign_key_checks=1` on the newly acquired pool connection the reference suite opens after the reconstruction. |
| SC-004 — Historical integrity: no orphan lines, unintended deletions or rewritten identities; valid distinct pairs and ordinary product-field changes remain possible — full US3 facet | RR1 (US3.1/US3.2/US3.9, T031), RR2 (US3.3/US3.9 and the `orders.prodid` length-vs-FK matrix, T031), RR3 (US3.4–7, T032), RR4 (US3.10/US3.11 Q1 cascade/restrict, T033), RR5 (US3.10/US3.11 mixed reference atomicity, T033) and RR6 (US3.8 identical enforcement after reconstruction, T034) collectively cover every US3 facet of SC-004. Each refused write is paired with `assertReferenceRowsEqual` over the six reference tables, each permitted cascade/update is confirmed by `assertCount` or `queryProductValues`, and the sentinel definition/data and `foreign_key_checks=1` on the newly acquired session remain byte-identical before and after the whole T034 sequence (steps 3 and 5 above). The accepted inverses (valid distinct pairs, duplicate demand requests, permitted non-identity metadata updates, Q1 cascades when no restricting relationship applies) are the SC-004 inverse: ordinary business writes remain possible on the recreated schema. No orphan is introduced, no unintended row is deleted, no identity is rewritten by a refused write, and no unrelated `sentinel_owned` row is mutated. |

Spec-contract preservation:

- No shared constraint from T009/T010/T016–T019 was weakened to make the T034 replay pass. `migrations/schema/000001_complete_mysql_schema.up.sql` continues to define `product.pid VARCHAR(30) PRIMARY KEY`, `user.email VARCHAR(100) PRIMARY KEY`, composite `(orderid, prodid)` for `orders`, `uq_usercart_email_pid`, `uq_transactions_transaction_id`, `productid` `ON DELETE NO ACTION ON UPDATE NO ACTION`, `fk_usercart_email` and `fk_user_demand_email` `ON DELETE CASCADE ON UPDATE CASCADE`, `fk_usercart_pid`, `fk_transactions_email` and `fk_transactions_pid` `ON DELETE RESTRICT ON UPDATE CASCADE`, and `orders.prodid VARCHAR(45)`. The sentinel fixture `tests/schema/testdata/sentinel/000001_unrelated_object.up.sql` continues to define `sentinel_owned` with `label VARCHAR(64) COLLATE utf8mb4_0900_bin NOT NULL`, `PRIMARY KEY (id)` and `UNIQUE KEY uq_sentinel_owned_label (label)`, byte-identical to the step 3 and step 5 `SHOW CREATE TABLE` snapshots.
- No hand-applied DDL was issued against the Go test session between steps 2 and 4: the mysql client discard in step 1 runs before the Go tests acquire any connection, and the Go tests re-open the pool with `foreign_key_checks=1` through `openConfiguredDB`. `foreign_key_checks` was never disabled on any session the Go tests use; the step-1 session-only `SET FOREIGN_KEY_CHECKS=0`/`SET FOREIGN_KEY_CHECKS=1` wrapper applies solely to the discard mysql client session and is not shared with the Go pool (confirmed by steps 3 and 5 reporting `@@session.foreign_key_checks = 1`).
- The `fkChildErr` (1452/23000), `fkParentErr` (1451/23000), `duplicateErrSpec` (1062/23000) and `tooLongErr` (1406/22001) error class/SQLSTATE expectations match the shared `expectedSQLError` specifications used in T031/T032/T033 and are unchanged.
- `assertReferenceEnforcementMatrix`, `assertForeignKeyChecksEnabled`, `snapshotReferenceRows` and `assertReferenceRowsEqual` were introduced in T009 and have not been modified; `assertSentinelUnchanged`, `seedSentinel` and `snapshotSentinel` were introduced in T011 and have not been modified. The lifecycle test applies `sentinel_owned` via the sentinel fixture migration version 1 and the schema via the schema migration — no DDL was authored outside those versioned migrations for the T034 replay.

Disposable-instance discard:

```sh
LD_LIBRARY_PATH="$PWD/.tools/libroot/usr/lib/x86_64-linux-gnu:$PWD/.tools/libroot/lib/x86_64-linux-gnu:$PWD/.tools/mysql-8.4.2-linux-glibc2.28-x86_64/lib" ./.tools/mysql-8.4.2-linux-glibc2.28-x86_64/bin/mysql --protocol=tcp -h127.0.0.1 -P3307 -uschema_test -pxxxxx -D 'shopping-cart' -e "SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS orders, usercart, user_demand, transactions, product, \`user\`, sentinel_owned, schema_migrations_schema, schema_migrations_credentials, schema_migrations_credentials_finalize, schema_migrations_fixture_legacy, schema_migrations_fixture_sentinel; DROP PROCEDURE IF EXISTS schema_guard_finalize_credentials; SET FOREIGN_KEY_CHECKS=1; SHOW TABLES;"
```

Exit status: `0`. `SHOW TABLES` returns no rows after the discard, so the schema-only track, the sentinel fixture track and all synthetic product/user/order/cart/demand/transaction rows seeded by the populated lifecycle and the complete reference matrix (including the internal reconstruction fixture) are cleaned up through `DROP TABLE` on the disposable namespace. The session-only `SET FOREIGN_KEY_CHECKS=0`/`SET FOREIGN_KEY_CHECKS=1` wrapper on the mysql client discard session is never shared with the Go test pool, which continued to present `foreign_key_checks=1` on every newly acquired connection throughout steps 2–5. No forced history or hand-applied DDL was used, no `migrate.Force` was invoked, and no push, merge or deployment was performed. Passwords in any DSN above are redacted.
