# Implementation Plan: Complete MySQL Schema

**Branch**: `001-complete-mysql-schema` (setup script's feature identifier; no Git repository or actual branch exists in this workspace) | **Date**: 2026-11-10 | **Spec**: [spec.md](spec.md)

**Input**: `specs/001-complete-mysql-schema/spec.md`  
**Work unit**: run-89ffa3642d6a4d44-wu  
**Governing constitution**: [.specify/memory/constitution.md](../../.specify/memory/constitution.md), version 1.0.1.

## Summary

Provide the reproducible MySQL persistence foundation for the operator's declared source java, target go, and goal "java to go". Reconstruct exactly six business tables in `shopping-cart`: product, user, usercart, orders, transactions and user_demand. `orders` contains order lines; there is no separate header table. Preserve the verified field register and apply the operator's Q1/Q2/R1 replacements, including DECIMAL(10,2) money, corrected parent identities, explicit referential actions and protected credentials.

Deliver the executable replacement for the incomplete mysql_database component as versioned SQL migration pairs consumed by golang-migrate. Use three explicitly selected tracks: complete schema creation/removal, account-preserving credential conversion, and guarded credential finalization. Write Go acceptance tests first, demonstrate their expected failures, then implement each configuration/environment check, migration-entry preflight and SQL migration and prove the constraints against actual MySQL. This invocation produces design documents only; tasks and implementation follow later.

## Technical Context

**Language/Version**: MySQL 8.4 SQL for DDL; Go 1.24 language baseline for a small validation harness. Java is the documented source estate; no Java application is ported here. Toolchain patch versions are locked during implementation and recorded with results.

**Primary Dependencies**: golang-migrate v4 with MySQL and file drivers, go-sql-driver/mysql v1, Go standard testing/database/sql packages, and a disposable MySQL 8.4 container. Add module dependencies only alongside code importing them. Pin module versions in go.mod/go.sum and container digests when creating the harness; no dependency is installed by this planning step.

**Storage**: Existing isolated namespace `shopping-cart`; all six tables InnoDB, utf8mb4 with explicit utf8mb4_0900_bin collation. Per-track golang-migrate metadata tables are administrative objects, separately inventoried. Infrastructure supplies the database namespace; migrations never drop the database.

**Testing**: Table-driven Go acceptance/integration tests against MySQL 8.4; metadata inspection plus successful/rejected SQL writes and before/after row assertions. Require gofmt/goimports, go vet, golangci-lint and go test -race. Record red-before-green evidence separately for configuration/environment refusals, occupied-bootstrap and incompatible-credential-entry preflights, and corresponding SQL behavior under spec FR-014. Scaffolding supplies callable checks and read-only connectivity without implementing those refusals early. No mocks or SQLite substitute for database constraints.

**Target Platform**: Disposable Linux MySQL container and Go test runner. Record actual server/tool versions, image digest, collation and connection settings. No production platform or existing legacy server version is inferred.

**Project Type**: SQL schema/migration deliverable with Go validation tooling, scoped to mysql_database.

**Performance Goals**: No evidenced throughput, latency or production volume target. Correctness, complete constraint coverage and repeatable setup/removal are the gates.

**Constraints**: No business seeds, general data migration, plaintext credentials, unrelated-object removal, disabled foreign keys, added business rules or application functionality. Q2 alone authorizes an account-preserving credential-column migration. Every validation connection must use STRICT_ALL_TABLES, NO_ENGINE_SUBSTITUTION and NO_AUTO_VALUE_ON_ZERO, utf8mb4, and foreign_key_checks=1. DDL is not assumed to roll back as a multi-statement transaction.

**Scale/Scope**: Six business relations, 40 legacy schema-rule dispositions, FR-001–016, three user stories and SC-001–005. Only Q1 authorizes generated BIGINT ids and positive cart/transaction quantities. No endpoint, page, role/session/audit schema, checkout transaction or mail infrastructure is added.

## Constitution Check

Pre-research evaluation passes for the authorized planning scope. The same gates were re-evaluated against the Phase 1 design below. PASS here means design compliance, never executed tests or platform acceptance.

| Gate | Before Phase 0 | After Phase 1 | Evidence / design consequence |
| --- | --- | --- | --- |
| I: Java-to-Go direction and source traceability | PASS | PASS | Start at legacy README; source requirements through spec E1–E8. Q1/Q2/R1 are recorded operator decisions. Environment choices are labelled choices. |
| I: Clarifications and deliberate differences | PASS | PASS | Spec O1/O2/R1 are resolved. Preserve D1–D5 and all 40 dispositions; see research and data model. No conflicting language destination found. |
| II: Layering, size, dependencies, Go gates | PASS | PASS | SQL owns persistence constraints. Go code is test/migration tooling; no business DAOs, services or HTTP adapters. Files stay below 1,000 lines; dependencies require imports. |
| III: Protected credentials and access boundaries | PASS | PASS | Q2 replaces password with encoded-hash capacity and reset state. Credential down refuses to restore plaintext. Application authentication, role checks and membership remain later features. |
| IV: Acceptance before implementation | PASS | PASS | Tests and meaningful red evidence precede Go configuration/environment checks, each migration-entry preflight and SQL behavior; retain green results afterward. Configuration/environment turns green before domain red gates; lifecycle/upgrade reds precede their entry preflights. Go race and lint gates apply to the harness. Synthetic fixtures only. |
| IV: Endpoint × role and page coverage | NOT APPLICABLE | NOT APPLICABLE | Feature exposes zero HTTP endpoints and zero pages. The endpoint × admin/editor/viewer matrix has zero rows; no browser tests are applicable. No broader coverage is waived. |
| V: Configuration, secrets, errors, observability | PASS | PASS | Only internal/config reads environment once into typed Config. Tests receive that config. No DSN/hash/credential values in logs. SQL diagnostics remain SQL errors; HTTP error envelopes/request IDs have no interface here. |
| Migration discipline | PASS | PASS | All feature schema changes, including fixture schema definitions, use versioned up/down SQL via golang-migrate. Independently selected histories prevent accidental finalization. |
| Scope, changelog, acceptance authority | PASS | PASS | Artifacts stay in this working tree; no push/merge/deploy. Every generated document records its purpose/change. Local checks permit review only. |

The requirements checklist has stale review prose describing O1/O2 as unanswered and older monetary definitions; the newer specification explicitly records the answers and R1 correction. This plan consumes those recorded answers rather than treating old checklist prose as a new blocker. Generic checklist objections to specifying SQL types and mandatory migration tools do not negate the operator's explicit schema requirement.

No unresolved requirement clarification or constitutional exception is needed. Remaining UNVERIFIED legacy case/whitespace, runtime, endpoint and screen evidence is a recorded evidence limitation; target database behavior has its own explicit validation gates.

## Project Structure

### Documentation (this feature)

```text
specs/001-complete-mysql-schema/
├── spec.md                         # Existing resolved requirements
├── plan.md                         # This plan
├── research.md                     # Decisions, rationale, alternatives
├── data-model.md                    # Complete target persisted contract
├── quickstart.md                    # Future runnable validation guide
├── contracts/
│   ├── schema-contract.md          # Persistence interface and case coverage
│   └── migration-contract.md       # Track selection, lifecycle, credentials
└── tasks.md                        # Future speckit-tasks output; not created here
```

### Source Code (repository root; proposed, not created by planning)

```text
go.mod
go.sum
internal/config/config.go
internal/config/config_test.go
migrations/
├── schema/000001_complete_mysql_schema.{up,down}.sql
├── credentials/000001_replace_legacy_password.{up,down}.sql
└── credentials-finalize/000001_require_password_hash.{up,down}.sql
tests/schema/
├── environment_test.go
├── definitions_test.go
├── values_test.go
├── references_test.go
├── lifecycle_test.go
├── credentials_test.go
├── helpers_test.go
└── testdata/
    ├── legacy/000001_legacy_user.{up,down}.sql
    └── sentinel/000001_unrelated_object.{up,down}.sql
compose.schema.yaml
.golangci.yml
CHANGELOG.md
```

**Structure Decision**: Keep migration streams under migrations and database acceptance tooling under tests/schema. The existing workspace contains planning infrastructure and specs, not an application implementation. The new config package serves only the validation tooling until later work uses it; it does not implement roadmap item 002's connection utility. The schema migration pair is the executable replacement for the incomplete mysql_database.sql; avoid a second hand-applied dump that could drift from it. Synthetic account and sentinel rows are test DML, separate from all schema migrations.

## Phase 0: Research Outcomes

[research.md](research.md) resolves the environment, migration packaging and validation decisions. Two research agents assessed migration lifecycle/credentials and MySQL constraints using the permitted documents and earlier local artifacts. No external schema, surrounding codebase or live database supplied estate facts.

Chosen environment: MySQL 8.4, InnoDB/utf8mb4, explicit binary collation and strict connection modes. Preserve orders.prodid VARCHAR(45) referencing product.pid VARCHAR(30); successful DDL application is a required compatibility test. DECIMAL excess-fraction rounding is a separate environment case, not strict-mode rejection. Omitted generated ids and explicit signed boundary ids use isolated fixtures.

Chosen lifecycle: bootstrap up creates an empty reset-ready schema; bounded down/up removes and recreates only the six tables in dependency order. Legacy credential replacement has its own source/history and retains account/profile rows. Finalization is a separately invoked migration with a SQL readiness guard and writers quiesced. Credential replacement down explicitly refuses before mutation; discarded passwords cannot be recovered and plaintext storage cannot return.

## Phase 1: Design and Handoff

[data-model.md](data-model.md) is the field/relationship contract. [schema-contract.md](contracts/schema-contract.md) specifies environment checks, SQL outcomes and acceptance coverage. [migration-contract.md](contracts/migration-contract.md) specifies track entry states, history, bounded teardown and credential transitions. [quickstart.md](quickstart.md) defines future setup and test commands and the evidence to retain.

The next tasks phase must separate callable scaffolding from acceptance-visible validation. First provide typed configuration loading without validation, read-only connectivity/inspection and non-enforcing validation entry points. Write TestConfig and TestEnvironmentPreflight and record assertions failing because invalid configuration or an unsupported environment is accepted, before implementing either refusal. Use valid runtime configuration for startup and injected Config values/isolated connection settings for negative cases; tests must reach the check rather than aborting at startup. Prove actual container connectivity independently of those missing checks. Then implement the checks and require TestConfig, TestEnvironmentPreflight and TestEnvironment green before domain red gates.

Keep occupied-bootstrap and incompatible-credential-entry checks non-enforcing in shared migration scaffolding. Lifecycle/upgrade tests must first demonstrate missing refusals by invoking those entry checks directly, with no migration mutation or source-file lookup: use the synthetic legacy user fixture as occupied bootstrap inventory and an empty fixture as incompatible credential entry. Record these red assertions alongside the existing lifecycle/upgrade SQL reds. Only afterward implement the occupied-inventory preflight with the complete-schema work and the credential-entry preflight with credential replacement. Once the target schema exists, also demonstrate its incompatible-credential-entry case red before implementing the credential preflight. Verify their refusal cases turn green before invoking the corresponding migration, asserting unchanged business objects and histories. Finalization retains its separate tests/red gate before its preflight and SQL guard.

Preserve the existing shared metadata/value/reference SQL red gate and independent fixtures for restriction versus cascade, last-reference cases and BIGINT exhaustion. Record each refusal's red/green commands and results and the three user journeys' results and SC-001–005 coverage before claiming implementation completion.

MySQL-specific assertions, runtime gates and exact dependency locks remain implementation validation work. A failing compatibility test blocks delivery; it never authorizes changing a specified field length, referential action or credential contract. No Go application architecture, HTTP contract or new business entity is needed for this feature.

## Complexity Tracking

No constitutional violations or exceptions. The three migration tracks implement distinct specified entry states and release conditions; they are not extra business components.

## Planning Validation

Document checks passed for required plan sections, links, declared goal/changelogs, six target entities and their 36 fields, and absence of unresolved clarification/template markers. The project prerequisite checker reports research.md, data-model.md, contracts/ and quickstart.md as available. tasks.md remains absent as required by this phase boundary.

No .specify/extensions.yml exists at the pre-plan or post-plan check, so both hook stages are skipped under the skill rules. No actual Git branch exists; the reported feature identifier comes from setup_plan.py. Database, Go build/lint/race and migration-in-Git gates remain unexecuted; documentation checks do not establish implementation or platform acceptance.

## Changelog

- 2026-11-10 — Applied spec FR-014's C1 sequencing correction to the java to go validation design: separate scaffolding from configuration/environment and migration-entry refusals, require their own red/green gates, and retain the existing SQL and finalization gates. No implementation or runtime checks performed.

- 2026-11-10 — Created the Phase 0/1 plan for work unit run-89ffa3642d6a4d44-wu, consuming the resolved Q1/Q2/R1 contract for the java to go schema foundation. No tasks, implementation or database execution performed.
