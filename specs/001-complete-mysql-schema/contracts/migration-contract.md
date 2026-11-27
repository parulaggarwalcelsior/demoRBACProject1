# Migration Contract: mysql_database

**Date**: 2026-11-10 | **Goal**: java to go  
**Requirements**: [spec.md](../spec.md), FR-001–003/005/013–016, Q2; [constitution](../../../.specify/memory/constitution.md), migration constraints.

The Java-to-Go schema is consumed through golang-migrate versioned SQL sources. This document defines the future delivery interface; no migrations are implemented or applied by this planning invocation.

## Explicit migration tracks

| Track | Source directory | Version pair | History table | Entry state |
| --- | --- | --- | --- | --- |
| Complete schema | migrations/schema | 000001_complete_mysql_schema.up.sql / .down.sql | schema_migrations_schema | Existing isolated shopping-cart namespace with none of the six business tables. |
| Existing credentials | migrations/credentials | 000001_replace_legacy_password.up.sql / .down.sql | schema_migrations_credentials | Known legacy user table with password, without password_hash/must_reset_password. |
| Final credentials | migrations/credentials-finalize | 000001_require_password_hash.up.sql / .down.sql | schema_migrations_credentials_finalize | Either initial target or credential-upgraded user table, with every account reset and writers quiesced. |

Paths are relative to the repository root. Each track starts at version 1 independently; the migration driver must configure its history table explicitly. No command selects all three sources together. Histories are tooling metadata, not business relations, and must be included in the administrative-object inventory. A second run against an already applied clean version has no pending changes and must not reinitialize rows.

Infrastructure supplies the MySQL instance, shopping-cart namespace and connection account. The disposable container may use MYSQL_DATABASE to provision that namespace; do not put credentials in source. All feature table/routine changes and fixture schema definitions run through golang-migrate. Never import the original SQL dump or apply CREATE/ALTER/DROP statements by hand.

The harness reads environment once through internal/config into typed Config and passes connection/migration settings to helpers. It configures each source/history explicitly, sets the connection modes from [schema-contract.md](schema-contract.md), and redacts all DSNs, hashes and credential material from diagnostics. A one-connection session setting is insufficient for a pool or a separate migration driver connection.

Under spec FR-014, connection/migration scaffolding initially provides callable, non-enforcing check entry points. Configuration/environment refusal tests must demonstrate missing checks before their implementation and turn green before domain red gates. Occupied-bootstrap and incompatible-credential-entry preflights remain non-enforcing until lifecycle/upgrade tests demonstrate their missing refusals; test those checks directly without migration mutation or unauthored-source lookup. Use the legacy user fixture for occupied bootstrap and an empty fixture for incompatible credential entry; also record the initial target schema's incompatible-entry red once that schema exists, before implementing the credential preflight. Implement each preflight with its corresponding track and prove refusal before migrate is invoked, with unchanged business objects and histories. These are additional Go red/green gates, preserving all SQL red gates and the separate finalization gate.

## Complete schema lifecycle

Before up, inspect the entire owned business inventory. No in-scope table may already exist; do not silently adopt, overwrite or merge an existing schema. Use CREATE TABLE without IF NOT EXISTS so a violated prerequisite also fails at the database boundary. Infrastructure and history metadata may exist. Reject the wrong namespace and unsupported environment before feature DDL.

Up creates product and user first, then orders, usercart, user_demand and transactions. Its result is the initial credential state from [data-model.md](../data-model.md), with zero business rows and all constraints enforced. No legacy password column, business seed, partially copied image INSERT or separate order-header table is present.

Down drops only orders, usercart, user_demand, transactions, product and user, in that order. Bounded DROP TABLE IF EXISTS is appropriate for the declared inventory; do not enumerate arbitrary database objects. Preserve the namespace, per-track history management and every unrelated object. Keep foreign_key_checks=1 throughout. An unrelated external reference may block removal; report that failure instead of dropping the external object or bypassing its constraint.

Reconstruction means down followed by up in a disposable environment. It intentionally discards in-scope synthetic records and restores the same definitions, including nullable initial hashes. Snapshot unrelated sentinel definitions/data before and after. If finalization was applied, first down its track, then down/up the complete schema; an old applied finalization history must not silently outlive a reconstructed initial schema.

## Existing-account credential replacement

Preflight verifies the expected legacy user columns and absence of the target credential columns. Snapshot synthetic account identities/profile values/counts without reading password contents. The fixture's legacy password values are NULL only; no plaintext fixture data is needed.

Up uses a single ALTER of user to drop password and add password_hash VARCHAR(255) NULL DEFAULT NULL plus must_reset_password BOOLEAN NOT NULL DEFAULT TRUE. Existing rows initialize to NULL/TRUE through the new columns' defaults. No UPDATE reads password, no hash is derived from it, and no account/profile row is deleted or reinserted.

Successful upgrade must preserve all profile values, identities, row counts and unrelated definitions/data. This track changes credentials only. It does not resize pre-existing user.email or transform old cart/transaction/demand layouts. Consequently it proves Q2, not complete target-schema conformance of an arbitrary legacy database. The complete-schema track proves Q1/R1 target conformance independently.

Down is packaged as a versioned SQL migration that deliberately raises an error before business-schema mutation. The error states that protected-credential replacement cannot be downgraded to plaintext storage. It leaves protected columns, hashes and profiles intact. A refused down is not a successful downgrade; do not mark history as reverted. golang-migrate may mark the attempted step dirty, which must be reported and inspected. Recovery uses a reviewed forward repair; no recreated password field, discarded hash, copied plaintext or blind metadata force is allowed.

## Guarded finalization

The application authentication feature eventually writes complete encoded Argon2id hashes and clears reset flags. This schema feature supplies the migration; it does not perform those sign-in operations. Tests simulate readiness with synthetic encoded strings and ordinary DML.

Before finalization, quiesce account writers for the full preflight/guard/ALTER interval. Check that there is no row with password_hash IS NULL or a truthy must_reset_password flag. A nonzero flag is pending under MySQL boolean evaluation; do not treat a non-null hash alone as readiness. An empty table is vacuously ready, but that does not authorize automatic finalization during ordinary bootstrap.

The migration itself repeats the SQL readiness guard before changing nullability. A migration-owned short-lived procedure named schema_guard_finalize_credentials may perform the conditional SQLSTATE 45000 refusal and ALTER. The file is server SQL consumed by golang-migrate, without the mysql client's DELIMITER directives. This helper is administrative scope only and must be removed after successful completion. No permanent guard table or new reset constraint is added.

On a pending-account refusal, hash nullability and every business value must remain unchanged. Normal harness preflight should reject before invoking migrate; test the SQL guard independently through migrate as well. If that invocation leaves dirty history or a guard routine after failure, report the exact administrative state and clean it through a versioned, scoped repair path or discard that isolated test instance. Never claim multi-statement DDL rolled back transactionally.

When all accounts are ready, up changes password_hash to VARCHAR(255) NOT NULL, with no default. Preserve all hashes, flags and profile values. Subsequent NULL hash writes and omission fail. Down restores VARCHAR(255) NULL DEFAULT NULL only; it retains hashes/flags/profiles and never recreates password. Test non-null/TRUE and null/FALSE pending states independently, plus all-ready success and down preservation.

## Failure and ownership rules

- Check successful migration version and dirty=false after each successful operation; do not infer clean history from table existence alone.
- MySQL DDL may commit earlier statements before a later statement fails. Abort on the first migration error, retain the original diagnostic and inspect owned state. No automatic retry, generic rebuild or force-to-version command is part of normal behavior.
- Fixture schema migrations under tests/schema/testdata/legacy and sentinel use their own history names. Fixture rows are minimal synthetic DML. Helpers clean up only objects they own, and may discard an entire explicitly disposable test container.
- Initial setup and successful recreation expose six business tables plus declared administrative metadata. Successful finalization leaves no guard routine. No unrelated object is a cleanup target.
- Migration artifacts must be tracked in the implementation repository with up/down pairs and a changelog. The planning workspace currently has no Git repository; no version-control, database or implementation gate has passed here.

## Changelog

- 2026-11-10 — Propagated C1's acceptance-first sequencing to the java to go migration interface: callable scaffolding precedes tests/red evidence, with entry-preflight implementation deferred to its track and refusal-green evidence required before mutation. No entry-state or SQL contract changed.

- 2026-11-10 — Defined migration selection, bounded destruction, Q2's irreversible security boundary and guarded finalization for the java to go foundation, work unit run-89ffa3642d6a4d44-wu.
