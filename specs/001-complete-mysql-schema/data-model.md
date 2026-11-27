# Data Model: Complete MySQL Schema

**Date**: 2026-11-10 | **Goal**: java to go | **Authority**: [spec.md](spec.md), especially the confirmed field contract and E8 Q1/Q2/R1.

This is the persistence interface for the Java-to-Go rebuild, scoped to mysql_database. It defines six business tables in `shopping-cart`. All use InnoDB and utf8mb4; utf8mb4_0900_bin is the explicit target collation chosen in [research.md](research.md). Each field below traces to a preserved legacy rule or a recorded operator replacement. No Java bean default supplies a missing database attribute.

`None` in the default column means no literal default: callers must supply required fields. AUTO_INCREMENT is generation rather than a constant default. All INT and BIGINT columns are signed. The initial target includes protected credentials immediately; it never includes legacy password.

## Product — product

| Field | Type | Nullable | Default | Identity / evidence |
| --- | --- | --- | --- | --- |
| pid | VARCHAR(30) | No | None | Primary key; R1 replaces legacy length 45, PRD mysql_database FR-001–003. |
| pname | VARCHAR(100) | Yes | NULL | PRD FR-004–005. |
| ptype | VARCHAR(20) | Yes | NULL | PRD FR-006–007. |
| pinfo | VARCHAR(350) | Yes | NULL | PRD FR-008–009. |
| pprice | DECIMAL(10,2) | No | 0.00 | Q1 replaces PRD FR-010–011. |
| pquantity | INT | Yes | NULL | PRD FR-012. |
| image | LONGBLOB | Yes | NULL | PRD FR-013. |

No generated product identity, positive-only price/stock rule, required image or placeholder bytes are specified. Null stock is distinct from zero. Required identity is non-null; no email/nonempty/format normalization policy is added.

## Account — user

| Field | Type | Nullable | Default | Identity / evidence |
| --- | --- | --- | --- | --- |
| email | VARCHAR(100) | No | None | Primary key; R1 replaces PRD FR-027–029 length 60. |
| name | VARCHAR(30) | Yes | NULL | PRD FR-030–031. |
| mobile | BIGINT | Yes | NULL | PRD FR-032. |
| address | VARCHAR(250) | Yes | NULL | PRD FR-033–034. |
| pincode | INT | Yes | NULL | PRD FR-035. |
| password_hash | VARCHAR(255) | Yes initially; No after finalization | NULL initially; no default after finalization | Q2 replaces nullable password VARCHAR(20), PRD FR-036–037; full encoded Argon2id capacity. |
| must_reset_password | BOOLEAN | No | TRUE | Q2; existing accounts initialize TRUE. |

BOOLEAN follows MySQL's native representation; no extra binary-only CHECK is authorized. SQL storage capacity does not validate an Argon2id encoding or implement authentication. The final migration removes the initial null default when making the hash required. Profile values, identity and protected hashes survive the account-preserving migration paths.

## Cart entry — usercart

| Field | Type | Nullable | Default | Identity / evidence |
| --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | No | Generated | Primary key; Q1. |
| email | VARCHAR(100) | No | None | FK user.email; Q1 corrected by R1. |
| pid | VARCHAR(30) | No | None | FK product.pid; Q1 corrected by R1. |
| quantity | INT | No | 1 | Enforced CHECK >=1; Q1. |
| added_at | DATETIME | No | CURRENT_TIMESTAMP | Q1. |

UNIQUE (email, pid) permits one cart row per corrected account/product pair. Indexes required for FKs support the relationship; they do not introduce more uniqueness. Timestamps are stored DATETIME values, without an automatic update clause or fractional-second extension.

## Order line — orders

| Field | Type | Nullable | Default | Identity / evidence |
| --- | --- | --- | --- | --- |
| orderid | VARCHAR(45) | No | None | First primary-key member; PRD FR-014–015/018. |
| prodid | VARCHAR(45) | No | None | Second primary-key member, FK product.pid; PRD FR-016–018/024–026. |
| quantity | INT | Yes | NULL | PRD FR-019. |
| amount | DECIMAL(10,2) | No | 0.00 | Q1 replaces PRD FR-020–021. |
| shipped | INT | No | 0 | PRD FR-022–023. |

PRIMARY KEY (orderid, prodid) allows distinct products in one order and the same product in different orders. No separate order-header entity or order-to-transaction FK is established. The productid FK has NO ACTION on both parent delete and update. Do not impose positive-only quantity or binary-only shipped rules; signed values, including shipped=2, remain valid.

The child column keeps its specified capacity of 45 although product.pid allows only 30. Metadata proves the child length; successful writes require a matching parent of at most 30. Child strings of 31–45 characters fail referential integrity, not a newly invented 30-character child declaration.

## Transaction — transactions

| Field | Type | Nullable | Default | Identity / evidence |
| --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | No | Generated | Primary key; Q1. |
| transaction_id | VARCHAR(40) | No | None | UNIQUE; Q1 replaces PRD FR-038–040 transid identity. |
| email | VARCHAR(100) | No | None | FK user.email; Q1/R1. |
| pid | VARCHAR(30) | No | None | FK product.pid; Q1/R1. |
| quantity | INT | No | None | Enforced CHECK >=1; Q1. |
| amount | DECIMAL(10,2) | No | 0.00 | Q1 general monetary rule. |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Q1. |

The legacy transid/time/username fields are superseded, not retained aliases. Unique transaction_id and generated row identity are distinct contracts. No transaction-to-order reference is added from a historical equality join.

## Demand request — user_demand

| Field | Type | Nullable | Default | Identity / evidence |
| --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | No | Generated | Primary key; Q1. |
| email | VARCHAR(100) | No | None | FK user.email; Q1/R1. |
| book_name | VARCHAR(100) | No | None | Q1 replacement name retained exactly despite the electronics domain. |
| author | VARCHAR(100) | Yes | Implicit NULL | Q1 permits NULL without supplying an explicit default. |
| requested_at | DATETIME | No | CURRENT_TIMESTAMP | Q1. |

No product id, quantity, demanded-product pair uniqueness or books parent is part of this replacement. Repeated otherwise identical requests with distinct ids are valid. Omitted author follows nullable MySQL behavior; explicit NULL is accepted.

## Relationship and index contract

| Child → parent | Constraint name | ON UPDATE | ON DELETE | Authority |
| --- | --- | --- | --- | --- |
| orders.prodid → product.pid | productid | NO ACTION | NO ACTION | PRD FR-024–026, spec FR-007. |
| usercart.email → user.email | fk_usercart_email | CASCADE | CASCADE | Q1/R1, spec FR-009. |
| usercart.pid → product.pid | fk_usercart_pid | CASCADE | RESTRICT | Q1/R1, spec FR-009. |
| user_demand.email → user.email | fk_user_demand_email | CASCADE | CASCADE | Q1/R1, spec FR-009. |
| transactions.email → user.email | fk_transactions_email | CASCADE | RESTRICT | Q1/R1, spec FR-008–009. |
| transactions.pid → product.pid | fk_transactions_pid | CASCADE | RESTRICT | Q1/R1, spec FR-008–009. |

Only productid's name is legacy-required; the other names are design identifiers. Use uq_usercart_email_pid and uq_transactions_transaction_id for the two supplied unique keys, and ck_usercart_quantity/ck_transactions_quantity for quantity checks. Explicit nonunique indexes support orders.prodid, usercart.pid and each transaction FK; usercart's composite unique index already starts with email. A nonunique email index supports user_demand. No speculative performance indexes are required.

Deleting a product or changing its identity fails while any order line remains. Q1 product references also restrict deletion and otherwise cascade identity updates. Mixed-reference failure must leave every affected row unchanged. User deletion cascades only when no transaction restricts it; tests must prove that a failed restricted delete does not partially delete cart/demand rows.

## Numeric and timestamp validation

- INT range: -2147483648 through 2147483647; BIGINT range: -9223372036854775808 through 9223372036854775807. Test one beyond each bound without converting probes to floating point. Only cart/transaction quantity is restricted to >=1.
- All three monetary fields use signed DECIMAL(10,2), range -99999999.99 through 99999999.99, NOT NULL, default 0.00. Preserve exact two-place values and reject out-of-range magnitudes. Excess-fraction conversion is separately characterized as specified in research and the schema contract.
- AUTO_INCREMENT omission generates distinct identities; explicit signed boundaries and zero are separate fixtures. Explicit NULL requests generation under native MySQL semantics; no null primary key is stored. No sequence continuity or gaplessness is promised.
- Timestamps default to the server's current DATETIME in a declared UTC validation session. Explicit valid timestamps remain unchanged; NULL is rejected. No application timestamp or clock-generated business identity is introduced.

## Credential transitions

| State | Hash storage | Reset flag | Permitted schema transition |
| --- | --- | --- | --- |
| Legacy credential fixture | Legacy nullable password; no target hash fields | Absent | Credential replacement only; discard password column, never its contents into another field. |
| Initial target / upgraded account | Nullable password_hash; existing accounts NULL | Required; existing/new default TRUE | Authentication later writes encoded hashes and clears flags; no such application code here. |
| Ready for finalization | Every hash non-null | No TRUE flag remains | Guarded versioned transition may require hash NOT NULL. |
| Finalized | password_hash NOT NULL; no default | Preserved | Finalization down relaxes nullability only; it never restores plaintext. |

The SQL guard must refuse both null-hash and flagged-account states before altering nullability. Writer quiescence is an operational prerequisite. The legacy credential-only track leaves pre-existing profile column definitions unchanged; it is not a full legacy-to-target conversion.

## Legacy rule accounting and differences

| Legacy PRD mysql_database rules | Target disposition | Named difference |
| --- | --- | --- |
| FR-001–003 | Preserve required product PK; R1 changes length 45 → 30. | D5 |
| FR-004–009 | Preserve product descriptive limits/null/defaults. | None |
| FR-010–011 | Replace nullable DECIMAL(12,2) with required DECIMAL(10,2), default 0.00. | D4 |
| FR-012–013 | Preserve signed nullable stock and optional LONGBLOB. | None |
| FR-014–019 | Preserve order-line identities/composite key and nullable signed quantity. | None |
| FR-020–021 | Preserve DECIMAL(10,2) precision; require amount/default 0.00. | D4 |
| FR-022–026 | Preserve integer shipping/default and productid NO ACTION reference. | None |
| FR-027–029 | Preserve required email PK; R1 changes length 60 → 100. | D5 |
| FR-030–035 | Preserve profile field sizes/types/null/defaults. | None |
| FR-036–037 | Replace password with Q2 protected fields/staged migrations. | D2 |
| FR-038–040 | Replace transid VARCHAR(45) PK with Q1 generated id and unique transaction_id VARCHAR(40). | D3 |

D1 repairs the truncated artifact without business seeds. Q1 supplies the additional cart/demand/transaction fields; Q1/R1 authorize their references and checks. The register accounts for all 40 original rules by preservation or deliberate replacement. [PRD](../../../../old/docs/prd.md#mysql_database-40), [BRD](../../../../old/docs/brd.md#mysql_database-40) and [LLD](../../../../old/docs/lld.md#mysql_database) corroborate the legacy side; E8 in the spec is the replacement authority.

## Changelog

- 2026-11-10 — Designed the six-table Java-to-Go contract for work unit run-89ffa3642d6a4d44-wu, with complete field, relationship, credential-state and rule-disposition accounting. No executable DDL or runtime validation produced.
