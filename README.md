# Reference App

Reference App is a Go HTTP API that demonstrates explicit dependency wiring,
PostgreSQL persistence, and OAuth2 password and refresh-token grants. It
exposes registration and token endpoints, bearer-protected user profiles, and
public-read/protected-write posts.

## Requirements

- Go 1.26.5 or later
- PostgreSQL for integration and migration commands
- Optional: Google Cloud CLI authenticated to the deployment project for the
  Cloud SQL-backed commands

## Quick start

Create a local environment file and provide a reachable PostgreSQL URL:

```sh
cp .env.example .env
export DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/reference_app?sslmode=disable'
make migrate-up
go run ./cmd/api
```

The API listens on `:8080` by default. See the generated OpenAPI document at
[`docs/swagger.yaml`](docs/swagger.yaml) for the complete HTTP contract.

## Commands

```sh
make test            # run all tests
make test-unit       # run short tests
make test-race       # run tests with the race detector
make vet             # run go vet
make format          # format Go and supported text files
make format-check    # verify formatting
make sqlc            # generate SQLC database code
make sqlc-vet        # validate SQL queries
make swagger         # regenerate OpenAPI documentation
make migrate-up      # apply migrations using DATABASE_URL
make migrate-status  # show migration status using DATABASE_URL
make cloud-migrate   # apply migrations using Secret Manager database URL
make cloud-smoke     # migrate and run the Cloud SQL-backed smoke test
```

## Runtime configuration

Copy [`.env.example`](.env.example) for local development. Do not commit
`.env` or private key material. In deployed environments, `DATABASE_URL` is
injected from Secret Manager secret `reference-app-database-url`.

The current Cloud SQL instance uses public connectivity with a configured
`0.0.0.0/0` authorized network. Narrow that rule before production exposure.

## Project layout

- `cmd/api`: application entrypoint and dependency wiring
- `internal/auth`: OAuth2 grants, refresh-token lifecycle, and bearer middleware
- `internal/users`: protected user-profile endpoints
- `internal/posts`: public reads and protected mutations
- `internal/database`: PostgreSQL connection setup
- `db/migrations`: Goose schema migrations
- `db/queries`: SQLC query definitions
- `docs`: generated OpenAPI contract and task records
- `scripts`: operational smoke tests

## API boundaries

Each domain registers its own routes. Authentication is applied at the domain
route group, and handlers use the authenticated user UUID from request context
instead of rechecking credentials. Auth requests use JSON;
`POST /auth/oauth/token` accepts `password` and `refresh_token` grants.
