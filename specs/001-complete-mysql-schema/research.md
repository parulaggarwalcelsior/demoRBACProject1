# Research: Complete MySQL Schema

**Date**: 2026-11-10 | **Feature**: [spec.md](spec.md) | **Goal**: java to go

This research supports the Java-to-Go persistence foundation only. Two research agents independently assessed migration packaging/credentials and database constraints. Sources were entered through [legacy README](../../../../old/docs/README.md); only the permitted legacy documents and existing local artifacts supplied estate requirements. No source code outside the workspace, external schema, database execution or dependency installation was used.

## Evidence and clarification disposition

| Question | Resolution | Authority |
| --- | --- | --- |
| Missing cart, demand and transaction definitions | Use the supplied Q1 contracts, including InnoDB/utf8mb4 and global money defaults. | [spec.md](spec.md), E8 Q1; O1 resolved. |
| Erroneous parent names and identity lengths | user.email VARCHAR(100), product.pid VARCHAR(30), corrected child email/pid; retain orders.prodid VARCHAR(45). | Spec E8 R1; R1 resolved. |
| Credential layout and scope | Drop password; nullable password_hash VARCHAR(255), reset flag TRUE; separate existing-account migration; final NOT NULL after all resets. | Spec E8 Q2; O2 resolved; constitution III. |
| Supported environment and connection behavior | Explicit target choices below, verified during implementation rather than inferred from legacy. | Spec Assumptions and FR-012 delegate these to planning. |
| Older checklist prose still describes gaps | Use the later explicit answers in spec; preserve old review history without treating it as a new unanswered question. | Spec Clarifications and changelog supersede stale checklist notes. |

No unresolved requirement decision blocks Phase 1. No read document declares a target language conflicting with Go. The unavailable data-model/API/screen inventories remain evidence limitations, not proof that the estate lacks persistence or interfaces. Legacy equality/whitespace and runtime behavior remain UNVERIFIED; the target contract below does not claim legacy runtime parity for those unknowns.

## Decision 1 — Database environment

**Decision**: Validate on MySQL 8.4, InnoDB, utf8mb4 and explicit utf8mb4_0900_bin collation on all business tables and character FK columns. Use Go 1.24 as the harness language baseline. Lock exact toolchain/module versions and image digest in the implementation change and record the actual server patch in results.

**Rationale**: Q1 fixes engine/charset but leaves version/collation to planning. A named modern MySQL target makes enforced quantity CHECK constraints and input handling testable. Consistent collation avoids incompatible FK columns. The chosen binary collation makes target case/accent distinctions explicit; verify its NO PAD/trailing-space behavior on the selected server. This is a target environment choice, not evidence of legacy comparison rules.

**Alternatives considered**: Server defaults leave identity equality dependent on provisioning. A case-insensitive collation adds case folding without a legacy requirement. SQLite, MariaDB and older MySQL releases are not substitutes for this validation target. A SQL-only shell test suite would not satisfy the constitution's Go testing gates as directly.

## Decision 2 — Strict writes, exact money and generated identities

**Decision**: Require STRICT_ALL_TABLES, NO_ENGINE_SUBSTITUTION and NO_AUTO_VALUE_ON_ZERO on every test/migration connection; keep foreign_key_checks=1. Use ordinary single-row writes without IGNORE. Bind/read monetary values and out-of-range numeric probes as exact decimal strings, never float64. Compare actual stored values and driver SQL diagnostics.

**Rationale**: The required rejection of overlength/range inputs depends on connection behavior. NO_AUTO_VALUE_ON_ZERO permits explicit zero ids without inventing a positive-only rule. Omitted AUTO_INCREMENT ids are generated; explicit NULL also requests generation under native MySQL semantics and never persists a null primary key. Signed boundary/max-id cases need fresh tables because AUTO_INCREMENT allocation is not a gapless sequence and inserting the maximum can exhaust generation.

**Decision**: Separate two-place fidelity/range rejection from excess-fraction behavior. Target examples are 12.345 → 12.35 and -12.345 → -12.35 using exact decimal input; record stored values and diagnostics in an environment characterization case. Do not claim strict mode rejects every excess-fraction input. If actual MySQL behavior differs from these target expectations, keep that case failing and reconcile the environment documentation before handoff.

**Alternatives considered**: Floats hide precision failures. Positive-only money/id constraints or a universal fractional-input rejection trigger would invent business policy. Reusing a fixture after maximum-id insertion makes generation results misleading.

## Decision 3 — Foreign keys and checks

**Decision**: Keep the orders relationship named productid with NO ACTION on delete/update, while Q1 references use their explicit cascades/restrictions. Keep orders.prodid VARCHAR(45) and product.pid VARCHAR(30), with matching charset/collation. Name supporting indexes and the two enforced quantity checks deterministically; these are implementation identifiers, not new business rules.

**Rationale**: The specified differing VARCHAR lengths must be exercised on actual MySQL; they are not a reason to alter R1. Metadata verifies the child capacity of 45, while valid child writes can reference only existing parent ids of at most 30. Cart and transaction quantity CHECK constraints enforce >=1, with NOT NULL tested separately. No demand uniqueness, order-header table, order-to-transaction FK, positive-only stock/money or binary shipping constraint is authorized.

**Decision**: Include mixed-reference tests: order lines plus cascading cart/transactions must block product id change without partially rewriting children; transaction-restricted user deletion must retain cascading cart/demand rows too.

**Alternatives considered**: Forcing equal string lengths violates the operator contract. Disabling FKs to demonstrate a 45-character child write defeats the required behavior. Service existence checks/equality joins do not establish further unique keys or relationships.

## Decision 4 — Migration tracks, namespace and bounded reconstruction

**Decision**: Use golang-migrate v4 and three explicit source directories with separate history tables: schema, credentials, credentials-finalize. The container infrastructure provisions the namespace shopping-cart before migration connection. Bootstrap up creates six empty tables; down removes only those tables in dependency order. Reconstruct disposable synthetic state through down/up, keeping foreign keys enabled and unrelated objects intact.

**Rationale**: Default up must not apply a future credential-finalization step while reset-pending accounts exist. Existing legacy accounts must not traverse destructive bootstrap to receive Q2. Distinct histories prevent either accidental coupling. Database provisioning is an infrastructure prerequisite; all feature table changes remain versioned SQL. The supplied dump's truncated seed is not copied.

**Alternatives considered**: One linear stream would mix mutually different entry states or auto-finalize too early. A second standalone dump drifts from migration sources and invites hand application. DROP DATABASE or global FK disabling violates bounded-object behavior. Up replacing existing tables risks account loss; reconstruction belongs to the explicitly destructive lifecycle.

## Decision 5 — Existing credentials and the security boundary

**Decision**: Credential replacement applies only to a known legacy user table with password and without protected columns. Use one ALTER to add nullable password_hash, add required reset flag default TRUE and drop password. Existing accounts naturally initialize to NULL/TRUE without reading, copying or hashing the removed field. Test with synthetic profiles and only NULL legacy password values.

**Rationale**: Q2 authorizes this narrow account-preserving migration. It does not authorize converting all old tables or resizing existing parent keys; successful credential-only upgrade proves Q2, not full Q1/R1 conformance of an existing database. Authentication supplies encoded hashes and consumes reset flags in a later feature.

**Decision**: Supply a versioned credential-replacement down migration that deliberately refuses before mutation. It must never recreate the password column or discard protected hashes. Report the irreversible security boundary and use forward repair rather than pretending rollback succeeded. This implements the spec's prohibition on plaintext restoration while satisfying migration packaging; it is a documented limitation, not a reversible downgrade promise.

**Alternatives considered**: Restoring a nullable plaintext column reintroduces prohibited storage. Copying/hashing old values violates Q2. A no-op down misrepresents history as reverted while retaining an upgraded schema. Dropping protected columns would lose new credentials without restoring old ones.

## Decision 6 — Finalization, failure and concurrent writers

**Decision**: Finalization is opt-in, after a readiness query finds neither a NULL hash nor a TRUE reset flag. Require account writers quiesced throughout preflight, guard and ALTER. The SQL migration repeats the readiness guard and aborts before altering nullability when any account is pending. Plan a migration-owned short-lived routine using SQLSTATE 45000 for the guard; no client DELIMITER directive belongs in a golang-migrate SQL file. Finalization down restores only nullable hash storage and retains data.

**Rationale**: A migration-history lock does not stop application account writes. Checking null hashes alone overlooks reset-flagged accounts with non-null hashes. MySQL multi-statement DDL is not an all-or-nothing transaction. The guard's own helper is declared administrative scope and must be absent after successful completion; failed routines/history need explicit inspection and cleanup/forward repair, never blind force.

**Alternatives considered**: Automatically applying finalization at bootstrap makes future reset-pending inserts incompatible. Application-only guarding lets direct migration invocation bypass readiness. A persistent reset CHECK would impose an unsupplied ongoing business rule. Treating ALTER and preliminary queries as one rollback-safe transaction is unsupported.

## Decision 7 — Validation and layer boundaries

**Decision**: Tests consume a typed Config read once by internal/config. The harness imports the migration and MySQL drivers, uses table-driven Go tests, and creates fixture schema through versioned SQL. Use metadata plus behavioral tests and before/after snapshots. Endpoint/role matrix and browser coverage have zero applicable rows in this feature; the SQL interface is its external contract.

**Rationale**: This meets the constitution without implementing DBUtil, models, services, sessions or handlers. Schema creation returns zero business records. Fixture DML is synthetic and separate; no network mail or other external business side effect is permitted. Red failures must prove missing constraints/behavior, not merely a broken environment.

**Alternatives considered**: Metadata-only checks do not prove enforcement. Mocks cannot establish MySQL referential behavior. Introducing a web framework, Argon2 library or production DAO just for schema tests expands scope and adds unused dependencies.

## Changelog

- 2026-11-10 — Resolved planning choices for the java to go MySQL foundation in work unit run-89ffa3642d6a4d44-wu. Runtime compatibility and validation remain unexecuted implementation gates; no platform acceptance is claimed.
