# Identity Module

The Identity module owns user registration, activation, authentication, refresh
token rotation, logout, JWT validation, and role-aware authorization support.

## Capabilities

- Register users with unique username/email constraints.
- Hash passwords with bcrypt.
- Issue activation tokens and activate accounts.
- Authenticate active users.
- Issue short-lived JWT access tokens.
- Issue and rotate opaque refresh tokens.
- Logout by deleting refresh tokens.
- Attach authenticated user ID and roles to request context.
- Enforce role-based access in middleware.
- Expose a JWKS endpoint for the current public signing key.

## Structure

- `api`: HTTP handlers, router, auth middleware, response mapping.
- `domain`: identity entities and validation rules.
- `service`: application workflows, transaction boundaries, JWT/token logic.
- `postgres`: SQL repositories and transaction implementation.
- `store`: persistence-level sentinel errors.
- `module.go`: wiring between repositories, service, and router.

## Implementation Details

- Multi-step workflows use `WithinTx` so registration, activation, and token
  rotation are atomic.
- The service depends on small repository interfaces and a `RepositorySet`
  aggregate for transaction callbacks.
- Repository methods accept `context.Context` and apply bounded DB timeouts.
- Refresh tokens are stored as SHA-256 hashes, not plaintext.
- JWT access tokens include the user ID as `sub` and roles as custom claims.
- Database migrations enforce important invariants: unique users, unique role
  codes, valid token scopes, and cascade cleanup.
- Demo routes are disabled when the server runs with `-env production`.

## Tradeoffs

- Activation tokens are returned in the registration response. This keeps the
  learning flow self-contained; a production system would send them through an
  email/event workflow.
- JWT signing keys are generated in memory at startup. This is simple for local
  development, but restarts invalidate access tokens and multiple instances
  would not share keys.
- The API uses one `IdentityService` interface. It is slightly broad, but keeps
  wiring simple while the module is still small.
- The Postgres repository references the service transaction callback type.
  This is acceptable here, though a larger system might move transaction
  contracts to a narrower boundary package.
- Repository integration tests use a dedicated DB env var so they do not race
  with acceptance tests that migrate their own database.

## Left Out Deliberately

- Email delivery for activation. (Event driven architecture will help with this later.)
- Password reset and account recovery.
- Persistent JWT key storage and rotation.
- Rate limiting and abuse protection.
- Session/device management.
- Admin role assignment workflows.
- Production observability beyond basic structured logging.

These are valuable production concerns, but they are intentionally deferred so
the project can move on to the Projects service without turning Identity into a
hardening sprint.
