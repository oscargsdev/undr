# undr
Discover underground music.

## Motivation

The goal of this project is to teach myself how to design and build modern backend systems.
I'll start with a modular monolith and then evolve the architecture to microservices.

Current module I'm working on -> Identity service

## Problem

- Underground / independent music discovery is harmed by engagement-driven platforms
- Algorithms bury small projects
- Communities form despite platforms, not because of them

## What undr is

- Backend-first platform
- Explicit discovery (no ranking, no algorithmic feeds)
- Users choose:
    - Projects
    - Scenes
    - Subscriptions
- Notifications are event-driven

## What undr is NOT

- No algorithmic recommendations
- No feeds
- No social graph
- No streaming
- No frontend-heavy ambitions

## Local Databases

Use separate databases for development, acceptance tests, and repository
integration tests.

- `UNDR_DB_DSN`: development database DSN used by the server.
- `UNDR_TEST_DB_DSN`: acceptance-test database DSN used only by acceptance tests.
- `UNDR_REPOSITORY_TEST_DB_DSN`: repository integration-test database DSN. This
  database should have migrations applied before running repository integration
  tests.

Example:

```bash
export UNDR_DB_DSN='postgres://undrdb:pa55word@localhost/undrdb?sslmode=disable'
export UNDR_TEST_DB_DSN='postgres://undrdb:pa55word@localhost/undrdb_test?sslmode=disable'
export UNDR_REPOSITORY_TEST_DB_DSN='postgres://undrdb:pa55word@localhost/undrdb_repository_test?sslmode=disable'
```

## Create Test Database

Create the dedicated test databases once:

```bash
createdb -h localhost -U undrdb undrdb_test
createdb -h localhost -U undrdb undrdb_repository_test
```

If your app role does not have `CREATEDB`, run it with a superuser and set owner to `undrdb`:

```bash
sudo -u postgres createdb -O undrdb undrdb_test
sudo -u postgres createdb -O undrdb undrdb_repository_test
```

## Acceptance Test Migrations

The acceptance test harness now uses `UNDR_TEST_DB_DSN` only, and refuses to run if it matches `UNDR_DB_DSN`.

It runs migrations automatically for setup/cleanup with `migrate`:

- setup: `migrate -path=./migrations -database=$UNDR_TEST_DB_DSN down -all` then `up`
- cleanup: `migrate -path=./migrations -database=$UNDR_TEST_DB_DSN down -all`

The first migration enables `citext` (`CREATE EXTENSION IF NOT EXISTS citext`).
If your DB user cannot create extensions, run this once as a superuser in the test DB:

```bash
sudo -u postgres psql -d undrdb_test -c 'CREATE EXTENSION IF NOT EXISTS citext;'
```

Manual example:

```bash
migrate -path=./migrations -database="$UNDR_TEST_DB_DSN" up
```

## Repository Integration Test Migrations

Repository integration tests use `UNDR_REPOSITORY_TEST_DB_DSN` so they cannot
race with the acceptance test harness while it migrates `UNDR_TEST_DB_DSN`
down/up.

Apply migrations manually before running repository integration tests:

```bash
migrate -path=./migrations -database="$UNDR_REPOSITORY_TEST_DB_DSN" up
go test ./internal/identity/postgres -run Integration -count=1
```

## JWT Key Management

The identity service currently generates an in-memory RSA signing key when the
process starts. This is acceptable for the learning phase because it keeps local
setup simple, but it is not production-ready:

- restarting the service invalidates every existing access token
- multiple service instances would not share the same signing key
- there is no key rotation lifecycle
- the key ID is static instead of being generated from persisted key metadata

Before production, replace this with persisted signing keys, explicit key IDs,
rotation support, and JWKS output that includes active public keys until all
tokens signed by retired keys have expired.

## API Spec

The current OpenAPI spec for the identity API lives at docs/openapi.yaml

It documents the routes mounted under `/v1/identity`, including the demo routes
that are available outside production.
