.PHONY: help run test test-unit test-integration test-race test-cover vet format format-check migration migrate-up migrate-status cloud-migrate sqlc sqlc-vet swagger generate smoke cloud-smoke tidy clean

help:
	@echo "Available commands:"
	@echo "  make run                - Run the API using the Cloud SQL URL secret"
	@echo "  make test               - Run all unit tests"
	@echo "  make test-unit          - Run short unit tests"
	@echo "  make test-integration   - Run integration tests (requires DATABASE_URL)"
	@echo "  make test-race          - Run tests with race detector"
	@echo "  make test-cover         - Run tests and show coverage"
	@echo "  make vet                - Run go vet"
	@echo "  make format             - Format code with go fmt and prettier"
	@echo "  make format-check       - Verify code formatting"
	@echo "  make migration name=... - Create a new Goose migration"
	@echo "  make migrate-up         - Apply Goose migrations using an exported DATABASE_URL"
	@echo "  make migrate-status     - Show Goose migration status using an exported DATABASE_URL"
	@echo "  make cloud-migrate      - Apply Goose migrations using the Cloud SQL URL secret"
	@echo "  make sqlc               - Generate database code using sqlc"
	@echo "  make sqlc-vet           - Validate SQL queries"
	@echo "  make swagger            - Generate Swagger/OpenAPI documentation from Go annotations"
	@echo "  make generate           - Run all code generation steps"
	@echo "  make smoke              - Run Cloud SQL-backed API smoke test"
	@echo "  make cloud-smoke        - Run API smoke test against the Cloud SQL URL secret"
	@echo "  make tidy               - Run go mod tidy"
	@echo "  make clean              - Clean up binaries and unused imports"

run:
	@DATABASE_URL="$$(gcloud secrets versions access latest --project=slap-agent-builder --secret=reference-app-database-url)"; export DATABASE_URL; go run cmd/api/main.go

test:
	go test ./...

test-unit:
	go test -short ./...

test-integration:
	go test -tags=integration ./...

test-race:
	go test -race ./...

test-cover:
	go test -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet:
	go vet ./...

format:
	go fmt ./...
	@if command -v prettier > /dev/null; then \
		prettier --write "**/*.{json,yaml,yml,md}"; \
	fi

format-check:
	go fmt -n ./...
	@if command -v prettier > /dev/null; then \
		prettier --check "**/*.{json,yaml,yml,md}"; \
	fi

migration:
	@if [ -z "$(name)" ]; then echo "Error: 'name' is required (e.g. make migration name=create_users)"; exit 1; fi
	goose -dir db/migrations create $(name) sql

migrate-up:
	@: "$${DATABASE_URL:?DATABASE_URL is required}"; go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir db/migrations postgres "$$DATABASE_URL" up

migrate-status:
	@: "$${DATABASE_URL:?DATABASE_URL is required}"; go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir db/migrations postgres "$$DATABASE_URL" status

cloud-migrate:
	@DATABASE_URL="$$(gcloud secrets versions access latest --project=slap-agent-builder --secret=reference-app-database-url)"; export DATABASE_URL; go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir db/migrations postgres "$$DATABASE_URL" up

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --generalInfo main.go --dir ./cmd/api,./internal/auth,./internal/users,./internal/posts,./internal/http/response,./internal/content --output ./docs --parseInternal

generate: sqlc-check swagger

smoke: cloud-smoke

cloud-smoke: cloud-migrate
	@DATABASE_URL="$$(gcloud secrets versions access latest --project=slap-agent-builder --secret=reference-app-database-url)"; export DATABASE_URL; ./scripts/smoke.sh

sqlc-check: sqlc-vet sqlc

sqlc:
	sqlc generate -f db/sqlc.yaml

sqlc-vet:
	sqlc vet -f db/sqlc.yaml

tidy:
	go mod tidy

clean:
	go clean
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports not found, skipping import cleanup"; \
	fi
