---
description: "Dependency-ordered implementation tasks for the Complete MySQL Schema feature"
---

# Tasks: Complete MySQL Schema

**Input**: Design documents from `/specs/001-complete-mysql-schema/`  
**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [schema contract](contracts/schema-contract.md), [migration contract](contracts/migration-contract.md), [quickstart.md](quickstart.md).  
**Governance**: [.specify/memory/constitution.md](../../.specify/memory/constitution.md), version 1.0.1.  
**Work unit**: run-e0554f1b7dd4476b-wu  
**Goal**: The operator declares source java, target go, and goal, verbatim, "java to go". These tasks implement only the mysql_database persistence foundation for that rebuild.

**Tests**: Required by spec FR-014, US1–3 and the constitution. Write table-driven Go acceptance tests and retain meaningful expected failures before each corresponding Go configuration/environment check, migration-entry preflight and SQL implementation, then retain passing results. Scaffolding must not implement a refusal early. Configuration/environment checks turn green before domain red gates; lifecycle/upgrade red gates include missing entry preflights as well as SQL behavior. No application endpoints or pages are introduced: the endpoint × admin/editor/viewer matrix has zero rows, and HTTP refusal/audit/error-envelope and Playwright coverage belong to later specifications.

**Organization**: Tasks are grouped by user story after shared prerequisites. All three stories have priority P1. A single complete-schema migration implements shared fields and references; therefore its definition, value and reference acceptance suites are written and demonstrated red in Phase 2, before any complete-schema SQL. US2 and US3 then verify those already implemented contracts rather than adding artificial service layers or weakening US1's complete-schema requirement. The separately implemented finalization behavior gets its own red-before-green cycle in US2.

## Format: `[ID] [P?] [Story] Description`

- Every executable task uses an unchecked checkbox and a sequential T001–T040 identifier.
- `[P]` identifies work in different files that can proceed together once its prerequisites are complete. It does not authorize concurrent mutation of a shared database.
- `[US1]`, `[US2]` and `[US3]` label user-story tasks only. Shared prerequisites and final cross-cutting tasks have no story label.
- Paths below are relative to this working tree unless explicitly identifying read-only legacy evidence.

## Path Conventions and Scope

Use the source layout in plan.md: `internal/config/`, `migrations/schema/`, `migrations/credentials/`, `migrations/credentials-finalize/`, `tests/schema/`, and `tests/schema/testdata/{legacy,sentinel}/`. Root tooling is `go.mod`, `go.sum`, `compose.schema.yaml` and `.golangci.yml`. Add `Makefile` and `.github/workflows/schema.yml` to make required checks fail automatically. Retain sanitized implementation evidence in `specs/001-complete-mysql-schema/validation.md` and the implementation changelog in `CHANGELOG.md`.

The six business tables are product, user, orders, usercart, user_demand and transactions in shopping-cart. Orders contains order lines. Preserve all 36 target columns, Q1/Q2/R1, D1–D5 and the 40 legacy-rule dispositions in data-model.md. No separate order header, users/books parents, legacy username/barcode aliases, authentication implementation, application DB utility, business seeds or general data conversion is included. Q2 alone authorizes the account-preserving credential-column migration. No permitted document read contradicts the declared Java-to-Go language direction.

Legacy documentation is read-only. Enter through `../../old/docs/README.md`; schema evidence is the mysql_database sections of PRD/BRD/LLD and overview, with the scope boundary in roadmap item 001. Recorded operator answers in spec.md override the superseded legacy definitions and stale checklist prose. Do not recover missing estate facts from surrounding code or invent UNVERIFIED behavior. Every new supporting document must state its Java-to-Go purpose and include a changelog entry.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish only the disposable SQL-validation project and reproducible Go tooling.

- [x] T001 Create the planned directories under `internal/config/`, `migrations/`, `tests/schema/testdata/legacy/` and `tests/schema/testdata/sentinel/`, and initialize `go.mod` with the Go 1.24 baseline and an exact compatible toolchain pin; defer external module requirements until importing code is added and do not create placeholder feature migrations before their red gates.
- [x] T002 [P] Create `compose.schema.yaml` for a disposable digest-pinned MySQL 8.4 instance, infrastructure-provisioned shopping-cart namespace, test-only connection, health check and owned volume; require runtime-injected SCHEMA_TEST_MYSQL_PASSWORD, keep credentials out of the file, and configure the planned strict modes, UTC and utf8mb4 environment without applying business DDL.
- [x] T003 [P] Create `.golangci.yml` and `Makefile` with pinned compatible goimports/golangci-lint tool versions and a schema-check target that fails on gofmt/goimports output or failing vet/lint/race commands; use the Go equivalents required by the constitution and add no unused module dependency or Python gate.

**Checkpoint**: Project paths and disposable tooling exist; no business schema has been implemented.

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish non-enforcing scaffolding, demonstrate missing configuration/environment refusals before implementing them, then establish complete-schema acceptance contracts before any story implements business DDL.

- [ ] T004 Create only callable scaffolding in `internal/config/config.go`, `tests/schema/helpers_test.go` and test startup in `tests/schema/environment_test.go`: typed Config loading once from SCHEMA_TEST_DSN and required runtime settings solely through internal/config, read-only connection/inspection helpers and non-enforcing configuration/environment validation entry points. Import go-sql-driver/mysql v1 with its exact version in `go.mod`/`go.sum` alongside its use. Write table-driven TestConfig in `internal/config/config_test.go` and TestEnvironmentPreflight in `tests/schema/environment_test.go` for valid inputs and missing/invalid configuration, wrong namespace and unsupported environment/settings. Keep startup on valid runtime configuration; pass synthetic Config values or isolated connection settings to negative cases without rereading environment or invoking migration mutation. Establish actual container connectivity with a read-only positive control, then run the refusal tests before adding validation and record assertions failing because invalid inputs are accepted, commands and exit status in `specs/001-complete-mysql-schema/validation.md` with goal/changelog. A startup abort, parse/connection/tool error or skip does not establish the missing refusal. Do not implement configuration/environment refusals in this task.
- [ ] T005 After T004's recorded configuration/environment red gate, implement required/invalid configuration validation in `internal/config/config.go` and namespace/environment refusals in `tests/schema/helpers_test.go`/`tests/schema/environment_test.go`, with redacted diagnostics and Config passed to helpers. Complete migration scaffolding with golang-migrate v4 file/MySQL imports pinned in `go.mod`/`go.sum`, retaining the MySQL driver pin from T004; supply explicit per-track and fixture histories, selected fixture-version targets, owned-object cleanup, metadata/row snapshots, SQL code/SQLSTATE assertions and isolated fixtures, and apply strict modes, UTC, utf8mb4 and foreign_key_checks=1 to every pooled and migration connection. Propagate migration errors/dirty state without automatic force or retry. Leave occupied-bootstrap and incompatible-credential-entry checks callable but non-enforcing until T011/T012 tests and T015 red evidence; do not implement finalization preflight before T024.
- [ ] T006 Re-run TestConfig from `internal/config/config_test.go` and TestEnvironmentPreflight from `tests/schema/environment_test.go` after T005 and retain green results alongside T004's red evidence in `specs/001-complete-mysql-schema/validation.md`. Add and run the read-only TestEnvironment baseline against the actual container; assert MySQL 8.4, selected namespace, InnoDB support, utf8mb4/utf8mb4_0900_bin, STRICT_ALL_TABLES, NO_ENGINE_SUBSTITUTION, NO_AUTO_VALUE_ON_ZERO, UTC and enabled FKs across newly acquired SQL and migration connections. Verify invalid inputs are refused before migration invocation with unchanged business objects/history; fail rather than skip on missing prerequisites. All three tests must be green before using any domain test failure as evidence.
- [ ] T007 [P] Write TestSchemaDefinitions in `tests/schema/definitions_test.go` from all 36 columns and six relationships in `specs/001-complete-mysql-schema/data-model.md`; inspect INFORMATION_SCHEMA/SHOW CREATE TABLE for complete inventory, empty business tables, exact sizes/signedness/nullability/defaults, PK order, unique/supporting indexes, named FK actions, enforced quantity checks, engine/charset/collation, initial protected credentials and absence of obsolete/extra business objects. Normalize metadata aliases without changing the required semantics.
- [ ] T008 [P] Write TestSchemaValues in `tests/schema/values_test.go` covering every schema-contract value row: independent omission/NULL/required cases; every VARCHAR exact/one-over limit with utf8mb4 probes; optional binary image round-trip; money fidelity/range/defaults and separate fractional-rounding characterization; signed INT/BIGINT endpoints and one-over bounds; permitted zero/negative stock/line quantities and shipped=2; generated/NULL-generated/explicit-zero/duplicate/boundary ids on separate fixtures; quantity checks; bounded CURRENT_TIMESTAMP defaults and explicit timestamps; author NULL and duplicate demand requests; full synthetic encoded-hash capacity 255/256. Use exact strings for money and overflow probes, assert the correct SQL error category and unchanged stored values, and do not require a successful 45-character order child whose parent can only be 30 characters.
- [ ] T009 [P] Write TestSchemaReferences in `tests/schema/references_test.go` for every schema-contract identity/reference row: duplicate product/user/transaction/generated identities and order/cart pairs, valid distinct pairs, all six FK insert/update orphan refusals, historical product delete/id-update protection with one/several/last references, permitted metadata updates, Q1 cascade/restrict actions, mixed-reference atomicity and identical enforcement after reconstruction. Snapshot all affected parents/children before failures, use separate fixtures for deletion versus update and cascade versus restriction, and distinguish orders.prodid length45 from matching parent length30.
- [ ] T010 Run the shared definition/value/reference suites from `tests/schema/definitions_test.go`, `tests/schema/values_test.go` and `tests/schema/references_test.go` before complete-schema migration files exist, and record commands, assertion failures and exit status in `specs/001-complete-mysql-schema/validation.md` with goal/changelog and coverage by story. Make the harness inspect the empty provisioned baseline when the target source is not yet authored so failures identify absent required objects/behavior; compilation, connection, missing-tool or missing-source-file errors do not count as the red gate. Preserve this evidence when the same tests later turn green.

**Checkpoint**: T001–T010 are complete, configuration/environment checks have T004 red and T006 green evidence, the actual environment works, and the shared US1/US2/US3 schema contracts have meaningful red evidence. Occupied-bootstrap and credential-entry checks remain non-enforcing pending T015. These prerequisites block all story implementation. No mock, SQLite substitute, test skip or missing configuration counts as a pass.

## Phase 3: User Story 1 — Establish a complete store schema (Priority: P1) — MVP

**Goal**: Deliver executable complete-schema creation/removal and the independent account-preserving credential replacement, without seeds, lost profiles or unrelated-object changes.

**Independent Test**: In an empty disposable shopping-cart namespace, apply only the schema track through golang-migrate, inspect complete definitions and zero business rows, populate synthetic references, then down/up and compare definitions and an unrelated sentinel. Independently upgrade a synthetic legacy user fixture through the credentials track, retaining all account/profile data and initializing NULL hash/TRUE reset state.

### Tests and Fixtures for User Story 1

- [ ] T011 [P] [US1] Write TestSchemaLifecycle in `tests/schema/lifecycle_test.go` for initial up, already-applied no-change up, occupied-inventory refusal, populated bounded down/up, identical metadata and re-enabled behavior, unchanged unrelated sentinel, and external-reference teardown refusal without bypass in a separate disposable fixture; apply the sentinel's external-reference version only after product exists, inspect partial DDL state on refusal and discard that fixture. Test the occupied-bootstrap entry check directly on the legacy user fixture before target migrations exist, asserting refusal before migration invocation and unchanged objects/history; keep that check non-enforcing until T015 records the missing refusal. Assert per-track version/dirty state and explicit finalization exclusion; defer finalized-schema combinations to T023 so US1 runs independently. Prepare tests before their preflight implementation or feature/fixture SQL.
- [ ] T012 [P] [US1] Write TestCredentialUpgrade in `tests/schema/credentials_test.go` for the legacy-only entry state, incompatible-entry refusal, profile/identity/count preservation without selecting password contents, target-column initialization, absent password, unchanged unrelated objects and pre-existing profile definitions, repeated clean up, and deliberate down refusal retaining protected data while surfacing actual dirty/version state. Test the credential-entry check directly on an empty fixture before target migrations exist, asserting refusal for missing user before migration invocation and unchanged objects/history; keep that check non-enforcing until T015 records the missing refusal. Include refusal of the initial target schema: demonstrate it red in T020 once that fixture exists, before implementing the preflight, then verify green in T022. Keep upgrade fixtures independent of bootstrap and do not claim arbitrary legacy databases become Q1/R1-conformant.
- [ ] T013 [P] [US1] After T011, add the versioned sentinel fixture pair `tests/schema/testdata/sentinel/000001_unrelated_object.up.sql` and `tests/schema/testdata/sentinel/000001_unrelated_object.down.sql` with its own history and a bounded owned table; keep this first version independent of business tables and put minimal synthetic rows in test DML. Add `tests/schema/testdata/sentinel/000002_external_reference.up.sql` and `tests/schema/testdata/sentinel/000002_external_reference.down.sql` to add/remove a fixture-only FK to product solely for the isolated teardown-refusal case. Normal sentinel setup targets version 1; apply version 2 only after schema up. Apply every fixture table/constraint change through golang-migrate.
- [ ] T014 [P] [US1] After T012, add `tests/schema/testdata/legacy/000001_legacy_user.up.sql` and `tests/schema/testdata/legacy/000001_legacy_user.down.sql` from the verified legacy user columns, including email VARCHAR(60) and nullable password VARCHAR(20); use a separate fixture history and synthetic account/profile DML with NULL passwords only. Never copy a real account or seed payload, and remove only fixture-owned objects.
- [ ] T015 [US1] Apply the legacy fixture and sentinel version 1 through golang-migrate and establish expected lifecycle/credential-upgrade failures from `tests/schema/lifecycle_test.go` and `tests/schema/credentials_test.go` before their target migrations or entry-preflight implementations. Separately run the direct occupied-bootstrap check on the legacy fixture and credential-entry check on an empty fixture, recording assertions failing because the still non-enforcing checks accept these states; do not invoke migration mutation or look up an unauthored source in these refusal reds. Retain the existing lifecycle/upgrade SQL red gate as well; leave the business-dependent external-reference fixture unapplied until product exists. Record each preflight and SQL red result, working fixture/environment checks, unchanged state, commands and exit status in `specs/001-complete-mysql-schema/validation.md`. T016 and T020 are blocked until their respective preflight and SQL reds are retained. Do not count inability to load an unauthored migration as the expected failure.

### Implementation for User Story 1

- [ ] T016 [US1] After T010/T015, implement the occupied-bootstrap inventory preflight in `tests/schema/helpers_test.go`; require T011's occupied-bootstrap preflight refusal cases green with unchanged business objects/history before invoking the schema migration, and retain results beside T015's red evidence. Then create `migrations/schema/000001_complete_mysql_schema.up.sql` with product and user first: product.pid VARCHAR(30), seven product columns, user.email VARCHAR(100), seven account columns including initially nullable password_hash VARCHAR(255) and reset BOOLEAN NOT NULL DEFAULT TRUE, exact data-model defaults/nullability, signed native numerics, required DECIMAL(10,2) DEFAULT 0.00 money and explicit InnoDB/utf8mb4/utf8mb4_0900_bin; use CREATE TABLE without IF NOT EXISTS and include no seed, password column, namespace drop or extra policy.
- [ ] T017 [US1] Extend `migrations/schema/000001_complete_mysql_schema.up.sql` with the five-column orders relation, PRIMARY KEY (orderid, prodid), VARCHAR(45) child identities, nullable signed quantity, required DECIMAL(10,2) amount/default0.00 and integer shipped/default0; add its supporting product index and productid FK to product.pid with explicit NO ACTION on delete/update. Preserve the parent/child length difference and do not add an order header or transaction FK.
- [ ] T018 [US1] Complete `migrations/schema/000001_complete_mysql_schema.up.sql` with the exact usercart, user_demand and transactions definitions from data-model.md: signed generated BIGINT PKs, required R1 email/pid references, supplied lengths/defaults/timestamps, uq_usercart_email_pid and uq_transactions_transaction_id, enforced ck_usercart_quantity/ck_transactions_quantity minimum1, and all six named FK/index contracts including previously added productid. Q1 FKs update CASCADE; cart/demand email deletes CASCADE, other Q1 deletes RESTRICT. Add no demand uniqueness, alias columns or unsupported checks; finish every statement.
- [ ] T019 [US1] Create `migrations/schema/000001_complete_mysql_schema.down.sql` with bounded DROP TABLE IF EXISTS in orders, usercart, user_demand, transactions, product, user order; preserve the namespace, administrative histories and unrelated objects, leave FKs enabled, and let external dependencies fail visibly rather than broadening cleanup.
- [ ] T020 [US1] After T012/T015, first run T012's initial-target credential-entry case on the schema now available from T016–T019 and retain its missing-refusal red assertion without invoking migration mutation. Then implement the legacy-only credential-entry preflight in `tests/schema/helpers_test.go`; require T012's incompatible-entry refusal cases green with unchanged business objects/history before invoking the credential migration, and retain results beside T015's red evidence. Create `migrations/credentials/000001_replace_legacy_password.up.sql` with a single user ALTER dropping password and adding password_hash VARCHAR(255) NULL DEFAULT NULL plus must_reset_password BOOLEAN NOT NULL DEFAULT TRUE; retain every account/profile value and legacy profile definition, initialize existing rows through defaults, and never read, copy, hash or migrate discarded password contents.
- [ ] T021 [US1] Create `migrations/credentials/000001_replace_legacy_password.down.sql` as executable server SQL that deliberately raises a diagnostic before business-schema mutation, refusing return to plaintext storage; retain protected columns/hashes/profiles, introduce no client DELIMITER directive, and surface any administrative routine/history effects without reporting a successful downgrade.
- [ ] T022 [US1] Run TestSchemaDefinitions, TestSchemaLifecycle and TestCredentialUpgrade from `tests/schema/definitions_test.go`, `tests/schema/lifecycle_test.go` and `tests/schema/credentials_test.go` with race detection on isolated fixtures; verify zero initial business rows, all metadata including FKs/CHECKs, schema down/up preservation of the sentinel, clean successful histories/no-change repeated up, occupied-bootstrap and incompatible-credential-entry refusals before migration invocation with unchanged business objects/history (including credential-entry refusal on initial target schema), upgrade preservation and truthful refused-down state. Record US1.1–5, SC-001/002/005 and D1/D2/D3/D4/D5 outcomes in `specs/001-complete-mysql-schema/validation.md`; discard deliberately dirty disposable instances rather than forcing history.

**Checkpoint**: US1 is independently usable and testable. The full shared DDL already includes the fields and references that US2/US3 will verify. This MVP includes Q2 replacement; it does not auto-finalize credentials or claim completion of the remaining stories' validation.

## Phase 4: User Story 2 — Preserve field values and documented defaults (Priority: P1)

**Goal**: Prove exact field compatibility and deliver the separately selected, guarded transition to required protected hashes after every account has reset.

**Independent Test**: Apply the schema to disposable fixtures, run every omission/NULL/length/numeric/default case without application services, then exercise credential finalization separately against pending and ready account states. Check stored values and expected SQL errors, not only metadata.

### Tests for User Story 2

- [ ] T023 [US2] Extend `tests/schema/credentials_test.go` with TestCredentialFinalization before its implementation: test NULL/TRUE, NULL/FALSE, non-null/TRUE and other nonzero reset flags independently; test empty/all-ready states, both harness preflight and direct migration SQL guard, writers-quiesced precondition, readiness preserving all hashes/flags/profiles, subsequent hash NULL/omission refusal, down restoring only nullability, no plaintext restoration, helper-routine cleanup and truthful failed-step administrative state. Extend `tests/schema/lifecycle_test.go` now with finalization-down before bootstrap reconstruction, no implicit finalization and no stale finalized history over recreated nullable hashes. Existing value tests from T008 cover initial hash capacity and reset defaults.
- [ ] T024 [US2] Run TestCredentialFinalization from `tests/schema/credentials_test.go` on working initial/upgraded fixtures before the finalization SQL/helper behavior exists; capture an expected target assertion such as the ready schema still permitting NULL hashes in `specs/001-complete-mysql-schema/validation.md`. Retain prior red evidence for the shared value contract from T010; do not manufacture a new red result after those values are implemented.

### Implementation and Verification for User Story 2

- [ ] T025 [P] [US2] After T024, create `migrations/credentials-finalize/000001_require_password_hash.up.sql` with a migration-owned schema_guard_finalize_credentials routine that raises SQLSTATE 45000 while any hash is NULL or reset flag is truthy, otherwise changes only password_hash to VARCHAR(255) NOT NULL with no default; remove the routine after success, keep every business value unchanged, and use server SQL without DELIMITER directives or a permanent reset CHECK.
- [ ] T026 [P] [US2] After T024, create `migrations/credentials-finalize/000001_require_password_hash.down.sql` to restore only password_hash VARCHAR(255) NULL DEFAULT NULL; preserve hashes/reset flags/profiles and never recreate password or alter unrelated definitions.
- [ ] T027 [P] [US2] After T024, extend `tests/schema/helpers_test.go` with explicit finalization source/history selection, preflight rejecting NULL hashes or any nonzero reset flag, and a documented quiesced-writer test boundary covering preflight/SQL guard/ALTER. Provide a deliberate direct-through-migrate guard test path, inspect failed history/routine ownership, and discard that isolated fixture on refusal; never assume the migration lock or a DDL rollback protects account writers.
- [ ] T028 [US2] Execute the complete TestSchemaValues suite from `tests/schema/values_test.go` against actual initial target schema, checking all US2.1–11 boundaries/defaults and SQL error categories; verify finalized hash NULL/omission behavior with the credential cases and preserve explicit initial-state versus finalized-state expectations. Use independent max-id/overflow fixtures, exact decimal comparisons and before/after row assertions; record per-case results in `specs/001-complete-mysql-schema/validation.md` without treating credential-only legacy email lengths as full target conformance.
- [ ] T029 [US2] Execute TestCredentialFinalization from `tests/schema/credentials_test.go` and the finalization-aware lifecycle cases from `tests/schema/lifecycle_test.go` after T025–T027; prove pending-state preflight and SQL refusals, all-ready/empty success, preserved values and absent guard on success, safe down, separate histories, explicit finalization-down before schema reconstruction and no stale applied finalization over recreated nullable hashes. Record US2.12, Q2 and SC-003/005 in `specs/001-complete-mysql-schema/validation.md` and report dirty failures as failures/refusals, never successful reversions.
- [ ] T030 [US2] Execute the environment-characterization cases prepared in `tests/schema/environment_test.go` and `tests/schema/values_test.go`, recording actual target case/accent/trailing-space behavior and exact 12.345→12.35 / -12.345→-12.35 decimal conversion with warnings/notes in `specs/001-complete-mysql-schema/validation.md`. Keep these target-environment choices distinct from UNVERIFIED legacy behavior, and keep unexpected results visibly failing until the environment contract is reconciled.

**Checkpoint**: Every specified field/default/range case has an executed result, and credential finalization is available only through its explicitly selected track after readiness and writer quiescence. No sign-in, password hashing or reset workflow has been implemented.

## Phase 5: User Story 3 — Protect identities and historical order references (Priority: P1)

**Goal**: Prove uniqueness, orphan prevention, historical product protection and the operator's corrected cascade/restrict actions.

**Independent Test**: On a freshly established schema, use only synthetic products/accounts/children to exercise duplicates, orphan inserts/updates, referenced-parent changes, last-reference removal and successful Q1 cascades. Compare affected rows after every rejected write. These tests do not require credential finalization, checkout or a product service.

### Verification of the Test-First Shared Implementation

The reference tests were written in T009 and demonstrated red in T010; their SQL constraints were implemented once in T016–T018. This phase completes the independently testable historical-integrity journey without duplicating or deferring those constraints to a second schema artifact.

- [ ] T031 [US3] Run the identity/pair/orphan groups in `tests/schema/references_test.go` against the schema-only track: reject identical product/user/transaction/generated identities, duplicate order/cart pairs and all FK insert/update orphans with correct SQL error codes/states and unchanged rows, while accepting valid distinct pairs and duplicate demand requests. Verify a matching product id up to 30 characters succeeds, a nonexistent orders.prodid of 31–45 characters fails FK and a value over 45 characters fails length; record US3.1–3/9 in `specs/001-complete-mysql-schema/validation.md`.
- [ ] T032 [US3] Run historical-product cases in `tests/schema/references_test.go`: reject delete/id update with one/several references, continue rejecting after all but the last reference are removed, then allow each operation in separate fixtures after the last reference and any other restricting reference is removed; permit valid non-identity updates throughout. Compare all affected rows after refusal and record US3.4–7 and SC-004 in `specs/001-complete-mysql-schema/validation.md`.
- [ ] T033 [US3] Run Q1-action and mixed-reference groups in `tests/schema/references_test.go`: separately prove user-to-cart/demand delete cascades, transaction-to-user and cart/transaction-to-product delete restrictions, account/product id-update cascades, order-line NO ACTION overriding attempted mixed product rewrites, and a transaction-restricted user delete leaving cart/demand untouched. Record US3.9–11 and SC-004 with complete before/after snapshots in `specs/001-complete-mysql-schema/validation.md`.
- [ ] T034 [US3] Re-run the complete reference case matrix from `tests/schema/references_test.go` after the populated schema lifecycle in `tests/schema/lifecycle_test.go`; assert identical results, foreign_key_checks=1 on newly acquired connections and unchanged sentinel definition/data, with no hand-applied DDL or disabled enforcement. Record US3.8 and the full US3/SC-004 outcome in `specs/001-complete-mysql-schema/validation.md`.

**Checkpoint**: Every defined identity/reference case has an executed outcome, without orphans, unintended deletions or rewritten protected identities. A local pass remains evidence for review, not platform acceptance.

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Make the implementation reproducible, traceable and reviewable across all stories.

- [ ] T035 [P] Update `specs/001-complete-mysql-schema/quickstart.md` to match the implemented commands, pinned tools/image, runtime-only configuration, three explicit histories/entry states, irreversible credential-down refusal, finalization quiescence, expected partial-DDL/dirty-state handling and disposable cleanup; preserve Java-to-Go purpose/changelog, remove future-only claims only where implemented, and keep absent/unexecuted gates explicit.
- [ ] T036 [P] Add `.github/workflows/schema.yml` to provision the disposable MySQL environment, inject ephemeral configuration without committed credentials, run the non-skipping Makefile schema-check gate with pinned tools and fail on formatting/vet/lint/race or acceptance failures, retaining sanitized diagnostics and cleaning only its owned instance. Add no push, merge or deployment step.
- [ ] T037 Complete the environment and administrative-object inventory in `specs/001-complete-mysql-schema/validation.md`: record actual Go/tool/module/server versions and image digest, all three feature histories plus fixture histories, selected modes/collation, successful versions/dirty=false and exact failed-step dirty/routine state; retain red-before-green commands/results without DSNs, password material, hashes or real data.
- [ ] T038 Run the complete workflow in `specs/001-complete-mysql-schema/quickstart.md` and Makefile schema-check, including gofmt/goimports with zero reported files, go vet, golangci-lint and go test -race ./... -count=1, on the pinned disposable MySQL instance; retain sanitized per-case results and cleanup evidence in `specs/001-complete-mysql-schema/validation.md`. Fix conformance failures in their owning files, rerun affected checks, and report missing/unexecuted gates without claiming passes.
- [ ] T039 Review `migrations/`, `internal/config/`, `tests/schema/` and `specs/001-complete-mysql-schema/validation.md` against every FR-001–016, US1–3 scenario, SC-001–005, D1–D5 and all 40 legacy-rule dispositions; add an explicit task/case/source/status traceability matrix, confirm all modules remain under 1,000 lines, dependencies have imports, scope/secret/configuration rules hold and up/down pairs are tracked in the implementation Git checkout. If no Git repository is available, record the migration-in-Git gate as unmet rather than asserting tracking; local test results never decide referee/platform acceptance.
- [ ] T040 Add the implementation change entry to `CHANGELOG.md` describing the Java-to-Go schema foundation, all three migration tracks, Q1/Q2/R1 differences, destructive synthetic reconstruction and credential downgrade limitation, and link the actual validation evidence in `specs/001-complete-mysql-schema/validation.md`; report remaining failures/unexecuted gates without pushing, merging, deploying or declaring the change accepted.

## Dependencies & Execution Order

### Phase Dependencies

| Phase | Depends on | Completion result |
| --- | --- | --- |
| Setup, T001–T003 | Existing resolved design artifacts | Paths, disposable DB/tool definitions; no business DDL. |
| Foundational, T004–T010 | Setup; T004 scaffolding/tests/red before T005 configuration/environment refusals; T006 green before shared case authoring/execution | Working environment with recorded configuration/environment red/green evidence and all shared complete-schema acceptance suites demonstrated red. |
| US1, T011–T022 | Foundational; T011/T012 tests and T015 preflight/SQL reds before T016 inventory and T020 credential-entry refusals or corresponding SQL | Complete shared schema/lifecycle and independent credential replacement, including tested entry preflights. |
| US2, T023–T030 | US1 for ready target/upgrade fixtures; T023/T024 before T025–T027 | Value-fidelity evidence and guarded finalization. |
| US3, T031–T034 | US1 shared schema and Phase2 tests; independent of US2 finalization | Identity/reference evidence, including lifecycle replay. |
| Polish, T035–T040 | All stories; T035/T036 can proceed together; T037 before final gate run and traceability review | Reproducible tooling, results, documented limits and changelog. |

### Story Completion Graph

```text
Resolved spec + plan + contracts
              |
        Setup T001–T003
              |
    Foundation T004–T010
    (shared US1/US2/US3 tests red)
              |
         US1 T011–T022
         /           \
 US2 T023–T030     US3 T031–T034
         \           /
          Polish T035–T040
```

The default execution follows increasing task IDs and story order. T004 provides callable scaffolding and configuration/environment tests/red evidence; T005 may implement those refusals only afterward, and T006 must prove them green before domain red gates. US3 may start after T022 when using a different disposable instance from US2. Within US1, T011 and T012 can be authored together; T013 waits for T011 and T014 waits for T012; T015 waits for both fixtures and records direct entry-preflight reds as well as SQL reds. T016 owns occupied-bootstrap validation and T020 owns credential-entry validation; neither may move into T005. T016–T018 edit the same up file and must remain serial. T020/T021 follow their written tests/red gate. T025–T027 can be authored together after T024, but T028/T029 execute only after all required implementations are ready. T034 repeats the complete reference suite after recreation, not merely a metadata comparison.

### Shared Contract-to-Story Mapping

| Shared acceptance work | Owning journeys | Implementation and completion evidence |
| --- | --- | --- |
| Configuration/environment, T004 tests/red | FR-014/015 and all journeys' validation prerequisites | Go refusals T005; green T006 before domain reds. |
| Definitions, T007; red, T010 | US1 complete setup; supports US2/US3 field and FK expectations | SQL T016–T019; US1 T022. |
| Values, T008; red, T010 | US2.1–11; also US1 empty/initial state and US3 identity boundaries | Shared SQL T016–T018; US2 T028/T030. |
| References, T009; red, T010 | US3.1–11; also US1 dependency-safe setup/recreation | Shared SQL T017/T018; US3 T031–T034. |
| Lifecycle/upgrade entry checks and SQL, T011/T012; fixtures T013/T014; red T015 | US1.1–5, US3.8, Q2 initial migration | Go preflights T016/T020 with refusal-green evidence; SQL T016–T021; T022, then T029/T034 for later lifecycle combinations. |
| Finalization, T023; red T024 | US2.12 | SQL/helper T025–T027; T029. |

All acceptance tests and their recorded failures precede the corresponding Go refusal or SQL implementation. Test a missing preflight through its callable entry point without migration mutation; a later DDL rejection is not evidence of an entry-preflight refusal. All shared schema acceptance tests precede the migration that implements their behavior. Do not wait until US2 or US3 to invent an expected failure for a constraint already delivered in US1. Do not interpret a broken harness, no available server, dirty unrelated fixture or migration-file lookup error as a demonstrated missing behavior.

### Parallel Opportunities

- After T001: T002 and T003 edit independent infrastructure/tool files.
- After T006: T007, T008 and T009 author independent case files with completed helpers.
- US1: T011/T012 are independent case files; after those tests are written, T013/T014 are independent fixture tracks.
- US2: after T024, T025/T026/T027 edit different files with the same completed contract/test prerequisite.
- US3: the four verification tasks share references/lifecycle fixtures and a result document; execute them serially within one instance. The entire US3 journey can run alongside US2 on a separately provisioned disposable MySQL instance, retaining the required shopping-cart namespace in each.
- Final documentation and CI: T035/T036 edit different files. Serialize writes to validation.md.
- Do not use t.Parallel or concurrent migration commands against the same shopping-cart instance. Race detection does not provide database isolation.

## Parallel Example: User Story 1

After T010, author these independent tests together:

```text
T011: TestSchemaLifecycle in tests/schema/lifecycle_test.go
T012: TestCredentialUpgrade in tests/schema/credentials_test.go
```

After both test tasks are complete, author T013's sentinel SQL pair and T014's legacy SQL pair together. Apply fixtures and establish the direct entry-preflight and SQL red gates serially before implementing either the Go preflights or feature SQL. Do not parallelize edits to migrations/schema/000001_complete_mysql_schema.up.sql or the shared preflight helpers.

## Parallel Example: User Story 2

After T024, implement the independently specified files together:

```text
T025: migrations/credentials-finalize/000001_require_password_hash.up.sql
T026: migrations/credentials-finalize/000001_require_password_hash.down.sql
T027: finalization preflight/track support in tests/schema/helpers_test.go
```

Run the combined finalization suite only after all three are complete. A shared credentials_test.go remains a single-writer file.

## Parallel Example: User Story 3

After T022, a cross-story execution can use two separate disposable instances:

```text
Instance A: US3 T031→T032→T033→T034, tests/schema/references_test.go and lifecycle_test.go
Instance B: US2 T028/T029, tests/schema/values_test.go and credentials_test.go
```

Each process receives its own runtime Config/DSN targeting shopping-cart on its own instance. Keep diagnostics separate and merge sanitized evidence into validation.md serially. There is no safe intra-story parallel write/test opportunity for US3 on one instance; its row comparisons and lifecycle runs must remain deterministic.

## Implementation Strategy

### MVP First (User Story 1)

1. Complete setup and the shared acceptance-first foundation: configuration/environment tests and red evidence, their implementation/green evidence, then red evidence for all common DDL behavior.
2. Complete US1's lifecycle/upgrade tests and fixture migrations, demonstrate both entry-preflight and SQL failures, then implement its preflights and complete-schema/credential-replacement pairs.
3. Validate US1 independently against actual MySQL and retain results. This is a reviewable persistence MVP, not a platform acceptance claim or deployment instruction.
4. Continue through US2, US3 and final gates to complete the full feature; the MVP checkpoint does not waive them.

### Incremental Delivery

1. Deliver the shared empty/recreatable schema plus protected credential replacement through US1.
2. Verify all value/null/default boundaries and implement opt-in finalization through US2's separate red-before-green cycle.
3. Independently prove historical integrity and all corrected FK actions through US3; it can run before US2 completes on isolated infrastructure.
4. Finish the complete gated workflow, evidence matrix, tracked migration review and changelog. Never turn unexecuted cases into passes or alter a specified field/action to make a compatibility test green.

### Failure Boundaries

Use ordinary DML for synthetic rows; every feature or fixture CREATE/ALTER/DROP goes through a versioned golang-migrate source. MySQL multi-statement DDL is not assumed to roll back. On an error, stop, inspect the exact source/version, dirty history and owned objects, retain sanitized evidence, and discard only the explicitly disposable instance or use a reviewed versioned forward repair. Do not hand-apply SQL, force migration history, disable FKs, restore plaintext storage, rebuild existing accounts through bootstrap, or remove an unrelated object. Finalization requires quiesced writers and does not run with normal schema up, even on an empty account table.

## Task and Coverage Accounting

| Area | Tasks | Count |
| --- | --- | --- |
| Setup | T001–T003 | 3 |
| Shared foundation, including all complete-schema acceptance contracts | T004–T010 | 7 |
| US1 — complete schema, lifecycle, credential replacement | T011–T022 | 12 |
| US2 — value fidelity, guarded finalization | T023–T030 | 8 |
| US3 — identity/reference integrity | T031–T034 | 4 |
| Polish and cross-cutting concerns | T035–T040 | 6 |
| Total | T001–T040 | 40 |

There are 14 `[P]` tasks across six independent authoring groups; cross-story runtime concurrency additionally requires separate disposable instances. The task list is not a completion claim: every task starts unchecked. At generation, database/build/lint/race, Git tracking and platform acceptance gates remain unexecuted. This workspace has no Git repository; no actual branch or tracked migration is asserted.

## Changelog

- 2026-11-10 — Resolved C1 for the java to go schema foundation: T004 records configuration/environment refusal reds before T005 implementation and T006 greens; T011/T012/T015 precede the occupied-bootstrap and credential-entry preflights now owned by T016/T020. Updated dependencies, coverage mapping and execution guidance while preserving all 40 task IDs and SQL/finalization red gates. No implementation or runtime checks performed.

- 2026-11-10 — Generated 40 dependency-ordered tasks for work unit run-e0554f1b7dd4476b-wu, serving the java to go mysql_database foundation. Consumed the resolved Q1/Q2/R1 specification and three-track design, placed shared contract tests before common DDL, mapped all three P1 journeys and final validation gates, and preserved scope, evidence and acceptance limits. No implementation, database execution, push, merge or deployment performed.
