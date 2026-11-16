# Feature Specification: Electronics Store Java-to-Go Rebuild

**Feature Branch**: None created; no branch hook is registered and this working tree has no Git repository.

**Created**: 2026-11-10

**Status**: Draft — evidence clarifications outstanding; no implementation or platform acceptance claimed.

**Input**: Build the candidate for run run-4e3ac24f2fa04792. The operator declares source language java, target language go, and the goal, verbatim, "java to go". Follow the recorded operator decisions in constitution v1.0.1.

This feature specifies rebuilding the documented electronics store in Go: accounts, authenticated catalog and images, carts, unmet demand, demo checkout, inventory, order history, shipping, and notifications. This invocation creates one feature; the advisory roadmap's component split does not create additional features. Language direction is an explicit operator requirement; implementation design belongs to planning.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Use the store within granted membership (Priority: P1)

As an account holder, I can sign in, use information and actions permitted by my role and membership, and sign out without exposing another user's records.

**Why this priority**: All store journeys depend on authenticated, server-enforced access. The operator explicitly replaces legacy access bypasses.

**Independent Test**: Use synthetic accounts, grants and memberships to exercise sign-in, protected reads, denied writes, cross-project requests and sign-out without checkout or live mail.

**Acceptance Scenarios**:

1. **Given** a valid account with a granted role and membership, **When** it signs in, **Then** its session identifies the stored account and granted role; submitted role text cannot elevate permissions.
2. **Given** an anonymous request, **When** it requests a protected operation, including registration, catalog, images or contact, **Then** it fails before protected disclosure or business effects; only login and password reset may be anonymous.
3. **Given** each of admin, editor and viewer, **When** it requests an operation outside its permission or applicable workspace/project membership, **Then** it receives HTTP 403, no business mutation or mail occurs, and a workspace/project refusal audit carries the response/log request ID.
4. **Given** a legacy customer mapped to viewer, **When** it reads its permitted project, **Then** it can read; **When** it writes, **Then** it is refused. Only customer writes explicitly documented in E3 can be granted to editor, within membership; no customer becomes admin.
5. **Given** an authenticated user, **When** it signs out, **Then** username, role and profile access are cleared, the documented login-success outcome appears, and former authentication state cannot authorize a protected request.
6. **Given** invalid or missing credentials, **When** sign-in is attempted, **Then** no authenticated session is granted and the common failure response contains a safe message and request ID.

**Traceability**: E1 Principles III–V and role decision; E2 qualified access rules, LoginSrv and LogoutSrv; E3 registration/login and logout.

---

### User Story 2 - Browse products and manage a personal cart (Priority: P1)

As a permitted shopper, I can find products, view images, maintain cart quantities and learn about unavailable stock.

**Why this priority**: Discovery and quantity management are primary customer activities.

**Independent Test**: Seed synthetic products, images, memberships and cart rows. Exercise reads as viewer and documented customer writes as permitted editor independently of checkout.

**Acceptance Scenarios**:

1. **Given** products with positive and zero stock, **When** a permitted user lists, filters or searches the catalog, **Then** documented type/name/description matching applies and zero-stock products remain visible.
2. **Given** a permitted image request, **When** an image exists, **Then** stored bytes are served; **When** absent, **Then** the documented fallback is used when available, with exact asset/content-type parity UNVERIFIED under O1.
3. **Given** cart quantity 2 and stock 5, **When** a permitted editor requests increment 4, **Then** cart quantity becomes 5, unmet demand is 1, and availability-email promise text appears only if demand recording succeeds.
4. **Given** zero stock and an existing cart, **When** the add journey receives a positive increment, **Then** the cart is unchanged and reports "Product is Out of Stock!"; shortage-demand behavior still follows FR-013.
5. **Given** cart quantity 3, **When** the add journey receives zero increment, **Then** documented single-unit removal leaves quantity 2; **When** quantity is instead replaced with zero, **Then** the existing row is deleted.
6. **Given** no cart row, **When** the documented service-level add is exercised, **Then** it changes nothing; **When** the replacement operation is exercised, **Then** it inserts the supplied quantity. Planning must preserve the distinction without inventing a public endpoint.
7. **Given** two cart rows with quantities 2 and 3, **When** total units are read, **Then** the count is 5 and other users' entries are excluded.
8. **Given** existing user/product demand quantity 1, **When** another shortage is recorded for that pair, **Then** recording succeeds without changing its original quantity.

**Traceability**: E2 CartServiceImpl FR-048–073, DemandServiceImpl FR-074–080, ProductServiceImpl FR-111–114 and FR-123–126, AddtoCart, UpdateToCart and ShowImage; E4 cart and handler descriptions.

---

### User Story 3 - Place a complete demo order and review history (Priority: P1)

As a permitted shopper, I can convert my cart into an order without a failed checkout leaving a partly saved purchase.

**Why this priority**: Checkout changes several record groups; the operator requires atomic persistence and continued purchase success despite mail failure.

**Independent Test**: Seed a cart, fixed prices and permitted editor. Inject failures at each persistence step and separately at notification delivery.

**Acceptance Scenarios**:

1. **Given** an empty cart, **When** checkout is requested, **Then** "Order Placement Failed!" is returned without order, stock, transaction or confirmation changes.
2. **Given** products priced 12.50 and 3.00 with cart quantities 2 and 4, **When** demo checkout succeeds with submitted amount 40.00, **Then** line amounts are 25.00 and 12.00, both lines share a transaction identity, recorded payment remains 40.00 without total comparison, shipping starts at zero, stock decreases by 2 and 4, and purchased cart entries are removed.
3. **Given** a populated cart, **When** any order insertion, cart removal, stock change, transaction insertion or final save fails, **Then** all checkout persistence remains as before the attempt, no completed-order confirmation is sent, and the legacy failure outcome is retained.
4. **Given** a saved order, **When** confirmation delivery fails, **Then** "Order Placed Successfully!" remains the outcome, order/stock remain saved, failure is logged, a background retry is retained, and temporary resources are cleaned up.
5. **Given** two users' orders in the same project, **When** either reads personal history, **Then** only its own transaction-linked lines appear, with documented current product details, timestamp, amounts and shipping state.
6. **Given** simultaneous successful product or transaction creations, **When** identities are allocated, **Then** there are no duplicates and documented visible identifier shape is retained independently of clock readings.

**Traceability**: E2 OrderServiceImpl FR-081–099, OrderServlet FR-162, TransServiceImpl FR-127–128, TransactionBean and IDUtil; E1 recorded parity decisions 1–3. This is a demo flow and asserts no real charge.

---

### User Story 4 - Maintain inventory and notify waiting customers (Priority: P2)

As a permitted administrator, I can maintain products while preserving existing orders and notifying customers of increased stock.

**Why this priority**: Inventory sustains shopping and historical product references must remain valid.

**Independent Test**: Use products with and without order references, synthetic demand records and controlled mail outcomes.

**Acceptance Scenarios**:

1. **Given** an admin within scope, **When** documented product fields and a content-valid image within the 16,177,215-byte per-file limit are submitted, **Then** the product is saved and "Product Added Successfully with Product Id: " plus its assigned identity appears.
2. **Given** a product, **When** the image-preserving update changes metadata, price and quantity, **Then** the image remains; mismatched old/new product identities are rejected.
3. **Given** demanders, **When** a successful image-preserving update increases stock, **Then** each receives a notification attempt regardless of demanded quantity and its demand is removed after the attempt even on failure. Failed delivery is logged and retained for background retry independently of demand cleanup.
4. **Given** unchanged/decreased stock, **When** an update succeeds, **Then** product-update success appears without stock-increase notifications.
5. **Given** a product referenced by order lines, **When** deletion or identity change is attempted, **Then** reference restrictions preserve existing records.
6. **Given** an unreferenced product in carts, **When** product deletion succeeds, **Then** matching cart entries are cleaned up; **When** deletion fails, **Then** cart cleanup does not run.
7. **Given** an over-limit upload or non-image content disguised by its filename, **When** registration is attempted, **Then** no product/image is saved, a correlated validation failure is returned, and temporary files/records are cleaned up.

**Traceability**: E2 mysql_database FR-024–026, ProductServiceImpl FR-100–122, AddProductSrv FR-137–138, UpdateProductSrv FR-170 and RemoveProductSrv FR-165; E1 parity decisions 3–4.

---

### User Story 5 - Ship a selected order line (Priority: P2)

As a permitted administrator, I can mark an order/product line shipped and see the corresponding fulfillment result.

**Why this priority**: Shipping completes the purchase lifecycle.

**Independent Test**: Seed an unshipped line, matching transaction/product, memberships and controlled notification outcomes.

**Acceptance Scenarios**:

1. **Given** a matching unshipped line and permitted admin, **When** shipment is requested, **Then** only that line changes from zero to one, "Order Has been shipped successfully!!" appears, the shipped-items outcome is displayed, and notification is attempted.
2. **Given** an already shipped or missing pair, **When** shipment is requested, **Then** no line changes; a result equal to FAILURE ignoring case selects the unshipped-items outcome without the success-notification branch.
3. **Given** editor, viewer or an admin outside scope, **When** shipment is requested, **Then** HTTP 403 is returned, membership refusal is audited, and no shipment/mail occurs.
4. **Given** an authorized shipment with submitted recipient and amount, **When** a documented non-FAILURE result occurs, **Then** notification uses those documented inputs. They cannot confer authority over another order or scope.

**Traceability**: E2 OrderServiceImpl FR-098–099 and ShipmentServlet FR-166–168; E3 admin duties; E1 III–V override the legacy non-null-role gate.

---

### User Story 6 - Register an account and send a contact message (Priority: P3)

As a permitted signed-in actor, I can use documented registration/contact forms and receive their documented outcomes.

**Why this priority**: Supporting journeys must survive with the operator's authentication restrictions.

**Independent Test**: Use an existing permitted session, synthetic registration details, duplicate/mismatch cases and a controlled mail receiver.

**Acceptance Scenarios**:

1. **Given** a permitted registration actor, **When** a unique email and matching non-null password/confirmation are submitted with documented profile fields, **Then** the account is created only after successful save, "User Registered Successfully!" appears, and welcome notification uses the first space-separated name component.
2. **Given** an existing email, **When** registration is submitted, **Then** "Email Id Already Registered!" appears without another account or new-account success mail.
3. **Given** a missing/mismatching password, **When** registration is submitted, **Then** "Password not matching!" appears and no account is created.
4. **Given** permitted contact submission, **When** mail delivery returns exact SUCCESS, **Then** "Comments Sent Successfully" appears; every other result retains the documented mail-configuration failure message.
5. **Given** no authentication or applicable membership, **When** either form is submitted, **Then** it is refused before account creation or mail. Creating an account alone grants neither membership nor admin privileges.

**Traceability**: E2 UserServiceImpl FR-129–136, RegisterSrv FR-163–164, FansMessage FR-148–149 and MailMessage FR-180–181; E3; E1 III–V. Missing first-account/reset contracts remain O2.

### Edge Cases

- E2's zero-increment description says product removal, but its quoted call invokes single-unit removal; E4 explicitly subtracts one. Preserve that concrete effect. Quantity replacement with zero deletes an existing row. Exercise quantity 1 and greater than 1 separately.
- Existing cart rows ignore negative replacement; absent rows accept zero/negative insertion. Stock sales subtract the supplied quantity without a nonnegative guard, so negative sales increase stock. Preserve documented rules without silently adding price/quantity policies (E2 FR-065–070/124; E4 CartServiceImpl).
- Service-level add requires an existing row; the handler may create one through replacement. Zero-stock add leaves cart unchanged, whereas replacement capped to zero can delete it. Do not collapse these paths (E2 CartServiceImpl, AddtoCart, UpdateToCart).
- Repeated demand keeps its original quantity. Missing-demand removal succeeds; missing-cart full removal fails. Cart aggregate counts sum quantities, but per-item lookup reads the first matching row; null identity inputs return zero without querying. Cart uniqueness is UNVERIFIED (E2 FR-064/071–080).
- Missing price/stock returns zero; missing transaction owner returns empty internally. Those fallbacks cannot authorize a request or conceal a failed web operation (E2 FR-123/126–128; E1 III/V).
- Test required/optional values, exact length limits and one beyond, duplicate identities, monetary precision and reference restrictions. Null and zero remain distinct where persistence permits null (E2 schema; E1 Migration Constraints).
- Inject checkout failure on first/later items and final save; assert unchanged orders, transaction, stock and cart together. Inject mail failure after save separately (D1/D4).
- Exercise identity allocations within one clock second and across the legacy 12-hour ambiguity. Creation timestamps remain separate from identity allocation (D3).
- Test uploads exactly at the limit and one byte above, filename/content mismatch, absent image, unavailable original fallback, and cleanup on success/failure. Image-type allowlist and effective MIME values remain UNVERIFIED (O1).
- Missing/malformed numeric inputs produce safe correlated failures without an unchecked crash or partial mutation; exact unsupplied legacy status/encoding remains UNVERIFIED (E1 V; E4 RegisterSrv/UpdateProductSrv/OrderServlet).
- Registration, availability and shipment mail failures also require logging, retry and cleanup. Contact still reports the documented immediate delivery result rather than claiming delivery merely because work was queued (FR-026–028).

## Requirements *(mandatory)*

### Functional Requirements

Local FR identifiers below are independent of E2's identifiers. E2 repeats FR-001–014 in its separate access section; such citations are explicitly qualified. US1–US6 refer to the stories above.

- **FR-001 — Declared rebuild**: The candidate MUST rebuild the documented Java store in Go for run run-4e3ac24f2fa04792, preserving sourced business behavior except D1–D6. Local results MUST NOT be presented as platform acceptance. **Basis/verification**: operator input, E1 I/Governance; US1–US6, SC-001.
- **FR-002 — Source-only scope**: Every rebuilt operation, page, field and behavior MUST cite permitted legacy documents or explicit operator requirements. Missing/truncated evidence MUST stay UNVERIFIED and open; routes, constraints, assets and business capabilities MUST NOT be invented. **Basis/verification**: E1 inventory decision, E5–E7; inventories, O1–O2, SC-007.
- **FR-003 — Authentication**: Every endpoint MUST require an authenticated session except login and password reset, including catalog, images, registration and contact. An anonymous exception supplies no missing workflow. **Basis/verification**: E1 III; US1.1–2, US6.5; D2, O2.
- **FR-004 — Roles and scope**: Every protected operation MUST enforce stored admin/editor/viewer permissions and applicable workspace/project membership on the server. Customers map to viewer for reads and to editor only for writes explicitly shown in E3, never admin. Membership does not grant another user's personal records. Inventory/shipping administration remains admin work as E3 states. Every plan permission cell MUST have a justified basis. **Basis/verification**: E1 III/role decision, E3; US1.3–4, US3.5, US5.3.
- **FR-005 — Refusal and coverage**: Role/membership refusal MUST return HTTP 403 before business effects. Workspace/project refusals MUST be audited with the request ID. Each endpoint MUST have a case per admin/editor/viewer, denied membership and anonymous cases as applicable; coverage MUST NOT manufacture positive grants. **Basis/verification**: E1 III–IV; US1.3, US5.3; SC-002.
- **FR-006 — Credentials and sessions**: All accounts including admins MUST use stored credentials with constitution-mandated password protection and hashed stored session tokens. Submitted role text MUST NOT confer privileges. Credentials, passwords and plaintext tokens MUST stay out of logs/audit. Logout MUST clear authenticated identity/profile access. **Basis/verification**: E1 III; E2 LoginSrv/LogoutSrv; D2, US1.1/5/6, SC-002.
- **FR-007 — Failure contract**: Every non-2xx response MUST use `{"ok": false, "error": {"code": "...", "message": "...", "request_id": "..."}}`, including authentication, refusal, validation, unexpected failures and any specified redirects. Code MUST identify the failure; message MUST be safe and retain documented legacy wording where applicable. Response, structured request log and refusal audit MUST share the request ID. **Basis/verification**: E1 V; D5; all negative cases, SC-002.
- **FR-008 — Product discovery**: Permitted users MUST list all products including zero stock, retrieve details/images, filter by contained type and search contained type/name/description. Matching MUST retain the documented lowercased stored-field comparison, original supplied search text and unescaped wildcard behavior, without adding unsourced case normalization. **Basis/verification**: E2 FR-111–114; E4; US2.1–2.
- **FR-009 — Images**: Images MUST retain documented raw-byte, fallback and content-type behavior subject to O1. Uploads MUST enforce 16,177,215 bytes per file and actual-content validation. Unavailable assets MUST use neutral placeholders and remain individually open with UNVERIFIED parity. **Basis/verification**: E2 FR-137/169, E4 ShowImage, E1 parity 4–5; D6, US4.7, SC-006.
- **FR-010 — Cart reads**: Listing/counts MUST use the authenticated user's permitted records. Cart/product totals MUST sum quantities. The separate per-item lookup MUST return zero without querying on null identities and otherwise the first matching quantity. **Basis/verification**: E2 FR-055–056/071–073; US2.7, edge cases.
- **FR-011 — Distinct cart changes**: Service-level addition MUST act only on an existing pair and add increment to quantity. The customer add journey MUST use existing quantity plus increment and its documented replacement path when appropriate. Zero increment MUST invoke single-unit removal, subtracting one and deleting at a nonpositive remainder. Full-product removal MUST remove matching rows and succeed only when at least one is deleted. **Basis/verification**: E2 FR-048–064/139–147, E4; US2.3–6.
- **FR-012 — Replacement**: Existing rows MUST store positive replacement, delete at zero and ignore negative replacement. Absent pairs MUST insert the supplied quantity, including zero/negative. Stock capping MUST remain in documented request flows rather than being added to uncapped replacement. Status MUST reflect documented positive changed-row outcomes and sourced wording. **Basis/verification**: E2 FR-065–070, E4; US2.5–6, edge cases.
- **FR-013 — Shortage and demand**: Requests above stock MUST cap to availability and record target minus stock as demand. Zero-stock add MUST leave cart unchanged and show the out-of-stock message while following its demand path. Promise text MUST appear only when demand recording succeeds. Existing demand MUST succeed unchanged; new demand MUST store supplied quantity and succeed only after save. Removal MUST target that user/product and succeed if absent. Product-demand lookup MUST include all demanders regardless of quantity. **Basis/verification**: E2 FR-050–053/074–080/141–146/171–175; US2.3–4/8.
- **FR-014 — Demo payment**: Checkout MUST treat the session user's supplied numeric amount as successful demo payment without comparing line totals, MUST NOT introduce real payment processing, and MUST fail an empty cart without effects. **Basis/verification**: E2 FR-081–082/162, E3; US3.1–2.
- **FR-015 — Order construction**: Checkout MUST use current price times quantity for each line, share one transaction identity, start lines unshipped, remove purchased cart rows, subtract purchased stock and save the transaction only after required preceding work succeeds. Supplied transaction amount and recorded creation time MUST be retained. **Basis/verification**: E2 FR-041–047/083–094; US3.2.
- **FR-016 — Atomic checkout**: All checkout persistence MUST succeed together or restore cart, stock, order lines and transactions to pre-attempt state, including on final-save failure. A failed attempt MUST NOT send completed-order confirmation. Success/failure MUST retain "Order Placed Successfully!" / "Order Placement Failed!". **Basis/verification**: E1 parity 1 overrides E2 partial-state consequences; D1, US3.1–4, SC-003.
- **FR-017 — History and sales**: Personal history MUST include only lines joined to the authenticated owner's transactions; detailed history MUST also match current product and show sourced identity/name/image/quantity/amount/time/shipping fields. Sold quantities MUST sum product order lines. Transaction-owner lookup MUST use transaction identity, return empty internally if missing, and never confer permission. **Basis/verification**: E2 FR-095–097/127–128, E4 OrderDetails; US3.5, edge cases.
- **FR-018 — Product changes**: Permitted admins MUST register supplied product fields, generate identity where required and retain supplied identity where the documented bean operation allows it. Full updates MUST replace all sourced fields including image; image-preserving updates MUST change only metadata/price/stock. Both MUST reject mismatched old/new identity. Price-only changes MUST replace supplied price. Success MUST require documented positive save result and use sourced status. **Basis/verification**: E2 FR-100–110/115–116/121–122/138/170; US4.1–2/4.
- **FR-019 — Deletion**: Product deletion MUST precede cart cleanup; cleanup MUST run only after positive product deletion. A referenced product MUST NOT be deleted or renamed while order lines reference it. **Basis/verification**: E2 FR-024–026/104–105/165; US4.5–6.
- **FR-020 — Stock and availability**: Stock sales MUST subtract supplied quantity without a new nonnegative guard and retain documented changed-row results. Missing stock/price MUST default to zero. Only a successful image-preserving increase MUST notify all product demanders, then remove each demand even on notification failure. The sourced notification-status suffix MUST depend on demand removal, not delivery proof. Retry information MUST survive removal. **Basis/verification**: E2 FR-117–126; E1 parity 3; US4.3–4, edge cases.
- **FR-021 — Shipping**: Permitted admin shipment MUST target the selected unshipped line and change zero to one. Positive transition MUST retain sourced success wording. FAILURE ignoring case MUST choose the unshipped outcome; other documented results MUST choose shipped outcome and notify using submitted recipient/amount after scope/permission checks. **Basis/verification**: E2 FR-098–099/166–168, E3; US5.1–4.
- **FR-022 — Registration**: Authenticated permitted registration MUST accept documented profile fields and password confirmation, prevent duplicate email, require non-null matching password/confirmation, and report success only after save. Duplicate/mismatch/success messages MUST remain exact. Creation MUST NOT itself grant membership or elevated role. Password protection supersedes legacy plaintext representation. **Basis/verification**: E2 FR-129–136/163–164; E1 III; US6.1–3/5, O2.
- **FR-023 — Profile**: Profile lookup MUST require authenticated identity. First name MUST retain the first single-space-separated component and address lookup MUST select by account email. Internal null/empty failure fallbacks MUST NOT bypass access checks, reveal another user or turn failed web operations into success. **Basis/verification**: E2 FR-132/135–136 and access-section FR-002; E4 UserServiceImpl; US1, US6.1.
- **FR-024 — Identity integrity**: Generated identities MUST be unique independently of clock generation, retaining visible P/T prefix plus fourteen decimal digits and existing length limits; creation times remain separate. Planning MUST apply E1's sequence-or-UUID requirement. **Basis/verification**: E2 FR-176–179; E1 parity 2; D3, US3.6, SC-004.
- **FR-025 — Data constraints**: Data MUST retain evidenced required fields, lengths, nullability/defaults, precision, keys and product references except explicitly superseded credential representation. Unknown constraints MUST remain UNVERIFIED. No separate order-header table or extra uniqueness/reference rule may be inferred from advisory descriptions. **Basis/verification**: E2 schema FR-001–040; E1 Migration Constraints; Key Entities, O2, SC-005.
- **FR-026 — Event mail**: Successful registration, saved order, documented shipment outcome and stock-increase demand processing MUST trigger documented notifications with sourced recipient/context, first name, identity, product, amount and demo notice. Advertised promotions MUST NOT become discount rules. **Basis/verification**: E2 FR-090/118/132/168, E4 MailMessage/JavaMailUtil, E3; US3.4, US4.3, US5.1/4, US6.1.
- **FR-027 — Retry and cleanup**: Failed event mail MUST NOT fail/reverse a completed order; failure MUST be logged and retained for background retry independently of demand cleanup. Temporary files/records MUST be cleaned on all paths; retained pending notification is business work, not disposable temporary data. **Basis/verification**: E1 parity 3; D4, US3.4, US4.3, SC-006. Retry scheduling is a plan decision, not an invented delivery guarantee.
- **FR-028 — Contact**: A permitted authenticated actor MUST submit documented name/email/comments to the designated legacy recipient. Only exact SUCCESS MUST show "Comments Sent Successfully"; all other results MUST retain "Failed: Please Configure mailer.email and password in application.properties first" as legacy message text, without imposing Java configuration on the Go target. **Basis/verification**: E2 FR-148–149/180–181, E4 FansMessage; US6.4–5.
- **FR-029 — Pages**: Documented flows MUST retain sourced outcomes, messages, fields and controls except D2/D5 security/failure differences. Page names below are evidence labels, not fabricated route mappings. Every rebuilt page MUST have browser coverage; incomplete visual evidence MUST remain UNVERIFIED. **Basis/verification**: E1 IV/parity 5, E4, E6; US1–US6, O1, SC-007.
- **FR-030 — Test-first gates**: Acceptance tests MUST precede implementation with expected failure established first, use synthetic data and cover sourced rules/differences. Plan MUST provide endpoint × role permissions and inherit all constitution formatting, static-analysis, unit/race, acceptance/browser, configuration, migration, logging and changelog gates. Missing evidence/gates MUST NOT be reported as passing. **Basis/verification**: E1 II/IV/V/Migration Constraints/Development Workflow; SC-001–007. This spec is no implementation, merge or deployment result.

### Key Entities *(include if feature involves data)*

Source-only inventory follows. Representation belongs to planning; a bean's zero default does not override a documented nullable persisted value.

| Entity | Attributes and relationships | Evidence / limitation |
| --- | --- | --- |
| Product | Required unique identity up to 45 characters; nullable/default-absent name up to 100, type up to 20, description up to 350, price with 12 decimal digits including 2 fractional, integer stock and image. | E2 schema FR-001–013; reference restrictions FR-024–026; generated identities change under D3. |
| User/account | Required unique email up to 60; nullable/default-absent name up to 30, large-integer mobile, address up to 250 and integer postal code. | E2 schema FR-027–035. Legacy nullable plaintext password up to 20 (FR-036–037) is replaced under D2/E1 III; replacement credential layout is design, not legacy evidence. |
| Cart entry | User identity, product identity, integer quantity. | E4 CartBean/CartServiceImpl, E2 FR-048–073. Complete persisted types, defaults, nullability, keys and references UNVERIFIED (O2). |
| Product demand | User identity, product identity, demanded integer quantity; repeat recording leaves original quantity. | E4 DemandBean/DemandServiceImpl, E2 FR-074–080. Persisted definitions UNVERIFIED (O2); duplicate business behavior does not prove a unique key. |
| Order line | Required order/transaction and product identities up to 45 each; one line per pair; nullable/default-absent integer quantity and amount with 10 decimal digits including 2 fractional; required shipping flag default zero. Existing product required and protected from referenced deletion/rename. | E2 schema FR-014–026, FR-041/092/096. No separate order-header table or transaction foreign key is established. |
| Transaction | Required unique identity up to 45; username, timestamp and supplied amount; history joins lines through order identity. | E2 schema FR-038–040 and FR-042–047/082/096; E4 TransactionBean. Persisted username/time/amount types/defaults/nullability UNVERIFIED (O2). |
| Order-detail view | Order/product identities, current product name/image, displayed quantity/amount, timestamp, shipping. | E4 OrderDetails, E2 FR-097; a display view, not another evidenced table. |
| Session and role grant | Authenticated account/profile context, granted admin/editor/viewer role, protected token. | E1 III; E2 LoginSrv/LogoutSrv describe legacy context. Storage/provisioning/credential conversion UNVERIFIED (O2). |
| Workspace/project membership | Scope identity, user membership and applicable grants. | Explicit operator requirement E1 III/role decision; no legacy schema established (O2). |
| Refusal audit/request correlation | Refused access and related user/scope, request ID shared by response/log; excludes credentials/plaintext tokens. | Explicit operator requirement E1 III–V, not a discovered legacy table. |
| Pending notification | Recipient, documented message context and retry state surviving delivery failure and demand cleanup. | E1 parity 3, E4 MailMessage; storage is plan design. |

### Evidence and scope inventory

| ID | Permitted source | Use and strength |
| --- | --- | --- |
| E1 | [Constitution v1.0.1](../../.specify/memory/constitution.md) | Recorded operator decisions establish direction, security, intentional differences and gates; not legacy runtime observations. |
| E2 | [Legacy PRD](../../../../old/docs/prd.md), cross-checked with [BRD](../../../../old/docs/brd.md) | Component-qualified rules with legacy filenames/evidence snippets: 181 requirements plus 14 access entries. Source characterization, not execution-grounded results. |
| E3 | [Legacy summary](../../../../old/docs/summary.md) | Operator-selected boundary for customer write grants. Model-composed summary, not a URL/schema inventory. |
| E4 | [Overview](../../../../old/docs/overview.md), [LLD](../../../../old/docs/lld.md), [HLD](../../../../old/docs/hld.md) | Component descriptions, signatures, data holders and effects; zero execution-grounded inventory. HLD edges are references, not a runtime call graph. Unattributed LLD symbols remain unattributed. |
| E5 | [API-surface limitation](../../../../old/docs/api_surface.md), [data-model limitation](../../../../old/docs/data_model.md) | Instrument did not record endpoints/persistence; this is not evidence they are absent. |
| E6 | [Screen limitation](../../../../old/docs/screens.md), E4 asset summaries | Complete screen inventory unavailable; style/script assets and image-bearing seed partly truncated. Binary assets unavailable in this document set. |
| E7 | [Security](../../../../old/docs/security.md), [edge cases](../../../../old/docs/edge_cases.md), [workflows](../../../../old/docs/workflows.md) | Searches could not run; no absence of checks/errors/orchestration may be inferred. |
| E8 | [Approach](../../../../old/docs/approach.md), [roadmap](../../../../old/docs/roadmap.md) | Advisory analysis/decomposition, not operator decisions. Both agree with Java-to-Go. D1–D6 supersede incompatible behavioral parity suggestions. |

The reading reports 33 behavior-bearing components and 195 rules; these are traceability denominators, not passing test counts. Accessors and client-library helpers require outcome assessment; their presence does not prove distinct public endpoints.

The following disposition covers every E2 numbered rule. "Preserved" means required by this specification, not tested or execution-verified. Where a row names an exception, only the named behavior changes; the remaining sourced outcomes stay required. Exact route, screen and physical-schema gaps remain O1/O2 rather than being invented.

| E2 rule group | Required disposition | Local coverage |
| --- | --- | --- |
| Requirements by area FR-001–035 | Preserve evidenced product/user/order-line constraints. | FR-019/025; Key Entities; O2 for additional missing definitions. |
| FR-036–037 | Replace legacy plaintext password storage/length representation under D2; preserve unrelated account fields. | FR-006/022/025. |
| FR-038–041 | Preserve transaction identity constraints and default unshipped order lines. | FR-015/024/025. |
| FR-042–047 | Preserve supplied/generated timestamps and constructor outcomes; generated identity mechanism changes under D3. | FR-015/024. |
| FR-048–080 | Preserve distinct cart operations, counts, statuses and demand behavior. | FR-010–013; cart edge cases. |
| FR-081–094 | Preserve empty-cart outcome, calculations, sequencing prerequisites and changed-record success; partial persistence changes under D1 and event-mail failure under D4. | FR-014–016/026–027. |
| FR-095–099 | Preserve sales/history joins and selected-line shipping transition/results within authorized scope. | FR-017/021. |
| FR-100–110 | Preserve product creation/update/removal outcomes; generation mechanism changes under D3. | FR-018–019/024. |
| FR-111–128 | Preserve matching, image-preserving updates, demand attempt/removal/status, stock/pricing and owner-lookup outcomes; add D4 retry/cleanup, and apply D2 before exposing records. | FR-008/017–020/026–027. |
| FR-129–136 | Preserve duplicate detection, save outcome, first-name/address helpers and valid-credential result; plaintext persistence and credential comparison change under D2. | FR-006/022–023/026. |
| FR-137–149 | Preserve upload size/fields, quantity/shortage/demand and contact outcomes; add D6 content validation and D2 authentication. | FR-009/011–013/018/028. |
| FR-150–157 | Replace hard-coded/admin-branch and submitted-role/password session authority under D2; preserve authenticated identity/profile context and documented successful destination outcomes. | FR-003–006/029. |
| FR-158–161 | Preserve clearing of logout identity/role/profile access; no plaintext session password is retained to clear under D2. | FR-006; US1.5. |
| FR-162–175 | Preserve demo amount, registration checks, selected removal/shipment, fallback and image-preserving/cart changes; apply D2 access, D5 failures and D6 unavailable-asset handling. | FR-009/012–014/018–023/029. |
| FR-176–179 | Preserve visible identifier shape; replace clock-dependent generation under D3. | FR-024. |
| FR-180–181 | Preserve exact generic-message SUCCESS/FAILURE result semantics. | FR-028; event notification retry separately FR-027. |
| Who can do what FR-001–014 | Replace plaintext credential/profile matching, hard-coded admin, submitted legacy role gates and execution after refused redirects under D2/D5. Preserve valid-account sign-in and properly authorized business operations. | FR-003–007; US1/US5.3/US6.5. |

| Operation owner | Documented inputs/outcomes | Contract status and source |
| --- | --- | --- |
| LoginSrv | Username/password/usertype, identity/profile session context, admin/customer home outcome. | Exact path/methods/full form/failure encoding UNVERIFIED; E2 access section and E4 LoginSrv. D2 replaces credentials/submitted role behavior. |
| LogoutSrv | Clears identity/password/role/profile attributes; login success outcome; POST delegates to GET. | Exact mapping UNVERIFIED; E4 LogoutSrv, E2 FR-158–161. |
| RegisterSrv | Name/mobile/email/address/pincode/password/confirmPassword; register.jsp status. | Exact mapping/methods/full validators UNVERIFIED; E4, E3; authentication required by E1. |
| AddProductSrv | Multipart name/type/info/price/quantity/image; addProduct.jsp status; upload limit. | Exact mapping/methods UNVERIFIED; E4 AddProductSrv. LLD unattributed doGet/doPost resembles this flow but ownership must not be guessed. |
| UpdateProductSrv | Product identity, metadata, price, quantity; image-preserving update; updateProduct.jsp status. | Exact mapping/methods/unsupplied parameter names UNVERIFIED; E4. |
| RemoveProductSrv | prodid; removeProduct.jsp status. | Exact mapping/methods UNVERIFIED; E4. |
| ShowImage | pid; stored bytes or images/noimage.jpg encoded as JPEG; described servlet sets no explicit response MIME. | Mapping/methods/deployed fallback bytes/effective MIME UNVERIFIED; E4. |
| AddtoCart | pqty, selected product, session user; target quantity/demand; home/cart status. | Mapping/methods/full parameter names and destination selection UNVERIFIED; E4, E2 FR-139–147. |
| UpdateToCart | pid/pqty, session user; cartDetails.jsp status script. | Mapping/methods/complete script UNVERIFIED; E4, E2 FR-171. |
| OrderServlet | Numeric amount, session username; orderDetails.jsp demo result. | Mapping/methods UNVERIFIED; E4, E2 FR-162. |
| ShipmentServlet | Order/product pair, submitted recipient/amount; shippedItems.jsp or unshippedItems.jsp. | Mapping/methods/unsupplied parameter spellings UNVERIFIED; E4, E2 FR-166–168. |
| FansMessage | Name/email/comments, designated legacy recipient; index.jsp and mail alert. | Mapping/methods/full validators UNVERIFIED; E4, E2 FR-148–149. |
| Product/cart/demand/user/order/transaction services | Business operations and lookups specified above. | E2/E4 business operations are not evidence of distinct public routes. |
| Password reset | Anonymous-route exception only. | E1 III requirement; route/inputs/recovery proof/workflow UNVERIFIED. Exception does not establish legacy functionality (O2). |

Page evidence names login.jsp, register.jsp, addProduct.jsp, updateProduct.jsp, removeProduct.jsp, cartDetails.jsp, orderDetails.jsp, shippedItems.jsp, unshippedItems.jsp and index.jsp, plus admin/customer home outcomes (E4 servlet descriptions). These names are not verified URL mappings or a complete screen inventory. Full layouts, navigation, validators and controls remain UNVERIFIED (E6/O1). Documented form fields above are retained; missing controls cannot be inferred.

Client open items (E4/E6): images/noimage.jpg and other image/screenshot contents have no binary evidence; bootstrap-grid styles, bootstrap and bootstrap.min styles, bootstrap.bundle and bootstrap.bundle.min scripts, and non-minified bootstrap script include reported truncation. Complete minified grid/reboot styles and bootstrap.min script are described, but their files and actual dependencies are not supplied here either. Each unavailable/truncated asset remains UNVERIFIED for exact parity and uses a neutral placeholder where necessary. No downloaded/invented replacement may be represented as the original.

### Named deliberate differences from legacy

These are operator decisions recorded in E1, not new choices made here. No read document contradicts the Java-to-Go target. Advisory conflicts below concern behavior and have recorded operator resolutions.

| ID | Legacy / advisory behavior | Required difference and acceptance consequence |
| --- | --- | --- |
| D1 — Atomic checkout | E4 OrderServiceImpl, E2 FR-085–091 and E8 roadmap item 018 allow sequential partial persistence. | E1 parity 1 requires a single transaction and rollback. US3.3 must reject partial state; success/failure wording remains. |
| D2 — Protected accounts and scope | E2 access rules/E4 describe plaintext passwords, hard-coded admin, session passwords/submitted roles, anonymous image/contact/registration, weak shipment checks and execution after denial redirects. E8 suggests carrying them forward. | E1 III/role decision requires Argon2id password protection, hashed stored tokens, granted roles, authenticated non-exempt endpoints, scope checks and immediate refusal termination. US1/US5.3/US6.5 and secret-storage/log checks replace insecure expectations. |
| D3 — Collision-free identity | E2 IDUtil and E8 item 003 use second-resolution 12-hour clock IDs without collision checks. | E1 parity 2 requires database sequence or UUID generation independent of time while preserving visible format. US3.6/SC-004 verify uniqueness and P/T plus fourteen digits. |
| D4 — Mail retry and cleanup | E4 event helpers swallow failure, demands are removed after attempts, resources may remain open; E8 proposes mirroring outcomes. | E1 parity 3 requires background retry/logging without order failure/rollback and cleanup on all paths. Retain documented immediate statuses and demand removal while independently retaining retry information. US3.4/US4.3 and failure-injection cases verify it. |
| D5 — Correlated failures | E4 describes unchecked parsing, stack traces and redirect/include/script failures; E8 suggests similar flows. | E1 V requires common non-2xx error envelope and structured request-ID logs; preserve sourced safe message text inside that contract. Do not reproduce raw traces or bare non-2xx pages/redirects. Every negative case verifies it. |
| D6 — Validated upload and explicit asset gaps | E4 has uploaded streams, absent explicit image MIME, fallback and incomplete assets; E8 items 031–032 suggest wiring assets directly. | E1 parity 4–5 retains actual legacy fallback/content types, validates real uploaded content and uses neutral placeholders with open UNVERIFIED asset parity. US2.2/US4.7/SC-006 verify known behavior; O1 blocks unsupported parity claims. |

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001 — Behavioral accounting**: All 195 source-characterized rules have a disposition: preserved, replaced by a named operator decision, or UNVERIFIED with an open evidence item. Each rebuilt journey has passing acceptance cases before it is reported complete. Customers/admins retain their documented store tasks and feedback, with no silently dropped rule.
- **SC-002 — Access and failures**: Every inventoried endpoint has cases for all three roles, relevant membership refusals and anonymous behavior. Across 100% of denied cases there are zero protected disclosures/business effects; every membership refusal is audited and every failure shares response/log request ID. Storage/log checks reveal zero plaintext passwords or session tokens in prohibited locations.
- **SC-003 — Purchase integrity**: In 100% of checkout cases failing a line save, cart removal, stock update, transaction save or final completion, all four record groups match pre-attempt state and no completed-order confirmation is sent. Successful cases preserve calculations, supplied amounts, feedback and personal history; mail failure reverses zero orders.
- **SC-004 — Identity integrity**: A synthetic batch of 1,000 successful product creations and 1,000 successful transaction creations, including requests within one second, has zero duplicate identities and 100% retain visible shape. This specifies volume for verification, not a speed/traffic claim.
- **SC-005 — Data compatibility**: All evidenced required/optional/default, length, precision, unique-key and product-reference boundary cases pass. Referenced products cannot be deleted/renamed. No unknown constraint counts as verified.
- **SC-006 — Reliable side effects**: For each event-mail family, an injected failed send produces a correlated failure and retryable notification without order reversal or temporary-resource leak; a controlled subsequent success reaches the intended synthetic recipient. All over-limit/content-invalid uploads fail without saved product/image effects. Every unavailable asset is individually listed with unverified parity.
- **SC-007 — Usable journeys**: Every rebuilt page has browser coverage for its supported controls, validation and destination; documented messages match unless a named difference applies. E3's registration/login, catalog/images, cart, checkout, shipping, inventory, contact and logout flows can each be demonstrated within granted membership. Missing route/screen/asset evidence remains visible and does not count as completed parity.

These are candidate validation targets, not measured results or platform acceptance. There is no evidenced legacy performance baseline or traffic target, so no response-time or production-capacity claim is invented.

## Assumptions

- Scope comes only from operator input, E1 and the permitted legacy documents. Unrelated surrounding application code is not migration evidence. All writing remains in this working tree; legacy documents are read-only.
- E1 v1.0.1 records operator resolutions for Go tooling/configuration/migrations, roles, source-only inventory and parity. Superseded Python mandates are already resolved; no different language is reopened here. The plan inherits required gates.
- This one feature covers the rebuild. E8's 32 component specifications are advisory decomposition; plan/tasks must decompose this specification before implementation.
- Synthetic pre-existing accounts/grants/memberships isolate story tests; they do not invent production provisioning or grant registration to anonymous users. Authentication is not membership; account creation alone grants no scope or role upgrade.
- Payment remains a demo; mail uses controlled synthetic recipients in validation. No live mail, real charge, client-data copying, deployment or data import is authorized by this specification.
- Layer boundaries, typed centralized configuration, versioned migration design, credential transition, retry policy and endpoint permission matrix belong to the plan under E1. Incomplete seed text is not a loadable schema/dataset.
- No supplied acceptance boundary waives a constitution rule. Unknown routes/controls/schema definitions remain UNVERIFIED; unknown evidence cannot be counted as parity. Missing information is grouped into two explicit markers below.

### Open evidence and clarification items

**O1 — Observable contracts and client evidence**: UNVERIFIED. E5 provides no route mappings, E6 no complete screen inventory, and E4 no deployed binary assets or effective image MIME values. All named operation/page evidence is retained above. Neutral placeholders are approved for missing assets but cannot settle original parity.

[NEEDS CLARIFICATION: Which permitted document supplies missing route/method mappings, complete page controls/validators, actual fallback image and served content types, supported upload image types, and missing client assets for Java-to-Go parity? If unavailable, should these contracts stay open and UNVERIFIED through planning, or will the operator provide explicit replacement contracts?]

**O2 — Persistence and account lifecycle**: UNVERIFIED. Product/user/order-line and transaction-identity constraints are documented; cart/demand definitions and transaction username/time/amount persisted definitions are incomplete. Initial membership, first-account provisioning, existing-credential conversion and password-reset workflow are absent. E1 requires protected credentials, role/membership/audit/retry records, but does not establish their legacy layout or full lifecycle.

[NEEDS CLARIFICATION: Which permitted evidence or explicit operator replacement contract defines missing persisted constraints, first-account/membership provisioning, transition of existing credentials, and the password-reset workflow? If unavailable, should these dependencies stay open and UNVERIFIED during Java-to-Go planning?]

### Changelog

- 2026-11-10 — Created the Java-to-Go specification for run run-4e3ac24f2fa04792 from the active template, constitution v1.0.1 and permitted legacy reading. Added acceptance journeys, source-qualified requirements/inventories, named operator parity differences, measurable outcomes and two grouped evidence clarifications. No application implementation, plan, tasks, push, merge or deployment performed.
