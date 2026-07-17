# Kilo Developer Workflow

This guide establishes the standard development loop, safety conventions, and architectural boundaries that Kilo agents must follow when modifying this codebase.

## 1. Safety & Secrets

- **Secrets Redaction:** NEVER log raw passwords, access tokens, refresh tokens, JWT secrets, or connection strings.
- **Config Management:** Use only `internal/config/config.go` for environment variable loading. Do not call `os.Getenv` in services, handlers, or repositories.

## 2. Directory & Import Boundaries

- **No sqlc outside Repositories:** Only concrete repository files (under `internal/*/repository/postgres.go`) may import sqlc-generated code. Handlers, services, and domain entities must remain unaware of sqlc models and use pure domain structures.
- **Feature Packages:** Keep handlers, services, routes, and entities cohesive inside `internal/users`, `internal/posts`, and `internal/auth`.

## 3. Dependency Injection

- Do not use reflection-based containers or DI packages.
- All constructor parameters must be explicitly wired inside `cmd/api/main.go`.

## 4. Standard Development Loop

When asked to implement new features or fix bugs:

1. **Analyze:** Inspect the migrations under `db/migrations` and queries under `db/queries`.
2. **Draft Queries:** Add any new queries inside `db/queries/*.sql` and run `make generate`.
3. **Write Unit Tests first:** Place business logic tests in `*_test.go` and implement the logic to satisfy them.
4. **Compile & Verify:** Run `make test` and `make vet` before concluding.
