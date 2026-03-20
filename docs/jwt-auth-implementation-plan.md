# JWT Auth Implementation Plan (Sign Up / Sign In + Email Verification)

## Scope and Confirmed Decisions

1. `phone` is required at backend sign up validation.
2. Include `POST /api/auth/resend-verification` in the first implementation slice.
3. Authentication stack: JWT for sign in, OTP email verification for account activation.

## Target API Contract

### 1) Register
- Method: `POST /api/auth/register`
- Request body:
```json
{
  "email": "user@example.com",
  "password": "Secret123!",
  "username": "john_doe",
  "full_name": "John Doe",
  "phone": "+84901234567"
}
```
- Behavior:
  - Create user as `is_verified = false`
  - Generate OTP and expiry
  - Save OTP in DB
  - Send verification email
- Response: `201 Created` with safe user payload (no password hash / OTP fields)

### 2) Verify OTP
- Method: `POST /api/auth/verify-otp`
- Request body:
```json
{
  "user_id": "<uuid>",
  "otp_code": "123456"
}
```
- Behavior:
  - Validate OTP and expiry
  - Set user verified
  - Clear OTP fields
- Response: `200 OK`

### 3) Resend Verification
- Method: `POST /api/auth/resend-verification`
- Request body:
```json
{
  "email": "user@example.com"
}
```
- Behavior:
  - Find unverified user
  - Regenerate OTP + expiry
  - Update DB + resend email
- Response: `200 OK`

### 4) Login
- Method: `POST /api/auth/login`
- Request body:
```json
{
  "email": "user@example.com",
  "password": "Secret123!"
}
```
- Behavior:
  - Validate credentials
  - Reject if `is_verified = false`
  - Issue JWT
- Response: `200 OK` with `token` + safe user payload

## Comprehensive Phases

## Phase 0 - Baseline and Safety Nets

### Goals
- Freeze API contract and status code policy.
- Ensure local env and DB schema are consistent.

### Tasks
1. Confirm env keys in `.env`:
   - `JWT_SECRET`, `JWT_EXPIRY_HOURS`
   - `OTP_EXPIRY_MINUTES`
   - `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`
2. Confirm `users` table has required columns and unique constraints:
   - `email UNIQUE`, `username UNIQUE`
   - `otp_code`, `otp_expiry`, `is_verified`
3. Define shared error responses for:
   - duplicate email/username (`409`)
   - invalid credentials (`401`)
   - unverified account (`403`)
   - invalid/expired OTP (`400`)

### Exit Criteria
- Team agrees on response semantics and endpoint payloads.

## Phase 1 - Repository Layer (Persistence)

### Goals
- Implement user persistence operations with explicit error mapping.

### Files
- `internal/repository/user_repository.go`
- Add postgres implementation file (e.g. `internal/repository/user_repository_postgres.go`)

### Tasks
1. Extend interface with `FindByUsername(ctx, username string)`.
2. Implement methods:
   - `FindByID`
   - `FindByEmail`
   - `FindByUsername`
   - `Create`
   - `UpdateOTP`
   - `SetVerified` (also clear OTP fields)
3. Map pg errors to domain-friendly conflicts for unique violations.

### Exit Criteria
- Repository can create/find/update users and OTP values reliably.

## Phase 2 - Email Service (OTP Delivery)

### Goals
- Send OTP verification emails through SMTP.

### Files
- `internal/service/email_service.go`
- Add implementation file (e.g. `internal/service/email_service_impl.go`)

### Tasks
1. Update interface to support OTP payload, e.g.:
   - `SendVerificationEmail(to, fullName, otpCode string, expiresInMinutes int) error`
2. Implement SMTP sender using config.
3. Add simple plain-text template (optional HTML follow-up).
4. Add timeouts and clear error messages.

### Exit Criteria
- OTP email is successfully delivered in dev/staging SMTP account.

## Phase 3 - Auth Service (Business Logic)

### Goals
- Implement all auth workflows in one cohesive service.

### Files
- `internal/service/auth_service.go`
- Add implementation file (e.g. `internal/service/auth_service_impl.go`)
- `internal/utils/otp.go`
- `internal/utils/jwt.go`

### Tasks
1. Expand auth interface to include:
   - Register with required phone
   - Login
   - VerifyOTP
   - ResendVerification
2. Register flow:
   - validate required fields (`email`, `password`, `username`, `full_name`, `phone`)
   - check duplicate email and username
   - hash password with bcrypt
   - create user as unverified
   - generate OTP + expiry using utils
   - persist OTP
   - send email
3. VerifyOTP flow:
   - get user by ID
   - compare OTP
   - check expiry (`IsOTPExpired`)
   - mark verified and clear OTP fields
4. Resend flow:
   - find by email
   - reject verified users
   - regenerate OTP + expiry
   - persist and send email
5. Login flow:
   - find by email
   - compare bcrypt hash
   - require verified user
   - issue JWT (`GenerateToken`)

### Exit Criteria
- Service methods return stable business errors and expected data.

## Phase 4 - HTTP Handlers and Validation

### Goals
- Expose clean auth endpoints and strict input validation.

### Files
- `internal/handler/auth_handler.go`
- `internal/utils/response.go`

### Tasks
1. Inject `AuthService` into `AuthHandler` via constructor.
2. Implement handlers:
   - `Register`
   - `VerifyOTP`
   - `ResendVerification`
   - `Login`
3. Validate request body and required fields.
4. Return safe output DTOs (never expose `password_hash`, `otp_code`, `otp_expiry`).
5. Use consistent status codes and JSON envelope.

### Exit Criteria
- All auth endpoints respond correctly for success and failure paths.

## Phase 5 - Routing and Wiring in Main

### Goals
- Wire dependencies in API bootstrap and register routes.

### Files
- `cmd/api/main.go`

### Tasks
1. Build dependency graph:
   - user repository
   - email service
   - auth service
   - auth handler
2. Register routes:
   - `POST /api/auth/register`
   - `POST /api/auth/verify-otp`
   - `POST /api/auth/resend-verification`
   - `POST /api/auth/login`
3. Keep JWT middleware on protected business routes only.

### Exit Criteria
- Server starts and auth routes are reachable end-to-end.

## Phase 6 - QA, Postman, and Regression Checks

### Goals
- Validate happy paths and negative scenarios.

### Test Setup
1. Start API server at `http://localhost:8080`.
2. Ensure DB has latest schema and SMTP credentials are valid.
3. Create Postman environment variables:
    - `base_url = http://localhost:8080`
    - `test_email = user1@example.com`
    - `test_password = Secret123!`
    - `test_username = john_doe_01`
    - `test_full_name = John Doe`
    - `test_phone = +84901234567`
    - `user_id =` (empty; fill from register response)
    - `otp_code =` (empty; fill from email)
    - `jwt_token =` (empty; fill from login response)

### Test Cases

1. Register success (unverified user created)
    - Objective: verify sign up creates account and sends OTP
    - Method/URL: `POST {{base_url}}/api/auth/register`
    - Request data:
    ```json
    {
       "email": "{{test_email}}",
       "password": "{{test_password}}",
       "username": "{{test_username}}",
       "full_name": "{{test_full_name}}",
       "phone": "{{test_phone}}"
    }
    ```
    - Expected result:
       - Status `201`
       - Response `success=true`
       - `data` contains `id`, `email`, `username`, `phone`, `is_verified=false`
       - Response does not include `password_hash`, `otp_code`, `otp_expiry`

2. Register duplicate email
    - Objective: ensure unique email is enforced
    - Method/URL: `POST {{base_url}}/api/auth/register`
    - Request data: same as case 1
    - Expected result:
       - Status `409`
       - Error indicates email already registered

3. Register duplicate username
    - Objective: ensure unique username is enforced
    - Method/URL: `POST {{base_url}}/api/auth/register`
    - Request data: same as case 1 but with a different email
    - Expected result:
       - Status `409`
       - Error indicates username already taken

4. Login before verification
    - Objective: ensure unverified users cannot sign in
    - Method/URL: `POST {{base_url}}/api/auth/login`
    - Request data:
    ```json
    {
       "email": "{{test_email}}",
       "password": "{{test_password}}"
    }
    ```
    - Expected result:
       - Status `403`
       - Error indicates account is not verified

5. Verify OTP success
    - Objective: verify user account activation
    - Method/URL: `POST {{base_url}}/api/auth/verify-otp`
    - Request data:
    ```json
    {
       "user_id": "{{user_id}}",
       "otp_code": "{{otp_code}}"
    }
    ```
    - Expected result:
       - Status `200`
       - Response contains success message

6. Verify OTP invalid code
    - Objective: reject wrong OTP
    - Method/URL: `POST {{base_url}}/api/auth/verify-otp`
    - Request data:
    ```json
    {
       "user_id": "{{user_id}}",
       "otp_code": "000000"
    }
    ```
    - Expected result:
       - Status `400`
       - Error indicates invalid OTP code

7. Verify OTP expired code
    - Objective: reject expired OTP
    - Method/URL: `POST {{base_url}}/api/auth/verify-otp`
    - Request data: valid user with deliberately expired OTP
    - Expected result:
       - Status `400`
       - Error indicates OTP expired

8. Resend verification success
    - Objective: generate and deliver a new OTP for unverified user
    - Method/URL: `POST {{base_url}}/api/auth/resend-verification`
    - Request data:
    ```json
    {
       "email": "{{test_email}}"
    }
    ```
    - Expected result:
       - Status `200`
       - Success message indicates code was resent

9. Resend verification for verified user
    - Objective: block resend when account already verified
    - Method/URL: `POST {{base_url}}/api/auth/resend-verification`
    - Request data:
    ```json
    {
       "email": "{{test_email}}"
    }
    ```
    - Expected result:
       - Status `409`
       - Error indicates account is already verified

10. Login after verification (JWT issued)
      - Objective: validate successful sign in and JWT issuance
      - Method/URL: `POST {{base_url}}/api/auth/login`
      - Request data:
      ```json
      {
         "email": "{{test_email}}",
         "password": "{{test_password}}"
      }
      ```
      - Expected result:
         - Status `200`
         - `data.token` exists and is non-empty
         - `data.user` contains safe fields only

11. Protected route smoke test with Bearer token
      - Objective: ensure JWT can access protected APIs
      - Method/URL: `POST {{base_url}}/api/bookings` (or any protected endpoint)
      - Request headers:
         - `Authorization: Bearer {{jwt_token}}`
      - Request data: minimal valid booking payload
      - Expected result:
         - Not `401` due to auth (business validation errors are acceptable if payload is incomplete)

### Regression Checks
1. Confirm auth responses never include:
    - `password_hash`
    - `otp_code`
    - `otp_expiry`
2. Confirm generic auth error for bad credentials (`invalid email or password`) to avoid account enumeration.
3. Confirm `phone` is required during registration.
4. Confirm JWT is not returned by register or verify endpoints.

### Postman Automation (optional but recommended)
1. In Register test script: store `user_id` from response data.
2. In Login test script: store `jwt_token` from response data.
3. Add a collection pre-request script to set `Authorization` header from `jwt_token` for protected requests.

### Exit Criteria
- All 11 core test cases pass.
- No sensitive fields appear in auth responses.
- JWT-protected endpoint accepts valid token and rejects missing/invalid token.

## Phase 7 - Hardening and Follow-ups

### Goals
- Reduce abuse risk and improve maintainability.

### Tasks
1. Add rate limit for register/login/resend/verify endpoints.
2. Add account lock or cooldown after repeated OTP failures.
3. Add structured auth audit logs (no secrets).
4. Optional: add refresh tokens and logout token invalidation strategy.
5. Optional: add background email queue for resilience.

### Exit Criteria
- Security and operability checks are in place for production rollout.

## Implementation Order (Recommended)

1. Phase 1 repository
2. Phase 2 email service
3. Phase 3 auth service
4. Phase 4 handlers
5. Phase 5 route wiring
6. Phase 6 testing
7. Phase 7 hardening

## Risks and Mitigations

1. SMTP outages can break registration flow.
Mitigation: keep unverified user, return clear error, support resend endpoint.

2. OTP brute-force attempts.
Mitigation: rate limits, attempt counters, temporary lockouts.

3. Sensitive data leaks through response structs.
Mitigation: use explicit response DTOs only.

4. Frontend/back-end contract drift.
Mitigation: versioned Postman collection and contract test checklist.

## Acceptance Checklist

- [ ] Phone is required at backend registration validation.
- [ ] Login is blocked for unverified users.
- [ ] Register creates unverified user and sends OTP email.
- [ ] Verify endpoint activates account and clears OTP data.
- [ ] Resend verification endpoint is implemented.
- [ ] JWT is issued only on successful verified login.
- [ ] Auth responses do not expose sensitive fields.
- [ ] Postman test suite for auth flows is complete.
