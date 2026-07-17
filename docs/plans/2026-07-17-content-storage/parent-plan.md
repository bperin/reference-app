# Content Storage Plan

Status: PLANNED

## Objective

Add a user-owned document content domain that reserves a GCS object through a signed upload URL, records MIME and object metadata in PostgreSQL, and idempotently marks uploads complete when an authenticated Eventarc Cloud Storage event reaches the API.

## Repository lookup

Qdrant was available. The checkout was indexed with:

```sh
python3 /Users/brian/.codex/skills/qdrant-repo-lookup/scripts/repo_lookup.py index --repo .
```

Result: 78 files, 138 chunks, and 138 points indexed in `repo_lookup_v1`.

Queries:

- `content mime type parent children database goose migration`
- `Google Cloud Storage GCS signed upload URL service`
- `Eventarc Cloud Storage callback API endpoint OpenAPI OAuth2Auth`
- `Status PLANNED IN_PROGRESS BLOCKED NEEDS_CONTEXT task`

Verified candidates included `README.md`, `Makefile`, `db/sqlc.yaml`, `db/migrations/000003_create_posts.sql`, `db/queries/posts.sql`, `cmd/api/main.go`, `internal/config/config.go`, `internal/posts/*`, and `internal/auth/openapi_contract_test.go`. No existing GCS or Eventarc implementation was found.

## Authority reviewed

- `/Users/brian/.config/kilo/AGENTS.md`: global evidence-first task, role, Go, OpenAPI, and execution rules.
- Repository root check: `/Users/brian/code/reference-app/AGENTS.md` does not exist; `git rev-parse --show-toplevel` confirmed `/Users/brian/code/reference-app` is the root.
- `README.md`: Go HTTP API, explicit wiring, PostgreSQL, Goose, sqlc, and source-generated OpenAPI conventions.
- `Makefile`: exact migration, sqlc, test, race, vet, format, and Swagger commands.
- `go.mod`: Go 1.26.5 and current dependency baseline; GCS SDK is not present.
- `db/sqlc.yaml`: one sqlc generation unit per domain using all Goose migrations as schema authority.
- `db/migrations/000003_create_posts.sql`: repository migration style.
- `db/queries/posts.sql`: named sqlc query style.
- `internal/posts/post.go`, `service.go`, `handler.go`, `routes.go`, and `repository/*`: feature package, consumer-side repository, DTO, route, and adapter conventions.
- `cmd/api/main.go`: visible composition root, OAuth2 annotations, and domain route registration.
- `internal/config/config.go` and `.env.example`: environment validation and redacted logging conventions.
- `internal/auth/openapi_contract_test.go`: generated Swagger and `OAuth2Auth[user]` contract checks.
- `internal/database/postgres.go`: embedded Goose migration execution.

Specialized classification: all Go implementation and review lanes must load `go-systems-programmer`. Tasks 05 and 06 also require `go-oauth2auth-openapi-docs`. No ADK work exists.

## Active task review

The existing `docs/plans/2026-07-17-content-storage/parent-plan.md` was adopted because it matches this objective. The only other `PLANNED` record found, `docs/superpowers/plans/2026-07-17-kilo-plan-task-command-implementation.md`, concerns Kilo command implementation and is unrelated. No matching `BLOCKED` or `NEEDS_CONTEXT` content task exists. `DONE` records were not adopted or included.

## Architecture models

### Conservative

The API signs PUT URLs, stores the declared MIME type, and trusts Eventarc object-finalized metadata to complete the row. It keeps the public URL nullable until completion. This is smallest, but MIME validation is limited to declared and object metadata values.

### Balanced (selected)

Create a pending content row before signing, bind object keys to the authenticated user and content UUID, constrain the signed PUT to the declared content type, then have an authenticated, idempotent Eventarc callback inspect authoritative GCS object attributes and update MIME, size, generation, absolute `gs://` URL, and public URL. Parent-child relations use a nullable self-reference with cycle/self-parent protection. This separates user and machine trust boundaries while fitting existing package conventions.

### Aggressive

Add asynchronous content inspection, quarantine, malware scanning, MIME sniffing from downloaded bytes, retries, and a dedicated worker. This offers stronger validation but introduces lifecycle and operations complexity outside the requested upload-completion flow.

## Planning decisions

- `project_id` is nullable metadata on content; bucket configuration remains required for signing and event filtering.
- A content row has at most one nullable `parent_id`; one parent naturally has multiple children through the self-reference.
- Upload state is explicit (`pending`, `uploaded`, `failed`) rather than inferred from URL nullability.
- `gcs_uri` is the absolute `gs://bucket/object` URL. `public_url` is nullable because bucket objects are not assumed public; the runtime may populate a configured public base URL only when explicitly supported.
- MIME type is declared at reservation and reconciled from GCS object metadata after finalization. Deep byte-sniffing is outside this plan.
- The user upload endpoint requires `OAuth2Auth[user]`. The Eventarc callback uses Google-issued OIDC authentication and must not use the user OAuth scope.
- Event handling is idempotent by bucket, object name, and generation; stale or duplicate events cannot regress completed metadata.
- No cloud, IAM, bucket, Eventarc trigger, database, migration, or deployment mutation occurs during planning.

## Ordered child tasks

1. [`tasks/01-content-schema-migration.md`](tasks/01-content-schema-migration.md)
2. [`tasks/02-content-persistence.md`](tasks/02-content-persistence.md)
3. [`tasks/03-gcs-signed-upload-adapter.md`](tasks/03-gcs-signed-upload-adapter.md)
4. [`tasks/04-content-domain-behavior.md`](tasks/04-content-domain-behavior.md)
5. [`tasks/05-user-upload-http-contract.md`](tasks/05-user-upload-http-contract.md)
6. [`tasks/06-eventarc-completion-callback.md`](tasks/06-eventarc-completion-callback.md)
7. [`tasks/07-content-flow-tests.md`](tasks/07-content-flow-tests.md)
8. [`tasks/08-content-operations-docs.md`](tasks/08-content-operations-docs.md)

## Dispatch decision

Task 01: `sequential`. It owns one migration file and has no independent implementation lane; the unrelated active root-instructions task owns different paths and is not part of this dispatch.

## Backlog entrypoint

Task 01 is the first unblocked task.
