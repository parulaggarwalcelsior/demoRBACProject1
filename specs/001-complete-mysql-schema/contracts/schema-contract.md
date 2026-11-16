# Schema Contract and Acceptance Coverage

**Date**: 2026-11-10 | **Goal**: java to go  
**Authority**: [spec.md](../spec.md), FR-001–016 and US1–3. Complete field/relationship definitions: [data-model.md](../data-model.md). Migration interface: [migration-contract.md](migration-contract.md).

The database is this feature's external interface to the future Go persistence layer. It exposes the six declared business tables, SQL write/read behavior and explicit migration entry points. It exposes no HTTP endpoint or page. The endpoint × admin/editor/viewer matrix has zero rows; SQL connection privileges are not application roles or membership enforcement.

## Environment preconditions

Use actual MySQL 8.4 with InnoDB. Every business table uses utf8mb4/utf8mb4_0900_bin, including compatible referenced/referencing character columns. Configure and inspect every validation/migration connection for STRICT_ALL_TABLES, NO_ENGINE_SUBSTITUTION, NO_AUTO_VALUE_ON_ZERO, foreign_key_checks=1, utf8mb4 client encoding and UTC session time zone. The named collation's target case/accent/trailing-space semantics must be recorded; do not claim these were discovered in legacy.

Reject a missing database configuration, wrong namespace, unsupported server or missing strict/integrity setting rather than skipping tests. The connection's SQL mode is part of the supported environment: DDL alone does not promise strict rejection for arbitrary permissive clients. Use ordinary writes, without IGNORE, REPLACE or upsert behavior that hides expected failures.

Inspect INFORMATION_SCHEMA and SHOW CREATE TABLE for columns, sizes, signedness, PK order, unique keys, FK targets/actions, CHECK enforcement, defaults, engine, charset and collation. Normalize metadata spelling such as MySQL BOOLEAN/TINYINT aliases and default-expression display without weakening semantics. Snapshot business rows before rejected parent mutations and compare every affected relation afterward.

## Behavioral case matrix

Every row below becomes table-driven Go acceptance cases before the corresponding implementation. Test omission, explicit NULL, valid values and each independent failure separately so one violated constraint cannot hide another.

| Case group | Required observations | Specification coverage |
| --- | --- | --- |
| Inventory and clean setup | Six target tables, no extra header/users/books relation or obsolete columns; zero business rows; complete statements; InnoDB/utf8mb4; metadata matches the field register. | US1.1/3; FR-001/003/013; SC-001/002. |
| Removal/recreation | Populated synthetic schema down/up succeeds, same definitions, constraints still enforced, unrelated sentinel definition/data retained. | US1.2/4, US3.8; FR-002; SC-001/004. |
| Product optional/default fields | Only pid required; omitted/null optional fields remain NULL, price omission 0.00, explicit NULL price fails; small synthetic binary image retained unchanged. | US2.1; FR-004/010–011. |
| Character limits | Exact limits and one over for every VARCHAR: product 30/100/20/350, user 100/30/250/hash255, orderid45/prodid45, Q1 email100/pid30/transaction_id40/book_name100/author100. Include utf8mb4 multibyte characters. No truncation of successful values. | US2.2–3/7–8/11 and line-identity edge case; FR-003–009/011. |
| Required values and NULL | Non-generated PK/reference fields reject omission/NULL, optional fields retain NULL, required money/shipping/reset/timestamps reject NULL. Transaction quantity omission fails. Generated ids follow native omission/NULL generation; no null PK is persisted. | US2.2–4/7–8; FR-005–011. |
| Order-line defaults | Quantity omission/null stays NULL, amount defaults 0.00, shipped defaults 0; null money/shipped fails. | US2.4; FR-006/010–011. |
| Exact money and range | Each of pprice/orders.amount/transactions.amount retains 0.00,12.34,99999999.99,-99999999.99 exactly; ±100000000.00 fails; omission 0.00; explicit NULL fails. No floating-point comparisons. | US2.5; FR-010; SC-003. |
| Permitted signed values | Zero/negative stock and line quantity allowed, negative money allowed, shipped=2 allowed. Do not add positive-only stock/line/money or binary-only shipping checks. | US2.6/9; FR-012. |
| Integer boundaries | Signed INT endpoints for stock/line quantity/pincode/shipped and BIGINT endpoints for mobile retained; one beyond each fails without clipping. | US2.9; FR-004–006/012; SC-003. |
| Q1 generated identities | Each id omitted twice generates distinct BIGINTs; duplicate explicit id fails; signed endpoints and zero retained on isolated fixtures; one beyond range fails. Do not assume rollback resets allocation or max-id insertion permits future generation. | US2.7–8/10; FR-008–009/012. |
| Quantity checks and timestamps | Cart quantity default1; cart/transaction 1 and INT max accepted, NULL/0/negative rejected, max+1 rejected; transaction omission rejected. Timestamps omission between server time bounds, explicit valid timestamp exact, NULL rejected. | US2.8/10; FR-008–009. |
| Duplicate identities/pairs | Identical product/user/transaction identities rejected without changing originals; duplicate order/product pair rejected; distinct allowed combinations succeed; duplicate cart pair rejected; distinct cart pairs succeed. | US3.1–2/9; FR-004–009; SC-004. |
| Absent parents | For every FK, insert and update to nonexistent parent fail. orders.prodid metadata remains45; a matching <=30 parent succeeds, nonexistent 31–45 child fails FK, >45 also violates storage length. Do not disable FKs to test a successful 45-character child. | US3.3/9 and identity edge case; FR-003/006–009. |
| Historical product protection | Delete and id update fail with one/several order references, all rows unchanged; removing all but final reference still blocks; after last removed each operation succeeds in separate fixtures if no other restricting references. Optional product metadata update remains allowed. | US3.4–7; FR-007; SC-004. |
| Q1 delete/update actions | Separate fixtures prove user→cart/demand delete cascades, transaction→user delete restricts, cart/transaction→product delete restricts; allowed account/product identity updates cascade to all Q1 children. | US3.10–11; FR-008–009. |
| Mixed reference atomicity | Product referenced by orders and Q1 children: rejected id change leaves every child intact. User referenced by transactions plus cart/demand: rejected delete leaves parent and cascading children intact. | US3.4–5/10–11; FR-007–009; SC-004. |
| Demand duplicates | Repeated email/book_name/author rows with different ids succeed; author omission/NULL accepted. No unsupplied demand unique key or prodid/quantity field. | US2.8, US3.9; FR-009/012. |
| Credential replacement | Separate legacy fixture upgrade retains identities/profile/count, drops password, initializes NULL hash/reset TRUE; no password value read/copied/converted. Refused replacement down retains protected fields/data. | US1.5, US2.3/11; FR-005/013/015–016. |
| Hash capacity | Full synthetic encoded Argon2id string retained exactly; 255-character storage-boundary probe retained, 256 rejected. SQL capacity tests do not claim password/hash validation. | US2.11; FR-005; SC-003. |
| Guarded hash finalization | NULL/TRUE, NULL/FALSE and non-null/TRUE states each block; all non-null/FALSE allows NOT NULL; hashes unchanged; subsequent NULL/omission fails. Down restores nullable state with hashes/flags/profiles intact. | US2.12; FR-005/015; SC-003. |
| Track/history behavior | Each source uses its own history; repeated applied up has no changes, finalization never runs implicitly, successful steps clean, refused/failed steps surfaced without blind force. | FR-002/014–016; SC-001/005. |

## Environment characterization

Excess fractional digits are separate from the exact two-place acceptance range. On the chosen target, check exact decimal inputs 12.345 and -12.345 against expected stored values 12.35 and -12.35; capture SQL warnings/notes as well. Strict mode must not be described as a universal fractional-digit rejection mechanism. Other rounding and collation behavior is a target environment characteristic, not an additional inferred legacy business rule.

For author omission, accept native nullable implicit NULL; do not misdescribe Q1 as supplying an explicit author default. For required identities, no nonempty or format checks are supplied. Compare duplicate fixtures using identical strings; characterize case/whitespace separately under the declared target collation.

## SQL failures and result evidence

Rejected writes must yield a driver-recognized SQL error of the expected category: duplicate key, required-value/default failure, overlength, out-of-range, FK child-write refusal, FK parent-mutation refusal or quantity CHECK violation. Record and assert the selected server's numeric code/SQLSTATE; do not couple tests to localized message wording. Guard refusal uses SQLSTATE 45000. An arbitrary database error is not a passing negative case.

Tests must check the absence of unexpected inserts/updates after failure, including every child touched by possible cascades. DML atomicity checks do not imply migration DDL atomicity. Tests keep synthetic decimal/id boundary inputs as strings where Go numeric range or floating-point conversion could hide an invalid input.

Retain per-case status, environment/version data, red-before-green evidence and a summary by US1–3 and SC-001–005. Every case must be executed or explicitly reported unexecuted/failed; skipped prerequisites are not passes. Runtime and build/lint checks are implementation gates, none executed during planning. Referee/platform gates retain acceptance authority.

## Changelog

- 2026-11-10 — Defined the SQL interface and complete acceptance coverage for the java to go MySQL foundation, work unit run-89ffa3642d6a4d44-wu. No endpoints, page contracts or executable test suite introduced.
