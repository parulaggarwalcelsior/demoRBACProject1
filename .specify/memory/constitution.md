<!--
Sync Impact Report
Version change: 1.0.0 → 1.0.1
Version rationale: clarification of five open questions using the operator's answers;
the declared Java-to-Go goal and existing principles remain in force.
Modified principles:
- II. Typed, Layered, Maintainable Code
- IV. Acceptance Tests Before Implementation
- V. Consistent Errors, Configuration, and Observability
Modified sections: Migration Constraints (Go migrations, role/scope mapping, evidence inventory).
Added sections: none.
Removed sections: none.
Template synchronization: no template or command changes; consumers read this file at runtime.
Artifact/test impact: subsequent artifacts MUST use the resolved Go gates, configuration,
migrations, role/scope mapping, and source-only inventory rules below.
Deferred items: none; the operator has resolved the parity decisions below.
No template placeholders or governance dates are deferred.
Changelog: 2026-11-09 — adopted v1.0.0 with the operator's standards, security rules,
test-first workflow, migration guardrails, and explicit unresolved decisions.
2026-11-10 — clarified five operator decisions in v1.0.1; retained the unanswered parity marker.
-->

# Electronics Store Java-to-Go Migration Constitution

## Core Principles

### I. Evidence-Based Java-to-Go Migration

The operator's declared source language is **java**, the target language is **go**, and the
goal, verbatim, is **"java to go"**. Every specification, plan, task, implementation change,
and supporting document MUST serve that rebuild of the documented electronics store:
catalog, carts, demo checkout, inventory, shipping, and notifications.

- Work MUST use `../../old/docs/README.md` as the entry point to the legacy evidence.
  The documents under `../../old/docs/` and earlier artifacts in this working tree are the
  permitted sources. The surrounding directory's codebase is not the migration subject.
- Requirements and acceptance cases MUST trace to the legacy documentation or an explicit
  operator requirement. A model's advisory approach or roadmap MUST NOT be presented as an
  operator decision or proof of observed runtime behavior.
- Observable business behavior, data constraints, side effects, and failure paths MUST be
  preserved unless an operator requirement explicitly changes them. Each such change MUST
  identify the legacy behavior, intended replacement, and acceptance-test consequences.
- If a document contradicts the declared goal, the document being written MUST contain a
  plain-text clarification marker naming the goal, conflicting file, and conflict. The
  author MUST NOT silently choose either direction. A prior document overrides the goal
  only where it records the operator's actual choice between them.
- Silence or an undeclared value is not a contradiction. Uncertainty MUST be recorded as
  a plain-text clarification marker; it MUST NOT be filled with an invented fact.

Rationale: the rebuild needs traceable behavior and an explicit Go destination without
mistaking gaps in the legacy reading for absent functionality.

### II. Typed, Layered, Maintainable Code

- Go code MUST pass `gofmt`/`goimports`, `go vet`, and `golangci-lint`; failures MUST
  fail the build. Python type hints, `mypy --strict`, and `ruff check` do not apply under
  the operator's resolution below.
- SOLID separation MUST be enforced: routers and handlers perform HTTP adaptation only,
  the DAO layer performs SQL and persistence mapping only, and the service layer owns
  business rules and workflow orchestration. HTTP handlers MUST NOT contain SQL or
  business policy; DAOs MUST NOT decide HTTP responses or business permissions.
- Every module MUST remain under 1,000 lines of code. Dead code, commented-out code,
  and copy-paste duplication MUST NOT be introduced or retained in changed code.
- A new dependency MUST have code that imports it in the same reviewable change.
  Speculative dependencies MUST NOT be added.

Rationale: explicit responsibilities and small modules make the Go replacement reviewable
and allow legacy business rules to be tested independently of HTTP and SQL.

### III. Authenticated Sessions and Server-Side Access Control

- Every endpoint MUST require an authenticated session. Login and password reset are the
  only anonymous-route exceptions; no public catalog, image, contact, or registration
  exception may be inferred from the legacy system.
- The role set MUST be `admin`, `editor`, and `viewer`. Permissions MUST be checked on the
  server for every protected operation; UI visibility MUST NOT serve as authorization.
- A user outside the requested workspace or project MUST receive HTTP 403, and the
  refusal MUST be written to the audit log with the request ID. A refusal MUST terminate
  processing before business side effects occur.
- Passwords MUST be hashed with Argon2id. Session tokens MUST be stored hashed.
  Passwords, plaintext session tokens, and credentials MUST NOT appear in logs or audit
  records. Legacy plaintext credential storage and hard-coded administrative credentials
  MUST NOT be reproduced.

Rationale: the operator explicitly requires server-enforced access boundaries, including
the removal of documented servlet flows that continue after failed authorization.

### IV. Acceptance Tests Before Implementation

- Acceptance tests MUST be written before their corresponding implementation. Their
  expected failure MUST be established before adding the behavior, and they MUST pass
  after implementation. Test data MUST be synthetic and MUST NOT contain real client data.
- Every endpoint MUST have a test matrix for `admin`, `editor`, and `viewer`, covering
  allowed behavior under granted permissions and HTTP 403 under refused permissions or
  workspace/project membership. A positive test MUST NOT grant a role a permission merely
  to satisfy coverage. Login and password-reset tests MUST also cover their explicitly
  permitted anonymous behavior.
- Tests MUST verify that workspace/project refusals create audit records and prevent
  side effects. Non-2xx cases MUST assert the common error envelope and request ID.
- Unit tests MUST use table-driven Go tests, and `go test -race` MUST pass; failures MUST
  fail the build. The operator's resolution below replaces the pytest requirement.
  Every page MUST have Playwright browser tests.
- All required tests MUST pass before a change can be merge-eligible. No change with a
  red test may merge. Local test results MUST NOT be claimed as platform acceptance.

Rationale: tests written first make preservation and intentional changes to legacy
behavior measurable, including server-side access control hidden by the UI.

### V. Consistent Errors, Configuration, and Observability

- Only the `internal/config` package MUST read environment variables, once at startup,
  into a typed `Config` struct. No other package may call `os.Getenv`, under the
  operator's resolution below.
- Secrets MUST NOT be embedded in source, fixtures, documents, or Git history.
- Every non-2xx response MUST use the single JSON envelope
  `{"ok": false, "error": {"code": "...", "message": "...", "request_id": "..."}}`.
  The fields MUST contain a machine-readable error code, a user-safe message, and the
  associated request ID. This rule includes authentication failures, 403 refusals,
  validation errors, unexpected failures, and redirects if any are specified.
- Logging MUST be structured and include the request ID for request-associated events.
  Print statements MUST NOT be used. The same request ID MUST correlate the response,
  application log, and any access-refusal audit record.

Rationale: centralized configuration and correlated failures make the Go system operable
without preserving legacy stack-trace printing or exposing sensitive data.

## Migration Constraints

The Java/JDBC/Servlet/JSP store described in `../../old/docs/overview.md` is the source
estate. `../../old/docs/approach.md` and `../../old/docs/roadmap.md` agree with the Java-to-Go
direction. No existing operator choice of a different target is recorded in the materials
read for this adoption. The operator chose the declared Go target over the conflicting
Python-specific tooling, configuration, and migration mandates, as recorded below.

- Every database change MUST be a versioned up/down SQL migration in git, using
  `golang-migrate`, never applied by hand. Schema reconstruction and later changes MUST
  preserve documented keys, nullability, defaults, monetary precision, and references
  unless an explicit requirement changes them.
- The incomplete SQL seed and partial client assets MUST NOT be treated as complete,
  executable artifacts. No real client data may be copied into tests or fixtures.
- The latest operator requirements explicitly change the legacy's plaintext passwords,
  hard-coded credentials, unauthenticated non-exempt endpoints, and authorization bypasses.
  Specifications MUST record those deltas against `../../old/docs/approach.md` and
  `../../old/docs/roadmap.md`; advisory parity suggestions do not authorize those behaviors.
- The reading reports no execution-grounded inventory. The unavailable API, data-model,
  screen, security-search, edge-case-search, and workflow-search outputs MUST be treated
  as evidence limitations, not proof that routes, persistence, pages, or checks are absent.
- No acceptance boundaries beyond the operator's supplied requirements are stated.
  Authors MUST NOT invent platform acceptance decisions or declare the migration accepted.

The following six rules record the operator's answers. Independent work MUST continue;
a dependent decision MUST NOT be silently made or reported as resolved without operator
evidence.

The operator chose the declared Java-to-Go goal over AGENTS.md's conflicting Python-specific
gates. Python type hints, mypy strict, ruff, and pytest MUST NOT apply. Their Go equivalents
MUST pass and failures MUST fail the build: `gofmt`/`goimports`, `go vet`, `golangci-lint`,
and `go test -race` with table-driven tests.

The operator chose the declared Java-to-Go goal over AGENTS.md's conflicting literal
`config.py` requirement, while retaining centralized configuration. Only the
`internal/config` package MUST read environment variables, once at startup, into a typed
`Config` struct. No other package may call `os.Getenv`.

The operator chose the declared Java-to-Go goal over AGENTS.md's conflicting Alembic
requirement, while retaining versioned migrations. Every database change MUST be a
versioned up/down SQL migration in git, using `golang-migrate`, and MUST NOT be applied
by hand.

Under the operator's role and membership decision, the legacy customer role MUST map to
`viewer`: read-only, on projects of which the customer is a member. It MUST map to `editor`
only for write actions `../../old/docs/summary.md` explicitly shows, and MUST never map to
`admin`. A user with no workspace or project membership MUST receive HTTP 403, written to
the audit log. Plan MUST produce the endpoint × role permission matrix, with one test per
role per endpoint.

Under the operator's inventory decision, the endpoint, page, and schema inventory MUST be
built only from the permitted documents, and every item MUST cite its source. Anything
missing or truncated MUST be marked UNVERIFIED and left open; it MUST NOT be invented.
Nothing may enter the rebuild without a source.

Under the operator's recorded parity decision, the following rules MUST override the
advisory choices in `../../old/docs/approach.md` and `../../old/docs/roadmap.md`.
Legacy behavior visible to users MUST be preserved, including responses, messages, and
flows. Legacy defects that lose or corrupt data MUST be fixed. Each fix MUST be recorded
as a named, deliberate difference from legacy and MUST have a test.

1. **Checkout and partial failure — FIX.** Checkout MUST run in one database transaction:
   either everything is saved or nothing is. A failure MUST leave no half-written order.
   User-visible success and failure behavior MUST remain as in legacy.
2. **Time-based ID collisions — FIX.** IDs MUST come from a database sequence or UUID,
   never from the clock. Any user-visible ID format MUST be preserved; only the generation
   mechanism MUST change.
3. **Email failure and cleanup — FIX.** A failed email MUST NOT fail or roll back an order.
   The failure MUST be logged and the email MUST be retried in the background. Temporary
   files and records MUST be cleaned up whether the step succeeds or fails.
4. **Image fallback and MIME — KEEP, with upload content validation.** Missing images MUST
   use the same legacy placeholder image, and served content types MUST remain as in
   legacy. Uploads MUST check a file's real type from its content, not trust its name.
5. **Incomplete client assets.** Missing assets MUST NOT be invented. A neutral placeholder
   MUST be used, each missing asset MUST be listed as an open item for the client, and its
   parity MUST be marked UNVERIFIED.

## Development Workflow

1. Work MUST proceed through specify, plan, tasks, then implement. Each phase MUST serve
   the Java-to-Go goal and consume the applicable earlier artifacts and this constitution.
   Adoption of this constitution does not itself authorize application implementation.
2. Specifications MUST identify legacy evidence, intended behavior, operator-directed
   differences, acceptance cases, and unresolved questions. Missing information MUST be
   recorded in the file as a plain-text clarification marker; this unattended session
   MUST NOT stop to ask the operator questions.
3. Plans MUST document layer boundaries, data and external-side-effect contracts,
   authentication and authorization enforcement, test gates, and the status of applicable
   clarification markers. Tasks MUST put acceptance tests before the implementation they
   verify and include role, refusal, audit, error-envelope, and page coverage.
4. Changes MUST be small and reviewable, and each MUST include a changelog entry.
   The adoption entry for this documentation-only change is in the Sync Impact Report.
   Code changes MUST demonstrate required build, lint, type, unit, acceptance, and browser
   checks; a missing or unresolved gate MUST NOT be reported as passing.
5. Reviews MUST check the Java-to-Go objective, source traceability, layering, module size,
   dependency imports, migration discipline, configuration and secret handling, access
   enforcement, error envelopes, request-ID logging, tests, and the changelog.
6. Work MUST stay within this working tree; legacy documents are read-only evidence.
   Agents MUST NOT push, merge, or deploy. The referee and platform gates decide acceptance.

## Governance

This constitution governs specifications, plans, tasks, implementation, and reviews for
this Java-to-Go rebuild, subject to the operator's explicit instructions. A contradictory
source or advisory document MUST NOT silently redefine the declared goal. Where an actual
operator resolution is recorded, subsequent artifacts MUST cite that resolution; otherwise
the conflicting document and unresolved choice MUST remain visible in a plain-text marker.

An amendment MUST state its rationale, affected principles, impact on existing artifacts
and tests, and any required migration or remediation. It MUST update the Sync Impact
Report, include a changelog entry, and record the amendment date. Resolving a clarification
MUST cite the operator decision rather than inventing an exception. In an unattended
session, unresolved questions MUST remain recorded while independent authorized work
continues. No unresolved marker grants a waiver of a mandatory rule.

Constitution versions MUST follow semantic versioning: MAJOR for incompatible principle
removals or redefinitions, MINOR for new principles or materially expanded obligations,
and PATCH for clarifications or other non-semantic refinements. Initial adoption is
version 1.0.0. The original ratification date MUST remain unchanged in future amendments;
the last-amended date MUST change only when the constitution changes.

Every plan and review MUST evaluate compliance with this constitution and identify any
unresolved conflict affecting its scope. Mandatory requirements MUST NOT be waived by an
implementation convenience or an advisory migration roadmap. Passing local checks permits
review only; the referee and platform gates retain acceptance authority.

**Version**: 1.0.1 | **Ratified**: 2026-11-09 | **Last Amended**: 2026-11-10
