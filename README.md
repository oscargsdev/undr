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

Use separate databases for development and acceptance tests.

- `UNDR_DB_DSN`: development database DSN used by the server.
- `UNDR_TEST_DB_DSN`: acceptance-test database DSN used only by acceptance tests.

Example:

```bash
export UNDR_DB_DSN='postgres://undrdb:pa55word@localhost/undrdb?sslmode=disable'
export UNDR_TEST_DB_DSN='postgres://undrdb:pa55word@localhost/undrdb_test?sslmode=disable'
```

## Create Test Database

Create the dedicated test database once:

```bash
createdb -h localhost -U undrdb undrdb_test
```

If your app role does not have `CREATEDB`, run it with a superuser and set owner to `undrdb`:

```bash
sudo -u postgres createdb -O undrdb undrdb_test
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
