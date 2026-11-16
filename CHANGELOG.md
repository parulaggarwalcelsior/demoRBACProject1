# Changelog

## 2026-11-16

- Implemented T005 Java-to-Go schema harness refusals and migration scaffolding: required configuration validation, namespace/environment preflight refusal, redacted DSN diagnostics, Config-passed helpers, golang-migrate v4 file/MySQL imports, explicit feature/fixture histories, scoped cleanup/snapshot utilities and SQL code/SQLSTATE assertion support.
- Added T004 Java-to-Go schema validation scaffolding: typed `internal/config` loading from `SCHEMA_TEST_DSN`, read-only schema test helpers, non-enforcing configuration/environment validation entry points, table-driven red-gate tests, and the MySQL SQL driver pin.
