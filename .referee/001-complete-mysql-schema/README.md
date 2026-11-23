# Complete MySQL schema acceptance tests

These independent Go integration tests judge the Java-to-Go persistence work order
in `specs/001-complete-mysql-schema/{spec,plan,tasks}.md`. They inspect database
metadata and execute the three delivered migration tracks through golang-migrate.
They do not import implementation packages or inspect implementation source.

Run the folder explicitly; recursive repository discovery skips `.referee`:

```sh
REFEREE_TARGET_ROOT=/absolute/path/to/implementation \
  go -C .referee/001-complete-mysql-schema test -mod=mod -race -count=1 -timeout=45m .
```

Go 1.24 or newer, Git, and Docker with a running daemon are prerequisites. Tests
start owned disposable MySQL 8.4 containers automatically. `REFEREE_MYSQL_IMAGE`
can select an available MySQL 8.4 image, including a digest-pinned image. The
implementation gate additionally needs the tools required by its `schema-check`
target. Missing prerequisites cause failures, never skips.

For an already provisioned **dedicated disposable MySQL 8.4 server**, supply
`REFEREE_MYSQL_ADMIN_DSN` and `REFEREE_DISPOSABLE_MYSQL=1` instead of Docker.
This explicitly authorizes discarding and recreating that server's `shopping-cart`
namespace between fixtures. The administrative account needs database creation,
DDL, routines, and DML privileges. Never point this option at a shared database.
The DSN is runtime configuration and must not be committed. Fixtures serialize
with a server advisory lock; do not run this folder concurrently with other
writers. Each failing/dirty migration fixture is discarded rather than forced.

Coverage:

| Tests | Work-order requirements |
| --- | --- |
| `TestDeliveryArtifacts`, `TestSchemaDefinitions` | US1.1/3, FR-001–010/013/016, all 36 fields, six FKs, exact inventory and initial credentials |
| `TestSchemaValues`, `TestGeneratedSignedBIGINTIdentities` | US2.1–11, FR-004–006/008–012, defaults, NULL, Unicode lengths, integer/decimal limits, images, timestamps and generated ids |
| `TestSchemaLifecycleAndReferenceReplay` | US1.2/4, all US3 cases before and after reconstruction, SC-001/004, row preservation on rejected writes |
| `TestTeardownRefusesExternalDependency` | Bounded removal, visible dirty/partial DDL failure, unrelated objects retained |
| `TestCredentialReplacementPreservesAccountsAndRefusesDown` | US1.5, Q2, credential-only legacy entry, profile preservation, initial reset state, irreversible plaintext downgrade refusal |
| `TestFinalizationSQLGuardRejectsEveryPendingState` | US2.12, direct SQL guard, NULL and every nonzero reset state, unchanged business state on failure |
| `TestCredentialFinalizationAndSafeDown`, `TestFinalizedSchemaRecreationUsesSeparateHistory` | US2.12, empty/ready/legacy success, writer quiescence, safe nullability-only down, separate histories, no stale finalization after reconstruction |
| `TestImplementationRequiredGatesAndRefusals` | T004–006/011–012/022/027/038, named implementation configuration/environment/entry-refusal suites, non-skipping execution, delivered format/vet/lint/race gate |
| `TestMigrationPairsAreTracked`, `TestReviewableValidationEvidence` | FR-014/015, SC-002/005, T035–040 handoff/traceability and tracked pairs |

The referee's direct SQL guard cases intentionally bypass Go preflights. The
implementation's test-only preflight API is unspecified, so its named public Go
test entry points are run separately to cover those harness checks. Historical
red-before-green ordering and truth of recorded evidence require review; a final
filesystem snapshot cannot prove chronology. The evidence test checks the
required accounting exists and does not treat prose as an executed SQL result.

All fixture table/constraint DDL is versioned test-only SQL applied with
golang-migrate's in-memory file source. Namespace creation and disposal belong to
the test infrastructure. No production data, original seeds, plaintext password
values, disabled foreign keys, automatic migration-history forcing, or schema
repair are used. Expected SQL errors are checked by category and SQLSTATE;
rejected DML is checked against snapshots of every business table. Successful
credential finalization runs with all account writers quiesced by fixture ownership.

Changelog: 2026-11-16 — Added blind referee acceptance tests for the complete
MySQL schema work order. Runtime prerequisite failures remain failures and are
not evidence that an implementation's schema behavior has been exercised.
