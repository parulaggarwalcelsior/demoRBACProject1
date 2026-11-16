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
