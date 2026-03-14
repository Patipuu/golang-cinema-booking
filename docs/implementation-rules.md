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

Comment for document
1) Architecture overview section
Change 1 – Explicitly list all internal packages as first‑class layers
Where: In the “Architecture” / “Layered architecture” section, where you list the core packages (currently likely mentioning handler, service, repository, domain, utils).
Edit: Expand the bullet list to include middleware, config, and database:
Add/ensure bullets like:
``internal/handler`` – HTTP endpoints (controllers) for the API.
``internal/service`` – business logic and use cases, depending on repositories and domain.
``internal/repository`` – data access interfaces for each aggregate.
``internal/domain`` – core domain models and simple domain helpers.
``internal/utils`` – shared helpers for JWT, OTP, JSON responses, validation, logging, etc.
internal/middleware – HTTP middleware (auth, logging) built on top of handlers/utils.
internal/config – application configuration (env/.env) and structured config types.
internal/database – database connection and migrations (e.g., PostgreSQL pool + schema.sql).
Why: These packages already exist and are used as distinct layers (middleware, config, database); the rules doc should treat them as first‑class parts of the architecture, not implicit add‑ons.
Change 2 – Clarify allowed dependencies between these layers
Where: In the “Dependencies between layers” / “Allowed imports” section.
Edit: Add a short dependency statement such as:
- Handlers depend on services, utils, middleware, and config (for wiring), but never on repositories or database directly.

- Services depend on repositories, domain, and utils, but never on HTTP, middleware, or database.

- Repositories depend on database and domain, but never on handlers, services, or middleware.

- Middleware depends on utils, config, and (optionally) services, but never on repositories or database.
Why: This aligns the rules with how your packages are structured today and makes the allowed dependency directions explicit.
2) Handler / service / repository responsibilities
Change 3 – Handlers must use shared JSON response helpers
Where: In the section describing handler responsibilities.
Edit: Add/adjust a bullet:
- Handlers must use the JSON response helpers from internal/utils/response.go (JSONSuccess, JSONError, JSONBadRequest, JSONUnauthorized, JSONNotFound, JSONInternal) instead of manually setting status codes or encoding JSON.
Why: utils/response.go already defines a standard Response envelope and helpers; the rules should encode that contract.
Change 4 – Handlers should use shared validation helpers
Where: Same handler section or a “Validation” subsection.
Edit: Add:
- Request body validation should be done via shared helpers from internal/utils/validator.go (NewValidator, ValidateStruct) rather than ad‑hoc validation logic in each handler.
Why: This matches utils/validator.go and keeps validation consistent across handlers.
Change 5 – Services and repositories interact only through domain models
Where: In service and repository sections.
Edit: Make explicit that:
- Service interfaces in internal/service operate on internal/domain types (e.g., User, Booking, Cinema, Payment) and must not depend on HTTP or DB packages.

- Repository interfaces in internal/repository accept context.Context and domain types and must not depend on HTTP or handler code.
Why: This mirrors the current interfaces and enforces the clean separation you already have.
3) Security / JWT / auth rules
Change 6 – Centralize JWT usage through internal/utils/jwt
Where: In the “Security”, “Auth”, or “JWT” section.
Edit: Add a rule like:
- JWT tokens must be generated and parsed via internal/utils/jwt.go (GenerateToken, ParseToken) and the Claims type. Handlers, services, and middleware must not use github.com/golang-jwt/jwt/v5 directly.
Why: The code already wraps JWT usage in utils/jwt.go, and auth_middleware.go depends on that; the rules should require all JWT usage to go through this wrapper.
Change 7 – Document the auth middleware pattern
Where: Either under “Security” or a new “Middleware – Auth” subsection.
Edit: Add a short description:
 Authentication is enforced via internal/middleware/AuthMiddleware(jwtSecret string), which:
 - Reads the Authorization: Bearer <token> header.
 - Uses utils.ParseToken to validate the token and extract utils.Claims.
 - Stores claims in context under a dedicated key (e.g., UserClaimsKey) for downstream handlers.
 - Uses utils.JSONUnauthorized for missing/invalid tokens.

> Handlers that require authentication should be wrapped with this middleware instead of re‑implementing JWT checks.
Why: This matches auth_middleware.go and encourages a single, consistent auth path.
4) Domain and pagination rules
Change 8 – Allow simple domain helpers (e.g., pagination)
Where: In the “Domain” / “Entities” section.
Edit: Add a note:
- The internal/domain package can include small, pure helper methods that encapsulate domain behavior, such as pagination helpers on Page (e.g., Page.Offset()) and result metadata via PageResult. These helpers must remain side‑effect free and not depend on external layers.
Why: Page.Offset() and PageResult already exist; the rules should explicitly allow this pattern so it doesn’t look like a violation of “domain is just structs”.
5) Middleware and logging rules
Change 9 – Add a “Middleware” subsection
Where: Under architecture or logging sections.
Edit: Add a subsection like:
- Middleware

- Lives in internal/middleware.
- Stays thin: focuses on cross‑cutting concerns (auth, logging) and delegates business decisions to services.
- Auth middleware uses internal/utils/jwt and internal/utils/response for token parsing and error responses.
- Logging middleware uses go.uber.org/zap via a logger created in internal/utils/logger.go and should not introduce business logic.
Why: This reflects auth_middleware.go, logger_middleware.go, and utils/logger.go, and makes the expectations for middleware explicit.
6) Config and database rules
Change 10 – Describe config loading via config.Load
Where: In a “Configuration” section.
Edit: Add:
- Configuration is loaded via internal/config.Load(), which reads environment variables (and optional .env) into a strongly typed Config struct. Other layers should depend on config.Config (or its nested structs like ServerConfig, DatabaseConfig, JWTConfig, OTPConfig, SMTPConfig) rather than reading environment variables directly.
Why: This matches config/config.go and enforces a single config entry point.
Change 11 – Describe DB access via database.Connect
Where: In “Database” or infrastructure rules.
Edit: Add:
- Database connections are created via internal/database.Connect(ctx, cfg.Database.DSN()), which returns a DB wrapper around a pgxpool.Pool. Repository implementations should depend on this wrapper (or the pool it exposes) instead of creating their own connections.
Why: This reflects database/postgres.go and keeps DB access centralized.

