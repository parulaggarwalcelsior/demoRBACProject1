# Quickstart: Validate the Complete MySQL Schema

**Date**: 2026-11-27 | **Goal**: java to go | **Feature**: [spec.md](spec.md)

This guide runs the implemented Java-to-Go persistence foundation for `mysql_database`.
It validates the six-table MySQL schema, the account-preserving credential replacement
track and the guarded credential-finalization track through the Go acceptance harness.
It does not implement Java application code, HTTP endpoints, roles, pages, sign-in,
password hashing, password reset workflows, checkout, mail or seed/data conversion.

## Prerequisites and runtime configuration

- Docker Compose and a Go 1.24-compatible toolchain.
- Runtime-only configuration:
  - `SCHEMA_TEST_MYSQL_PASSWORD` provisions the disposable MySQL container.
  - `SCHEMA_TEST_DSN` is read once by `internal/config` for the Go harness.
- No credential literal belongs in source, fixtures, documents or logs. Redact DSNs,
  passwords and hashes in retained evidence.
- The DSN must select the disposable `shopping-cart` namespace on the local validation
  instance, for example with `parseTime=true`, `charset=utf8mb4`,
  `collation=utf8mb4_0900_bin` and `loc=UTC`.
- `compose.schema.yaml` pins `mysql:8.4` by digest:
  `sha256:a81d043930e51745522bd5c430961cf8286f0a23f0a9a2a6e0108f7ea71b0902`.
- `Makefile` installs pinned local tools into `.tools/bin`: `goimports v0.31.0` and
  `golangci-lint v1.64.8`.
- Go modules are pinned in `go.mod`: `github.com/go-sql-driver/mysql v1.10.1` and
  `github.com/golang-migrate/migrate/v4 v4.18.3`.

Only `internal/config` reads environment. Test helpers pass typed `Config` to every SQL
and migration connection and apply `STRICT_ALL_TABLES`, `NO_ENGINE_SUBSTITUTION`,
`NO_AUTO_VALUE_ON_ZERO`, UTC, `utf8mb4`, `utf8mb4_0900_bin` and `foreign_key_checks=1`.

## Start the disposable database

From the working tree root, export the runtime values and start the pinned container:

```sh
export SCHEMA_TEST_MYSQL_PASSWORD='xxxxx'
export SCHEMA_TEST_DSN='schema_test:xxxxx@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC'
docker compose -f compose.schema.yaml up -d --wait
go test ./internal/config ./tests/schema -run '^(TestConfig|TestEnvironmentPreflight|TestEnvironment|TestEnvironmentRefusesInvalidBeforeMigration)$' -count=1 -v
```

Expected: configuration and environment tests fail rather than skip when configuration
or server preconditions are missing. When configured correctly they verify MySQL 8.4.x,
the `shopping-cart` namespace, InnoDB, strict modes, UTC, `utf8mb4_0900_bin`, enabled
FKs and refusal before migration mutation for invalid runtime state.

## Migration tracks and entry states

The implementation uses three explicit feature histories. Select them deliberately; no
command applies all three tracks together.

| Track | Source | History table | Entry state |
| --- | --- | --- | --- |
| Complete schema | `migrations/schema` | `schema_migrations_schema` | Existing disposable `shopping-cart` namespace with none of `product`, `user`, `orders`, `usercart`, `user_demand` or `transactions`. |
| Existing credentials | `migrations/credentials` | `schema_migrations_credentials` | Known legacy `user` table with `password`, and without `password_hash` or `must_reset_password`. |
| Credential finalization | `migrations/credentials-finalize` | `schema_migrations_credentials_finalize` | Initial target or upgraded account table where every account has a non-null hash, no truthy reset flag, and account writers are quiesced. |

Fixture migrations use separate histories:
`schema_migrations_fixture_legacy` for `tests/schema/testdata/legacy` and
`schema_migrations_fixture_sentinel` for `tests/schema/testdata/sentinel`.

## Run implemented validation suites

Run the core schema, lifecycle, value, reference and credential suites:

```sh
go test -race ./tests/schema -run '^TestSchema(Definitions|Values|References|Lifecycle)$' -count=1 -v

go test -race ./tests/schema -run '^TestCredential(Upgrade|Finalization)$' -count=1 -v
```

The schema tests apply the complete-schema track through golang-migrate, inspect metadata,
exercise values/defaults/references, run bounded down/up reconstruction and preserve the
sentinel fixture. The credential upgrade test uses only synthetic legacy accounts with
NULL legacy passwords. It preserves profile data, drops `password`, adds nullable
`password_hash` and `must_reset_password`, and does not claim arbitrary legacy databases
become Q1/R1 conformant.

Credential replacement down is intentionally irreversible: the down migration refuses
before mutation and must not recreate plaintext password storage or discard protected
fields. Treat the surfaced dirty/version state as a refusal requiring inspection or a
reviewed forward repair, not as a successful downgrade.

Finalization is opt-in. Quiesce account writers for the full preflight, SQL guard and
`ALTER` interval. The Go preflight rejects NULL hashes or any nonzero reset flag before
invoking migrate, and the SQL migration repeats the guard with
`schema_guard_finalize_credentials`. A direct-through-migrate pending-state refusal may
leave finalization history dirty and/or owned guard residue; inspect that state and
discard the isolated fixture rather than forcing history.

## Complete local gates

The implemented Makefile gate is:

```sh
make tools
make schema-check
```

`schema-check` runs `gofmt`/`goimports` checks over `internal` and `tests`,
`go vet ./...`, `golangci-lint run` and `go test -race ./... -count=1`. Formatting
commands must produce no listed files. Local green results are evidence for review only;
they are not platform acceptance.

CI, the full final evidence inventory, complete traceability review and release
changelog remain separate tasks. If those gates have not been run in the current session,
report them as absent or unexecuted rather than passing.

## Partial DDL, dirty state and cleanup

MySQL DDL is not assumed to roll back as one transaction. On any migration error:

1. Stop immediately.
2. Inspect the exact source, target version, history table, dirty flag, owned objects and
   any `schema_guard_finalize_credentials` routine.
3. Retain sanitized diagnostics in `specs/001-complete-mysql-schema/validation.md`.
4. Do not use blind `migrate.Force`, automatic retry, global FK disabling, hand-applied
   feature DDL, plaintext credential restoration or unrelated-object cleanup.
5. Continue only through a reviewed forward repair or by discarding the disposable
   validation instance/fixture.

After retaining diagnostics, remove the disposable stack and its owned volume:

```sh
docker compose -f compose.schema.yaml down --volumes
```

This is validation cleanup only. It is not a production rollback, legacy migration, push,
merge or deployment.

## Changelog

- 2026-11-27 — Updated the Java-to-Go schema quickstart to match the implemented
  commands, pinned image/tools/modules, runtime-only configuration, three migration
  histories and entry states, irreversible credential-down refusal, finalization
  quiescence, expected dirty/partial-DDL handling, disposable cleanup, and explicit
  remaining gates for T035.

- 2026-11-10 — Added C1's development-stage Go refusal red/green workflow for the java to
  go harness, including delayed bootstrap/credential-entry preflights; preserved SQL and
  finalization gates. Commands and results remained future implementation work.

- 2026-11-10 — Added the future runnable acceptance-first guide for the java to go schema
  foundation, work unit run-89ffa3642d6a4d44-wu. All database/build/lint results remained
  unexecuted during planning.
