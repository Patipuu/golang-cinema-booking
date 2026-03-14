# Project Implementation Rules

This document defines implementation rules for the entire project to keep code consistent, testable, and maintainable.

## 1) Architecture and Layer Boundaries

Use a strict layered architecture:

- handler layer (`internal/handler`): HTTP concerns only.
- service layer (`internal/service`): business logic only.
- repository layer (`internal/repository`): data persistence only.
- domain layer (`internal/domain`): core entities and business models.
- utils layer (`internal/utils`): pure helpers (stateless, reusable).

Mandatory boundary rules:

- Handlers must not call repositories directly.
- Repositories must not contain business logic.
- Services must not write HTTP responses or depend on `net/http`.
- Middleware should not contain endpoint-specific business logic.
- Domain models should not import handler/service/repository packages.

## 2) Dependency Direction

Allowed direction:

- handler -> service -> repository
- service -> utils/domain
- handler -> utils (for request/response helpers)

Not allowed:

- repository -> service/handler
- service -> handler
- cross-calling between unrelated handlers

## 3) Handler Rules

Handlers are responsible for:

- decoding and validating request input.
- calling exactly one service flow (or a clearly orchestrated sequence).
- mapping service errors to HTTP status codes.
- returning unified JSON response structure.
- returning safe DTOs (never expose internal sensitive fields).

Handlers must not:

- hash passwords.
- generate JWT directly.
- generate OTP directly.
- perform direct SQL or database logic.

## 4) Service Rules

Services are responsible for:

- implementing business rules.
- orchestrating repository operations.
- calling external integrations (email, payment providers, etc.).
- returning typed/sentinel business errors.

Services must:

- accept `context.Context` as the first parameter.
- normalize business-critical fields when needed (example: email lowercase).
- keep side effects explicit and ordered.

Services must not:

- depend on HTTP request/response objects.
- return HTTP status codes.

## 5) Repository Rules

Repositories are responsible for:

- database CRUD and query logic.
- mapping driver/database errors into repository-level errors.
- transaction handling where needed.

Repositories must:

- accept `context.Context` as first parameter.
- return clear not-found semantics (for example `ErrNotFound`).
- keep query methods focused and predictable.

Repositories must not:

- send emails or external network calls unrelated to data persistence.
- enforce API-level validation.

## 6) Error Handling Contract

Error flow policy:

- Repository returns repository errors (`ErrNotFound`, conflict mappings, etc.).
- Service maps repository errors into business errors (`ErrEmailExists`, `ErrInvalidCredentials`, etc.).
- Handler maps business errors to HTTP status codes and response messages.

Rules:

- Prefer sentinel errors + `errors.Is` for matching.
- Wrap technical errors with context using `fmt.Errorf("...: %w", err)`.
- Do not leak internal SQL/stack details to API clients.
- Keep error messages stable for frontend compatibility.

## 7) API Response Standards

Use the existing response helper utilities consistently.

Rules:

- Success response format must be consistent across endpoints.
- Error response format must be consistent across endpoints.
- Input validation errors return `400`.
- Authentication failures return `401`.
- Authorization/verification failures return `403`.
- Resource not found returns `404`.
- Conflict (duplicates/state conflict) returns `409`.
- Unexpected server errors return `500`.

## 8) Validation Standards

Validation split:

- handler: request shape and basic field validation.
- service: business validation and state validation.

Examples:

- Handler validates required JSON fields, min length, email format.
- Service validates domain conditions (duplicate username, user verified state, OTP expiry).

## 9) Security Rules

Mandatory:

- Passwords must be hashed with bcrypt before persistence.
- JWT secret must come from environment/config, never hardcoded.
- Sensitive fields must never be included in API response payloads.
- Protected endpoints must use auth middleware.
- Logs must not include passwords, OTP codes, tokens, or secrets.

## 10) Logging and Observability

Rules:

- Log request summary in middleware (method, path, status, latency).
- Log errors with operation context (which layer and action failed).
- Avoid noisy duplicate logs for the same error path.
- Use structured or consistent log format where possible.

## 11) Configuration Rules

Rules:

- All environment-dependent values must be loaded through `internal/config`.
- No hardcoded ports, secrets, SMTP credentials, or DB credentials.
- Add new config keys to config struct and document them in README.

## 12) Naming and Code Style

Rules:

- Follow Go naming conventions and `gofmt` formatting.
- Keep functions short and single-purpose.
- Prefer explicit, intention-revealing names.
- Avoid package-level global mutable state.
- Keep comments for non-obvious decisions, not trivial statements.

## 13) Testing Rules

Minimum expectations:

- Service layer tests for business logic and edge cases.
- Repository tests for DB behavior (can include integration tests).
- Handler tests for request validation and status code mapping.

Required auth scenarios:

- register success
- duplicate email/username
- login invalid credentials
- login unverified account
- OTP invalid/expired
- resend verification for verified user

## 14) Migration and Schema Rules

Rules:

- Schema updates must be reflected in migration files.
- Update domain model and repository queries in the same change set.
- Never make silent breaking schema changes.

## 15) Git and PR Rules

Rules:

- One feature or fix per PR when possible.
- Include summary, scope, and test evidence in PR description.
- Avoid unrelated refactors in functional PRs.
- Keep commits meaningful and atomic.

## 16) Definition of Done (DoD)

A change is done only when:

- it respects layer boundaries in this document.
- success and error responses are consistent.
- no sensitive data is exposed.
- tests pass (or test limitations are explicitly documented).
- documentation is updated when contract/config changes.

## 17) Special Rule for Auth Implementation

To keep consistency for auth and future modules:

- Keep handler and service separate.
- Do not combine HTTP and business logic in one function/file.
- Add new business rules in service first, then expose through handler.

---

Ownership: all contributors

Review cadence: update this document whenever architecture, API standards, or cross-cutting conventions change.
