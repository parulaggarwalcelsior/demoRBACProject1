# Feature Specification: Complete MySQL Schema

**Feature Branch**: None created; no branch hook is registered.

**Created**: 2026-11-10

**Work unit**: run-8fd0f0975fa64942-wu

**Short name**: complete-mysql-schema

**Status**: Clarified by operator replacement contract; not a draft. This is a specification, not executable DDL or a claim of implementation acceptance. The operator has resolved the parent-reference reconciliation under R1.

**Input**: Specification 1, `specs/001-complete-mysql-schema`, for component `mysql_database`. Specify executable MySQL DDL completing the supplied schema, including dropping/creating objects, verified field lengths, nullability, defaults, uniqueness, foreign keys, monetary precision, and protection of products referenced by order lines. Exclude general data migration and seed data except synthetic constraint-validation fixtures; Q2 explicitly includes the credential-column migration and existing-account reset initialization without migrating plaintext. The operator declares source language java, target language go, and the goal, verbatim, "java to go".

This feature gives the Java-to-Go rebuild a reproducible persistence foundation. It specifies only roadmap item 001, including Q2's credential columns and migration. Database utilities, Go models/services, HTTP flows, sign-in/hash-writing code, sessions, authorization workflows, email, checkout orchestration, application identity generation, and client assets remain separate work. The earlier broad rebuild specification supplies context but does not expand this invocation's scope.

## Clarifications

### Session 2026-11-10

- Q: Which permitted document or explicit operator replacement contract supplies the missing usercart, user_demand and transactions field definitions, lengths, nullability, defaults, keys, monetary precision and foreign-key actions for the Java-to-Go schema? If none is available, should these attributes remain UNVERIFIED pending evidence, leaving this feature a draft? → A: No document defines these tables; the operator supplies them here, so they are verified and the feature is not a draft. All tables InnoDB, utf8mb4. Money is DECIMAL(10,2) NOT NULL DEFAULT 0.00. Foreign keys ON UPDATE CASCADE, ON DELETE RESTRICT unless stated. usercart: id BIGINT PK AUTO_INCREMENT; email VARCHAR(100) NOT NULL FK user.email ON DELETE CASCADE; pid VARCHAR(30) NOT NULL FK product.pid ON DELETE RESTRICT; quantity INT NOT NULL DEFAULT 1 (at least 1); added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP; UNIQUE (email, pid). user_demand: id BIGINT PK AUTO_INCREMENT; email VARCHAR(100) NOT NULL FK user.email ON DELETE CASCADE; book_name VARCHAR(100) NOT NULL; author VARCHAR(100) NULL; requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP. transactions: id BIGINT PK AUTO_INCREMENT; transaction_id VARCHAR(40) NOT NULL UNIQUE; email VARCHAR(100) NOT NULL FK user.email ON DELETE RESTRICT; pid VARCHAR(30) NOT NULL FK product.pid ON DELETE RESTRICT; quantity INT NOT NULL (at least 1); amount DECIMAL(10,2) NOT NULL; created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP.
- Q: To honor the constitution's already chosen Argon2id requirement, should this schema feature replace or enlarge the legacy nullable password field with an explicitly defined protected-credential contract, add separately specified credential storage, or retain the legacy field unused until a named later schema change? Please supply the chosen field definition and corresponding scope disposition; plaintext storage remains prohibited. → A: Replace it. Drop the legacy nullable password column and add password_hash VARCHAR(255), holding the full argon2id string (algorithm, parameters, salt, hash). Add must_reset_password BOOLEAN NOT NULL DEFAULT TRUE. Existing users get password_hash NULL and must reset on next sign-in; password_hash becomes NOT NULL once every user has reset. Plaintext is never stored or migrated. In scope here: the columns and the migration. Out of scope: the sign-in code that writes the hash, which belongs to the authentication feature.

- Q: Q1 supplies references to users.username and books.barcode, while E3/E4 define user.email and product.pid; definitions for the new parent targets or an explicit mapping to the legacy parents are still required. → A: Use the spec's own parent keys; users.username and books.barcode were wrong names. Mapping: every reference to users.username becomes user.email VARCHAR(100), and every reference to books.barcode becomes product.pid VARCHAR(30). Corrected foreign keys: usercart: email VARCHAR(100) NOT NULL FK user.email ON DELETE CASCADE; pid VARCHAR(30) NOT NULL FK product.pid ON DELETE RESTRICT; UNIQUE (email, pid). user_demand: email VARCHAR(100) NOT NULL FK user.email ON DELETE CASCADE. transactions: email VARCHAR(100) NOT NULL FK user.email ON DELETE RESTRICT; pid VARCHAR(30) NOT NULL FK product.pid ON DELETE RESTRICT. All other columns, types, defaults and keys stay as given in the previous answer. The legacy username and barcode columns are not carried over.

- Q: How must acceptance-first sequencing cover the Go validation harness as well as SQL migrations? → A: Create only callable connection/migration scaffolding first. Write and demonstrate failing configuration/environment refusal cases before implementing those checks; separately demonstrate failing occupied-bootstrap and incompatible-credential-entry cases before implementing those preflights. Retain the existing SQL red gates and record each refusal's red/green evidence without allowing invalid inputs to reach migration mutation. This follows E1 Principle IV and Development Workflow rule 3; it changes sequencing, not schema behavior. (chosen by the fix run)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Establish a complete store schema (Priority: P1)

As the migration operator, I can establish the store's complete schema in an isolated validation environment without relying on the truncated seed file, so later Go work has a dependable persistence contract.

**Why this priority**: The supplied dump ends during a product-image seed insertion; an incomplete foundation prevents independent validation of every later persistence feature.

**Independent Test**: Use an empty isolated destination to apply the future schema artifact through the constitution's migration mechanism. Compare the resulting object definitions with the evidence register, without starting an application or importing legacy records.

**Acceptance Scenarios**:

1. **Given** an empty isolated destination and reconciled parent references, **When** the schema is established, **Then** all in-scope relations and their complete definitions exist, all tables use InnoDB and utf8mb4, execution ends without syntax or dependency errors, and there are zero business records.
2. **Given** an isolated destination containing synthetic rows and references from this schema, **When** its documented removal/recreation lifecycle is exercised, **Then** dependent objects can be removed and recreated without referential errors, and the recreated definitions match the initial definitions. Removal discards the synthetic rows; it promises no data preservation.
3. **Given** the supplied file's unfinished product-image seed statement, **When** the completed schema artifact is reviewed and applied, **Then** no unfinished statement or original seed payload is needed for successful setup.
4. **Given** an unrelated object outside this feature's declared inventory, **When** this feature's removal/recreation lifecycle runs, **Then** that object's definition and data remain unchanged.
5. **Given** pre-existing synthetic user accounts in the isolated legacy schema, **When** Q2's separate credential migration runs without reconstructing all tables, **Then** it drops the legacy password column, adds password_hash VARCHAR(255) initially nullable and must_reset_password BOOLEAN NOT NULL DEFAULT TRUE, retains account/profile records, and leaves every existing account with password_hash NULL and must_reset_password TRUE. No old password is copied, hashed or otherwise converted by the migration.

**Traceability**: E1 migration constraints; E2 `mysql_database`; E5 item 001; E8 Q1 table defaults and Q2's explicit credential-migration scope exception.

---

### User Story 2 - Preserve field values and documented defaults (Priority: P1)

As a maintainer of the rebuilt store, I can rely on the same documented required fields, optional fields, sizes and money values, so records are neither silently truncated nor changed into different values.

**Why this priority**: Field compatibility is necessary before the Java behavior can be rebuilt in Go.

**Independent Test**: Establish the schema and use synthetic records to check each confirmed field's maximum length, omitted value, explicit null, numeric type and default. Test each record family independently of application services.

**Acceptance Scenarios**:

1. **Given** a product with only a valid unique identity, **When** it is stored with optional fields omitted, **Then** its name, type, description, stock and image are null and its price is 0.00; **When** each optional field is explicitly null, **Then** null is accepted, but explicit null price is rejected under Q1's monetary replacement contract.
2. **Given** product identity/name/type/description values of 30/100/20/350 characters, **When** they are stored, **Then** they are retained unchanged; **When** any is one character longer, **Then** the validation case fails without a silently shortened value. Missing/null product identity is rejected.
3. **Given** a user with only a valid unique email, **When** optional name/mobile/address/postal fields are omitted or explicitly null, **Then** they remain null. Email/name/address accept 100/30/250 characters and fail at 101/31/251 without silent shortening; missing/null email is rejected. In Q2's initial credential state, omitted password_hash is null and omitted must_reset_password is TRUE; explicit null must_reset_password is rejected.
4. **Given** a line referencing an existing product with valid order/product identities, **When** quantity, amount and shipping are omitted, **Then** quantity is null, amount is 0.00 and shipping is zero; **When** shipping or amount is explicitly null, **Then** the write is rejected. Explicit null quantity is accepted.
5. **Given** distinct valid identities and all required parents, **When** product prices, line amounts and transaction amounts of 0.00, 12.34, 99999999.99 and -99999999.99 are stored, **Then** their exact values are retained; magnitudes of 100000000.00 fail rather than being clipped. Every monetary field rejects explicit null and defaults to 0.00 when omitted.
6. **Given** valid identities and no other violated constraint, **When** zero or negative values are supplied to the confirmed signed stock and line-quantity fields, **Then** no newly invented positive-only rule rejects them. Optional numeric fields remain distinct from zero.
7. **Given** a transaction_id of 40 characters and valid required parent references and quantity, **When** it is stored, **Then** it is retained and an omitted id receives a unique BIGINT primary key; **When** transaction_id is null, omitted, duplicated or 41 characters, **Then** the relevant required/unique/length case fails. The legacy transid definition is superseded.
8. **Given** Q1's verified cart, demand and transaction definitions in the field register, **When** each field's omission, explicit null and boundaries are exercised, **Then** all required values and limits are enforced. Email/pid/book_name/author accept 100/30/100/100 characters and reject one more; author accepts null. Cart quantity defaults to 1, cart and transaction quantities accept 1 and reject null, zero and negative values, and transaction quantity rejects omission. Omitted ids are generated uniquely; omitted timestamps use CURRENT_TIMESTAMP, explicit valid timestamps are retained, and explicit null timestamps are rejected. These cases use the operator's supplied definitions, not legacy bean assumptions.
9. **Given** otherwise valid records, **When** stock, line quantity, postal code and shipping are set to -2147483648 and 2147483647, and mobile to -9223372036854775808 and 9223372036854775807, **Then** those signed integer boundary values are retained; values one beyond either bound fail without clipping. Shipping values such as 2 remain valid integers rather than being rejected by an invented binary-only rule.
10. **Given** valid parents, **When** cart or transaction quantity is 2147483647, **Then** it is retained; 2147483648 fails without clipping. Explicit duplicate BIGINT ids fail in each Q1 table, and otherwise valid signed BIGINT boundary ids are retained without imposing an unsourced positive-only id rule.
11. **Given** the migrated user schema, **When** a full synthetic encoded Argon2id string containing algorithm, parameters, salt and hash is stored in password_hash, **Then** it is retained exactly, including values up to 255 characters; a 256-character value fails without truncation. The password column is absent. SQL storage capacity is tested here; validating submitted passwords and writing hashes belong to authentication.
12. **Given** any account still awaiting reset with password_hash NULL or must_reset_password TRUE, **When** final credential nullability is considered, **Then** the NOT NULL transition is not applied; **Given** every account has reset, every password_hash is non-null and no reset flag remains TRUE, **When** the final versioned schema transition runs, **Then** password_hash becomes NOT NULL, existing hashes remain unchanged and subsequent null hash writes fail. Authentication must require reset at next sign-in and write hashes; that code is outside this feature.

**Traceability**: E3 `mysql_database` FR-001–023 and FR-027–040; E4 matching rules; E8 Q1 overrides monetary and transaction definitions and supplies cart/demand fields; E8 R1 supplies the corrected identity columns and parent-key lengths; E1 III and E8 Q2 supply the protected credential contract and staged nullability.

---

### User Story 3 - Protect identities and historical order references (Priority: P1)

As a store operator, I can trust that each record has its documented identity and that product maintenance cannot break existing order lines.

**Why this priority**: Historical purchases must continue to point to real products, and duplicate identities must not replace existing records.

**Independent Test**: Use synthetic products and lines to exercise uniqueness, missing-parent writes, referenced-parent changes and the corresponding successful operations after the last reference is removed. No checkout or product service is required.

**Acceptance Scenarios**:

1. **Given** an existing product, user or transaction identity, **When** another record with the identical identity is inserted, **Then** the duplicate is rejected and the original is unchanged.
2. **Given** an existing order/product pair, **When** another line with the same pair is inserted, **Then** it is rejected; **When** a different product is added to that order or the same product to another order, **Then** the composite identity permits it, provided all other established constraints are met.
3. **Given** no product with a specified identity, **When** a line is inserted or updated to reference that identity, **Then** it is rejected and no orphan line is stored.
4. **Given** a product referenced by one or more order lines, **When** its deletion is attempted, **Then** deletion fails and the product and every referencing line remain unchanged.
5. **Given** a product referenced by one or more order lines, **When** its identity is changed, **Then** the change fails and neither product nor line identities are rewritten.
6. **Given** several lines referencing a product, **When** all but one reference are removed, **Then** product deletion and identity change remain blocked; **When** the final reference is removed, **Then** those two operations are permitted by the order-line relationship, subject to Q1's other references. Verify each operation on a separate fixture.
7. **Given** a referenced product, **When** an optional non-identity field is changed to another valid value, **Then** the product-reference rule does not block the change and the line still references the same product.
8. **Given** the recreated schema from story 1, **When** these reference and identity cases are repeated, **Then** their results are identical; setup must not leave constraint enforcement disabled.
9. **Given** the corrected user.email and product.pid parents under R1, **When** cart, demand or transaction rows are inserted or updated with a missing referenced email or pid, **Then** the write fails. Duplicate cart (email, pid) pairs fail, while distinct pairs succeed; no unsupplied demand uniqueness is imposed.
10. **Given** a user referenced only by cart and demand rows, **When** it is deleted, **Then** those rows cascade-delete; **Given** a transaction referencing the user or a cart/transaction referencing a product, **When** the corresponding parent is deleted, **Then** deletion is restricted with all rows unchanged. Use separate fixtures for cascade and restriction cases.
11. **Given** existing Q1 parent/child references, **When** a referenced user.email or product.pid is updated to a valid unused identity, **Then** the corresponding child references cascade-update. This default does not override the explicitly retained NO ACTION relationship from order lines to product identities.

**Traceability**: E3 FR-003, FR-018, FR-024–029 and FR-038–040; E4 `mysql_database`; E8 Q1 replacement identities, uniqueness and referential actions, corrected by E8 R1; explicit operator requirement.

### Edge Cases

- Omitting an optional field and explicitly storing null must both preserve documented null semantics; neither is equivalent to zero or an empty string.
- Required identity means non-null, not an invented non-empty, email-format or normalization rule. Case sensitivity and whitespace equivalence are UNVERIFIED environment semantics; duplicate cases use exactly identical values.
- Every confirmed character limit needs an exact-limit and one-over-limit case, including order identities at 45/46 characters, product identities at 30/31, and user/child emails at 100/101. orders.prodid retains VARCHAR(45), but a valid reference must match an existing product.pid of at most 30 characters; 45 characters cannot be a valid parent identity. Required line identities reject null and omission.
- Decimal precision counts all digits, including the two fractional digits. Q1 replaces the differing legacy monetary definitions with DECIMAL(10,2) NOT NULL DEFAULT 0.00 throughout. Fractional inputs beyond two places require the supported environment's behavior to be documented during planning; no unverified rounding policy is claimed.
- Shipping is a non-null integer with default zero. Evidence does not restrict stored values to zero and one; the schema must not invent that restriction from a later workflow's state transitions.
- Optional image absence is null, not a required placeholder blob. Validate a small synthetic binary value can be retained without reproducing the unfinished image seed.
- Nullable stock/line quantities are signed integers; mobile is a nullable signed large integer and postal code a nullable signed integer. Do not invent telephone, postal-format or positive-value checks. Validation must include type/range boundaries for the supported environment.
- Changing an existing line to a missing product must fail just as insertion does. A failed parent deletion or identity update must not cascade, null references, or remove existing lines.
- Removing some references is insufficient to unblock a product. Removing the last reference only removes this particular restriction; Q1's other declared references still apply.
- A query joining an order identity to a transaction identity does not establish a foreign key. A service checking whether a cart/demand pair exists does not establish a unique constraint.
- Missing or truncated persistence evidence cannot be replaced by a conventional shopping-cart schema or by Java bean defaults. Q1 is an explicit operator replacement contract and verifies its supplied fields even though no legacy document defines them.
- An empty destination, a populated synthetic destination, and a remove/recreate cycle must be handled without touching unrelated objects. Reconstruction is destructive. Q2 additionally requires a separate credential-column migration that retains existing accounts and initializes reset state; it does not authorize a general production-data upgrade or plaintext conversion.
- password_hash is initially nullable to represent an account awaiting reset, not a migrated plaintext password. Its full Argon2id encoding must fit without truncation. The later NOT NULL transition waits for all resets; it must not be applied with the initial schema migration while reset-pending accounts exist. Migration rollback must not restore plaintext credentials.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The feature MUST specify a complete executable MySQL schema artifact for `mysql_database`, serving only the declared Java-to-Go rebuild. Its namespace MUST preserve the evidenced `shopping-cart` name unless an explicit operator decision changes it. Acceptance: US1.1–3; E2/E3/E5 and operator input.
- **FR-002**: The artifact MUST support creating the in-scope objects and a documented removal/recreation lifecycle in dependency-safe order, including an initially empty destination. It MUST finish all statements and MUST leave constraints enforced. It MUST affect only declared in-scope objects. Acceptance: US1.1–4 and US3.8; E2/E5, with isolation as a validation assumption.
- **FR-003**: The feature MUST cover product, user, cart, order-line/order, transaction and demand persistence. The six legacy physical relations are `product`, `user`, `usercart`, `orders`, `transactions` and `user_demand`; Q1 supplies verified replacement contracts for cart, transactions and demand; R1 corrects their parent references to `user.email VARCHAR(100)` and `product.pid VARCHAR(30)`. The parent keys MUST use those corrected lengths; all child account references MUST use email VARCHAR(100), and cart/transaction product references MUST use pid VARCHAR(30). The erroneous username/barcode columns and users/books parent names MUST NOT be carried into the target schema. The documented `orders` relation contains order lines; no separate order-header or `order_lines` relation is evidenced. All tables MUST use InnoDB and utf8mb4. Acceptance: US1.1, US2.8/10 and US3.9–11; E2–E5, E8 Q1 and R1.
- **FR-004**: Product fields MUST use required unique primary identity product.pid VARCHAR(30) under R1 and preserve optional name/type/description/stock/image with their documented sizes, types and null defaults. Price MUST use Q1's monetary replacement in FR-010. Acceptance: US2.1–2, US2.5–6, US3.1 and image edge case; E3 FR-001–013, E8 Q1 and R1.
- **FR-005**: User identity MUST be required primary email VARCHAR(100) under R1. Optional name/mobile/address/postal fields MUST preserve their confirmed sizes, types and null defaults. Q2 MUST replace the legacy nullable password column with password_hash VARCHAR(255), initially NULL, storing the complete Argon2id algorithm/parameters/salt/hash string, and must_reset_password BOOLEAN NOT NULL DEFAULT TRUE. The credential migration MUST retain existing accounts, set their password_hash to NULL and require reset at next sign-in, with no plaintext stored or migrated. Once every user has reset, a versioned transition MUST make password_hash NOT NULL; it MUST NOT run while reset-pending accounts remain. Columns and migrations are in scope; sign-in, password reset handling and hash-writing code belong to authentication. Acceptance: US1.5, US2.3/11–12 and US3.1; E3 FR-027–037, E1 III, E8 Q2 and R1.
- **FR-006**: Order lines MUST retain required orders.orderid VARCHAR(45) and orders.prodid VARCHAR(45), with a composite primary identity allowing one line per order/product pair. The product reference MUST match an existing product.pid VARCHAR(30) under R1. Quantity MUST be nullable with default null; amount MUST follow FR-010; shipping MUST be a non-null integer defaulting to zero. Acceptance: US2.4–6, line-identity edge case and US3.2; E3 FR-014–023, E8 Q1 monetary replacement and R1 parent length.
- **FR-007**: Each order-line product reference MUST identify an existing product. The relationship named `productid` MUST preserve `ON DELETE NO ACTION` and `ON UPDATE NO ACTION`: reject orphan writes and reject deletion or identity change of referenced products without altering existing rows. Acceptance: US3.3–8; E3 FR-024–026.
- **FR-008**: Transactions MUST use Q1's replacement contract: id BIGINT primary key AUTO_INCREMENT; transaction_id VARCHAR(40) NOT NULL UNIQUE; email VARCHAR(100) NOT NULL referencing user.email; pid VARCHAR(30) NOT NULL referencing product.pid; quantity INT NOT NULL with minimum 1 and no default; amount per FR-010; created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP. The legacy transid VARCHAR(45) primary key is superseded, not retained as an extra requirement. Acceptance: US2.5/7–8/10 and US3.1/9–11; E8 Q1 and R1, replacing E3 FR-038–040.
- **FR-009**: Cart and demand MUST implement every Q1 field and key in the field register. Cart quantity MUST default to 1 and both cart and transaction quantities MUST be at least 1. Cart MUST enforce UNIQUE (email, pid). Cart/demand/transaction email fields MUST be VARCHAR(100) NOT NULL referencing user.email; cart/transaction pid fields MUST be VARCHAR(30) NOT NULL referencing product.pid. All Q1 foreign keys MUST use ON UPDATE CASCADE and ON DELETE RESTRICT, except cart/demand email references MUST use ON DELETE CASCADE. The explicitly stated order-line NO ACTION rules in FR-007 remain exceptions. No additional demand uniqueness or transaction-to-line reference is authorized. Acceptance: US2.8/10 and US3.9–11; E8 Q1 and R1.
- **FR-010**: Every monetary field, including product.pprice, orders.amount and transactions.amount, MUST use DECIMAL(10,2) NOT NULL DEFAULT 0.00 and preserve exact in-range two-place values without floating-point substitution or silent clipping. Q1 explicitly replaces the legacy nullable monetary contracts and product precision 12. Acceptance: US2.1/4–5 and money edge case; E8 Q1 supersedes E3 FR-010–011/020–021.
- **FR-011**: Omitted values, explicit nulls, required identities and literal defaults MUST be validated separately. Confirmed optional values MUST NOT be replaced with empty strings, zeros, a current timestamp or required image data. Acceptance: US2.1/3–4 and edge cases; E3 confirmed field definitions.
- **FR-012**: The schema MUST NOT add unsourced business validation such as positive-only stock/order-line quantities/prices, binary-only shipping, email-format checks or normalization. Q1 explicitly authorizes generated BIGINT identities for its three tables and minimum-1 cart/transaction quantities; these MUST be enforced. Environment choices affecting accepted values MUST be documented without being described as observed legacy behavior. Acceptance: US2.5–6/8/10 and boundary/default edge cases; E1 evidence-based migration, E3 definitions, E6, E8 Q1.
- **FR-013**: No general legacy data import, plaintext credential conversion, business seed dataset or unfinished seed statement MUST be included. Q2's specific credential-column migration and initialization of existing accounts to password_hash NULL and must_reset_password TRUE are in scope. Constraint-validation records MUST be minimal, synthetic, isolated from the schema artifact's empty result, and removable after validation. Acceptance: US1.1–3/5, US2.11–12 and fixture review; operator scope, E8 Q2 and E1 IV/migration constraints.
- **FR-014**: Acceptance cases MUST be written and shown to fail before the corresponding implementation, including Go configuration/environment refusals and migration-entry preflights as well as SQL behavior, then pass against the actual schema or the corresponding harness checks. Connection/migration scaffolding MUST NOT implement a refusal before its own acceptance cases demonstrate the missing behavior. Configuration/environment checks MUST turn green before domain red gates; occupied-bootstrap and incompatible-credential-entry preflights MUST wait for their lifecycle/upgrade tests and red evidence. Refusal tests MUST prove rejection before invoking migration mutation, without relying on a missing source file or later SQL error. Validation MUST inspect object definitions as well as exercise successful and rejected writes, including unchanged rows after reference failures. An unresolved or unexecuted case MUST NOT be reported as passing. Acceptance: US1–3 and the outcome accounting below; E1 IV and Development Workflow rule 3.
- **FR-015**: The implementation phase MUST deliver database changes as versioned up/down SQL migrations in Git using `golang-migrate`, as already chosen by the operator in E1. No database application occurs during this specification phase. Acceptance: future delivery review and US1 migration lifecycle; E1 migration decision.
- **FR-016**: The feature MUST record E8's operator-directed replacements and R1's resolved parent mapping without introducing users/books parents or username/barcode columns. It MUST include Q2's credential columns and versioned migrations but MUST NOT implement sign-in/hash-writing code, later services, role/membership/session schemas, notification/retry infrastructure or other application behavior. Acceptance: inventory/delta review, US1.5 and US2.11–12; operator item 001 boundary, E1 governance, E5, E8 Q2.

### Key Entities

- **Product**: A catalog item, identified by product identity, with optional descriptive, price, stock and binary-image values. A historical order line protects its identity and existence.
- **User**: A store account with optional contact/profile values. Identity is user.email VARCHAR(100), as corrected by R1. Q2 replaces password with an initially nullable 255-character encoded Argon2id hash and a required reset flag defaulting TRUE; hashes become required after every account resets.
- **Cart entry**: A Q1 row with a generated BIGINT primary key, required email and pid references, positive quantity defaulting to 1, added_at defaulting to CURRENT_TIMESTAMP and a unique (email, pid) pair.
- **Order line**: One product within an order, identified by the order/product pair, with optional quantity/amount and required shipping status. The `orders` relation stores these lines.
- **Transaction**: A Q1 checkout/payment record identified by a generated BIGINT primary key and a unique transaction_id of at most 40 characters, with required email, pid, positive quantity, fixed-precision amount and created_at timestamp.
- **Demand entry**: A Q1 request identified by a generated BIGINT key, with required email and book_name, nullable author and required requested_at timestamp. This replaces the legacy product/quantity-shaped concept; no quantity column or pair uniqueness is added.

### Evidence and scope register

Only permitted documents and existing local artifacts are evidence. E3/E4 reproduce source-line excerpts; their documents are labelled `model_asserted`, and E2 reports zero execution-grounded inventory items. This specification does not upgrade those descriptions into observed runtime results.

| Ref | Source | Use in this feature |
| --- | --- | --- |
| E1 | [Constitution v1.0.1](../../.specify/memory/constitution.md) | Declared Java-to-Go direction, source-only inventory, protected credentials, versioned migrations, acceptance-first validation, scope and acceptance authority. |
| E2 | [Legacy overview](../../../../old/docs/overview.md#mysql_database) | Schema reconstruction, existing-table drops, unfinished image-bearing seed and evidence strength. |
| E3 | [Legacy product requirements](../../../../old/docs/prd.md#mysql_database-40) | All 40 schema rules and their SQL excerpts; selected service query excerpts identify relations/columns. |
| E4 | [Legacy business requirements](../../../../old/docs/brd.md#mysql_database-40), [low-level design](../../../../old/docs/lld.md#mysql_database) | Corroborating schema rules; cart/demand/order/transaction query and bean descriptions. |
| E5 | [Roadmap item 001](../../../../old/docs/roadmap.md#001--complete-mysql-schema) | Exactly this feature, component and exclusions. Later roadmap items are outside scope. |
| E6 | [Data-model limitation](../../../../old/docs/data_model.md), [migration approach](../../../../old/docs/approach.md) | Persistence inventory unavailable, incomplete DDL and open authoritative-schema question. Advisory choices are not operator decisions. |
| E7 | [Earlier rebuild spec](../001-java-to-go/spec.md) | Prior persistence gap O2 and named protected-account difference D2; no complete missing schema or operator credential-layout resolution is recorded there. |
| E8 | Operator answers recorded in Clarifications, Session 2026-11-10 | Authoritative replacement contracts for missing tables, table engine/character set, monetary fields and default referential actions (Q1); credential fields, reset initialization, staged NOT NULL migration and authentication scope boundary (Q2). R1 corrects the parent keys and child names/types to user.email VARCHAR(100) and product.pid VARCHAR(30), retaining the other supplied definitions. These are explicit operator evidence, not observed legacy runtime behavior. |

#### Confirmed field contract

These definitions are requirements-level compatibility data expressly requested by the operator, not an implementation proposal. `NULL / NULL` means nullable with default null; `required` means non-null. A blank/unverified attribute must not be inferred as a default.

| Entity / field | Confirmed storage contract | Nullability / default | Evidence |
| --- | --- | --- | --- |
| product.pid | VARCHAR(30); primary key | required; no generated value evidenced | E8 R1 replaces E3 FR-001–003 length |
| product.pname | VARCHAR(100) | NULL / NULL | E3 FR-004–005 |
| product.ptype | VARCHAR(20) | NULL / NULL | E3 FR-006–007 |
| product.pinfo | VARCHAR(350) | NULL / NULL | E3 FR-008–009 |
| product.pprice | DECIMAL(10,2) | required / 0.00 | E8 Q1 replaces E3 FR-010–011 |
| product.pquantity | INT | NULL / NULL | E3 FR-012 |
| product.image | LONGBLOB | NULL / NULL | E3 FR-013 |
| orders.orderid | VARCHAR(45); first part of primary key | required | E3 FR-014–015/018; order-query naming below |
| orders.prodid | VARCHAR(45); second part of primary key; productid reference to product.pid | required; NO ACTION on parent delete and update | E3 FR-016–018/024–026 |
| orders.quantity | INT | NULL / NULL | E3 FR-019 |
| orders.amount | DECIMAL(10,2) | required / 0.00 | E8 Q1 replaces E3 FR-020–021 |
| orders.shipped | INT | required / 0 | E3 FR-022–023 |
| user.email | VARCHAR(100); primary key | required | E8 R1 replaces E3 FR-027–029 length; UserServiceImpl FR-133 confirms table name |
| user.name | VARCHAR(30) | NULL / NULL | E3 FR-030–031 |
| user.mobile | BIGINT | NULL / NULL | E3 FR-032 |
| user.address | VARCHAR(250) | NULL / NULL | E3 FR-033–034 |
| user.pincode | INT | NULL / NULL | E3 FR-035 |
| legacy user.password | VARCHAR(20); dropped by Q2, not a target field | NULL / NULL in legacy only | E3 FR-036–037; E1 III and E8 Q2 replace plaintext storage |
| user credential: password_hash | VARCHAR(255); complete encoded Argon2id string | initially nullable / NULL for existing users; NOT NULL only after all users reset | E8 Q2; R1 confirms user.email as the account parent |
| user credential: must_reset_password | BOOLEAN | required / TRUE; existing users initialized TRUE | E8 Q2 |
| usercart.id | BIGINT; primary key AUTO_INCREMENT | required / generated | E8 Q1 |
| usercart.email | VARCHAR(100); FK user.email; ON UPDATE CASCADE, ON DELETE CASCADE | required; no default supplied | E8 Q1 and R1 |
| usercart.pid | VARCHAR(30); FK product.pid; ON UPDATE CASCADE, ON DELETE RESTRICT | required; no default supplied | E8 Q1 and R1 |
| usercart.quantity | INT; at least 1 | required / 1 | E8 Q1 |
| usercart.added_at | DATETIME | required / CURRENT_TIMESTAMP | E8 Q1 |
| user_demand.id | BIGINT; primary key AUTO_INCREMENT | required / generated | E8 Q1 |
| user_demand.email | VARCHAR(100); FK user.email; ON UPDATE CASCADE, ON DELETE CASCADE | required; no default supplied | E8 Q1 and R1 |
| user_demand.book_name | VARCHAR(100) | required; no default supplied | E8 Q1 |
| user_demand.author | VARCHAR(100) | nullable; no explicit default supplied | E8 Q1 |
| user_demand.requested_at | DATETIME | required / CURRENT_TIMESTAMP | E8 Q1 |
| transactions.id | BIGINT; primary key AUTO_INCREMENT | required / generated | E8 Q1 |
| transactions.transaction_id | VARCHAR(40); UNIQUE | required; no default supplied | E8 Q1 replaces legacy transid VARCHAR(45) primary key, E3 FR-038–040 |
| transactions.email | VARCHAR(100); FK user.email; ON UPDATE CASCADE, ON DELETE RESTRICT | required; no default supplied | E8 Q1 and R1 |
| transactions.pid | VARCHAR(30); FK product.pid; ON UPDATE CASCADE, ON DELETE RESTRICT | required; no default supplied | E8 Q1 and R1 |
| transactions.quantity | INT; at least 1 | required; no default supplied | E8 Q1 |
| transactions.amount | DECIMAL(10,2) | required / 0.00 from Q1's general money rule | E8 Q1 |
| transactions.created_at | DATETIME | required / CURRENT_TIMESTAMP | E8 Q1 |

The table accounts for all 40 legacy schema rules by preservation or a named operator-directed replacement: FR-001–013 product; FR-014–026 lines and product reference; FR-027–037 user; FR-038–040 transaction identity. E4 repeats those rules rather than adding missing table definitions. E8 Q1 supplies the missing contracts and overrides the monetary and transaction-identity rules; E8 R1 corrects parent names, parent-key lengths and child identity columns. Cart additionally has UNIQUE (email, pid); no additional demand unique key is specified.

#### Relation inventory and unverified attributes

| Requested record family | Evidenced relation / fields | Confirmed limits of evidence |
| --- | --- | --- |
| Product | Legacy `shopping-cart.product`; fields above | Q1 mandates InnoDB/utf8mb4 and the monetary replacement. R1 confirms product.pid VARCHAR(30) as the parent of cart/transaction pid references. Collation remains a planning choice. |
| Users | Legacy `user`; profile and Q2 credential fields above | E3 UserServiceImpl FR-133/136 explicitly query user.email. R1 confirms user.email VARCHAR(100) as the parent of all child email references. Q2 verifies the replacement credential layout and migration. |
| Cart | Q1/R1 `usercart`: id, email, pid, quantity, added_at | Verified replacement of E3's legacy username/prodid/quantity shape; all fields, uniqueness and foreign-key actions are supplied by E8 Q1 with R1 correcting the parents and child columns. |
| Orders / order lines | `orders`; five fields above | E3 OrderServiceImpl FR-096 joins `orders o` on `o.orderid = t.transid`; FR-097 names quantity/amount/shipped. No separate order-header relation or order-to-transaction foreign key is established. |
| Transactions | Q1/R1 `transactions`: id, transaction_id, email, pid, quantity, amount, created_at | Verified E8 replacement; legacy transid/time names and the legacy 45-character primary transaction identity do not remain target requirements. No order-to-transaction FK is supplied. |
| Demand | Q1/R1 `user_demand`: id, email, book_name, author, requested_at | Verified E8 replacement; legacy prodid/quantity are not target fields. No additional demand uniqueness is supplied. |

O1 is resolved by the operator's verified replacement definitions; their attributes no longer remain UNVERIFIED. Equality joins and service existence checks still do not establish additional references or uniqueness. Q1's cascades are explicitly authorized. R1 explicitly corrects the erroneous parent names to user.email and product.pid; no additional parent tables or username/barcode columns remain in the target contract.

### Named differences and inherited decisions

| ID | Legacy/advisory behavior | Required disposition and validation consequence |
| --- | --- | --- |
| D1 — Complete the truncated schema artifact | E2 reports the supplied file ends during an image seed insertion. | Explicit operator scope replaces that incomplete artifact with complete executable DDL and no business seeds. US1.1–3 verifies setup without seed payloads. This is artifact repair, not data migration. |
| D3 — Operator-supplied table contracts | Legacy cart/demand SQL uses prodid and product/quantity demand; legacy transactions use transid VARCHAR(45) as primary key and time. Missing table attributes were previously UNVERIFIED. | E8 Q1 supplies verified replacement definitions, generated ids, cart pair uniqueness, positive cart/transaction quantities, timestamps and explicit user.email/product.pid references corrected by R1. US2.7–8/10 and US3.9–11 verify the replacements. Later Go services must consume these contracts; their implementation stays outside item 001. |
| D5 — Corrected parent identities | E3/E4 define user.email VARCHAR(60) and product.pid VARCHAR(45); the initial Q1 answer used erroneous users.username/books.barcode targets. | E8 R1 explicitly selects user.email VARCHAR(100) and product.pid VARCHAR(30), with matching email/pid child columns and cart UNIQUE (email, pid). No username/barcode columns or additional parent tables are retained. US2.2–3/8 and US3.9–11 verify corrected lengths, uniqueness and references. orders.prodid remains VARCHAR(45), and its NO ACTION rules still block referenced product deletion or identity changes. |
| D4 — Monetary and table defaults | E3 has nullable product price DECIMAL(12,2), nullable line amount DECIMAL(10,2), and no verified engine/charset. | E8 Q1 requires every monetary field DECIMAL(10,2) NOT NULL DEFAULT 0.00 and every table InnoDB/utf8mb4. US1.1 and US2.1/4–5 verify the new contracts, including reduced product-price capacity. Explicit order-line NO ACTION rules remain exceptions to the new default FK actions. |
| D2 — Protected credential storage | E3 FR-036–037 evidences a nullable password of at most 20 characters; E2 UserServiceImpl/E5 items 007 and 016 describe storing supplied plaintext. | E1 III and E8 Q2 require dropping password, adding password_hash VARCHAR(255) for the full Argon2id encoding and must_reset_password BOOLEAN NOT NULL DEFAULT TRUE. Existing users receive NULL hashes and must reset at next sign-in; after all resets, a versioned transition makes hashes NOT NULL. No plaintext is copied, converted, stored or restored. US1.5 and US2.3/11–12 verify field capacity, reset initialization and staged nullability. The column migrations are in item 001; sign-in/hash-writing remains in authentication. |

E1 also resolves atomic checkout, collision-safe identity and notification retry against advisory legacy choices. Those application changes belong to later features; they neither authorize extra tables here nor change the confirmed identity lengths in this feature. No permitted document read contradicts the declared Java-to-Go language direction. E7's broad scope is superseded for this invocation by the operator's explicit item 001 boundary.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001 — Repeatable foundation**: In an isolated validation environment, 100% of the resolved in-scope record families are available after initial setup and after a removal/recreation cycle, with zero unfinished statements, zero required business seed records and zero changes to unrelated objects. The operator can establish the foundation without repairing the artifact manually.
- **SC-002 — Complete constraint accounting**: All 40 source-characterized schema rules have an explicit preservation or operator-directed difference, and every remaining field/relationship has a cited definition or a visible UNVERIFIED item. Complete conformance requires zero unresolved schema-definition or credential-disposition items; an unknown does not count as verified.
- **SC-003 — Value fidelity**: 100% of the defined omission, explicit-null, required-value, exact-length, one-over-length, numeric-range and two-place-money cases yield their specified outcomes. Successful records retain their exact expected values, with zero silent shortening or clipping in the validation suite.
- **SC-004 — Historical integrity**: Across every defined missing-product, duplicate-pair, referenced-product deletion and identity-change case, zero orphan lines, unintended deletions or rewritten identities occur. Valid distinct pairs and ordinary product-field changes remain possible.
- **SC-005 — Reviewable handoff**: A reviewer can trace every in-scope field and acceptance case to supplied evidence or a recorded operator decision, and can distinguish confirmed behavior from every remaining gap without inspecting application code. All three operator/maintainer journeys have documented validation results before implementation completion is claimed.

These are proposed validation targets, not measured results or platform acceptance. No evidenced response-time, traffic or production-size target exists; the specification does not invent one for a schema-only deliverable.

## Assumptions

- Legacy documentation is read-only and was entered through its README. Only that reading, explicit operator requirements and earlier local artifacts supply estate evidence. No surrounding application code or external schema is used.
- This invocation writes this specification and its quality checklist. Executable DDL, acceptance-test implementation, plan and tasks belong to their later phases. No database has been changed and no import, push, merge or deployment is performed.
- Validation uses a disposable, isolated destination and synthetic records. Removal/recreation intentionally loses its in-scope records; up/down availability does not promise data recovery. The separate Q2 credential migration retains existing accounts and initializes reset state; broader data migration remains outside scope.
- Q1 establishes InnoDB and utf8mb4 for all tables as target requirements. The supported MySQL version, collation and validation mode remain planning dependencies, not established legacy facts. Planning must name the environment and require rejection of overlength/out-of-range validation inputs rather than silent shortening or clipping. Case folding, whitespace equivalence and excess-fraction rounding are not asserted as legacy rules.
- The terms orders and order lines refer to the evidenced orders line relation. R1 confirms the six physical relations and explicitly corrects the parent targets to user.email and product.pid; users/books parents and username/barcode target columns are excluded. No separate order header is assumed.
- E1 governs migration packaging and acceptance-first synthetic testing. Since this feature exposes no endpoint or page, endpoint role matrices, browser coverage and application authorization implementations remain with later features, without waiving their constitutional requirements.
- Protected credentials are mandatory under E1 and their columns/migration are verified by Q2. No sign-in/hash-writing, plaintext conversion, membership/session/audit or retry infrastructure is included. Authentication consumes the reset flag and supplies hashes; this schema feature supplies the initial and post-reset nullability contracts.
- No operator acceptance boundaries beyond the supplied requirements are stated. Local document checks or future test passes do not decide acceptance; that remains with the referee/platform gates.

### Open evidence and clarification items

**O1 — Complete persisted contract**: RESOLVED by E8 Q1. The operator supplies the verified usercart, user_demand and transactions contracts, global money/default referential rules and InnoDB/utf8mb4 requirements. The feature is not a draft; missing legacy documents do not invalidate this explicit operator evidence.

**R1 — Parent-reference reconciliation**: RESOLVED by the operator's recorded correction in E8. The parent keys are user.email VARCHAR(100) and product.pid VARCHAR(30), replacing the legacy parent-key lengths of 60 and 45 respectively. usercart uses required email/pid references and UNIQUE (email, pid); user_demand uses required email; transactions uses required email/pid. Every child reference supplied by Q1 uses ON UPDATE CASCADE; cart/demand email uses ON DELETE CASCADE, and transaction email plus cart/transaction pid use ON DELETE RESTRICT. No users/books parents or username/barcode columns are carried over. All other columns, types, defaults and keys remain as previously supplied, including orders.prodid VARCHAR(45) with its explicit NO ACTION rules.

**O2 — Credential field disposition**: RESOLVED by E8 Q2. Drop the legacy password column; add password_hash VARCHAR(255) for full encoded Argon2id strings, initially nullable, and must_reset_password BOOLEAN NOT NULL DEFAULT TRUE. Initialize existing users to NULL hash/reset required, never migrate plaintext, and make password_hash NOT NULL only after every user resets. Columns and migrations are in scope; sign-in/hash-writing code belongs to authentication.

### Changelog

- 2026-11-10 — Clarified FR-014 for finding C1: acceptance-first ordering covers Go configuration/environment and migration-entry refusals, with separate red/green evidence before their implementation. Recorded the fix-run sequencing choice for the java to go schema foundation; schema behavior and scope are unchanged.

- 2026-11-10 — Recorded the operator's R1 answer for work unit run-8fd0f0975fa64942-wu; resolved the parent-reference marker, corrected user.email/product.pid and child email/pid columns, updated cart uniqueness and identity-boundary/reference acceptance cases, and retained all other supplied definitions. This serves the java to go schema foundation; no implementation or platform acceptance is claimed.

- 2026-11-10 — Recorded E8 Q2 for work unit run-46c492e9970f4b94-wu; resolved O2, integrated the replacement credential columns and staged migration, added existing-account and encoded-hash acceptance cases, and narrowed the general no-data-migration exclusion to honor this explicit operator exception. Both supplied clarification questions are answered; application authentication remains outside item 001.

- 2026-11-10 — Recorded E8 Q1 for work unit run-46c492e9970f4b94-wu; resolved O1, integrated verified table contracts and acceptance cases, replaced monetary defaults, recorded deliberate differences, and retained the newly exposed parent-reference reconciliation without inventing aliases. The feature is no longer a draft; no implementation or platform acceptance is claimed.

- 2026-11-10 — Created specification 1 for work unit run-53683cca8ca24d32-wu using the resolved active template and constitution v1.0.1. Restricted scope to mysql_database, accounted for all 40 schema rules, added independent acceptance cases and outcomes, and retained two unresolved schema/credential clarifications. No implementation or platform acceptance claimed.

- 2026-11-10 — Revalidated specification 1 for work unit run-cacf435f7e794481-wu against constitution v1.0.1 and permitted evidence. Corrected the physical user-table name from UNVERIFIED to `user` using PRD UserServiceImpl FR-133/136 and LLD UserServiceImpl, explicitly listed the six evidenced relations, narrowed O1 to missing persisted definitions, and separated environment planning choices from missing estate facts. O1/O2 remain unresolved; Java-to-Go direction and mysql_database-only scope are retained.
