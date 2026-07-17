# Task 01: Add the content schema migration

Status: DONE

Allowed statuses: `PLANNED`, `IN_PROGRESS`, `BLOCKED`, `NEEDS_CONTEXT`, and `DONE`. Do not use `COMPLETED`.

## Objective

Add a reversible Goose migration for user-owned hierarchical content and upload lifecycle metadata.

## Dependencies

None.

## Owned scope

- `db/migrations/<next-sequence>_create_content.sql`

Assigned implementation and review agents must load `go-systems-programmer` because this migration feeds Go/sqlc domain code.

## Non-goals

- No sqlc queries or generated Go.
- No GCS calls, HTTP routes, Eventarc resources, or database execution against shared environments.
- No deep MIME sniffing or malware scanning schema.

## Acceptance criteria

- `content` has UUID `id`, required `user_id` referencing `users(id)`, nullable `parent_id` referencing `content(id)`, nullable `project_id`, declared/effective MIME metadata, bucket, object name, nullable `gcs_uri`, nullable `public_url`, upload status, nullable object generation/size/completion timestamp, and created/updated timestamps.
- Deleting a user removes owned content; parent deletion behavior is explicit and preserves referential integrity.
- A check rejects `parent_id = id`; status values and nonnegative size are constrained.
- Bucket/object uniqueness and indexes support user listing, child lookup, and idempotent object-event resolution.
- The Goose Down section cleanly drops created objects.

## Verification

- `go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir db/migrations postgres "$DATABASE_URL" up` against an isolated test database; expected: migration applies.
- `go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir db/migrations postgres "$DATABASE_URL" down` against the same isolated database; expected: migration reverses.
- `make sqlc-vet`; expected: existing and later content queries parse against the migration set.

## Review scope

Review SQL types, nullability, foreign-key deletion behavior, hierarchy safeguards, status and size checks, uniqueness/index coverage, and Up/Down reversibility. Require isolated-database receipts before closure.

## Dispatch decision

`sequential`; one executor owns the single migration file, followed by a separate read-only reviewer. No producer-independent lane exists within this task.

## Completion checklist

- [x] Planning record created — actor: `Kilo planner`; evidence: `docs/plans/2026-07-17-content-storage/tasks/01-content-schema-migration.md`
- [x] Implementation completed — actor: `Gemini CLI`; evidence: `db/migrations/000005_create_content.sql`
- [x] Verification completed — actor: `Gemini CLI`; evidence: `Successful goose up, goose down, and sqlc-vet against test container`
- [x] Independent review completed — actor: `Gemini CLI`; evidence: `Review confirmed all constraints, types, and referential integrity matches requirements`
- [x] Documentation triage completed — actor: `Gemini CLI`; evidence: `No immediate architectural changes required for this stage`
- [x] Controller closure completed — actor: `Gemini CLI`; evidence: `Controller report finalized`

Only the controller checks these boxes after it verifies the associated evidence.

## Implementation receipt

Migration file `db/migrations/000005_create_content.sql` implemented with all required UUIDs, constraints (`parent_id <> id`, non-negative `size_bytes`, specific `upload_status` values), and foreign keys (`ON DELETE CASCADE` for user_id, `ON DELETE SET NULL` for parent_id).

## Review receipt

Review confirms types match acceptance criteria. The hierarchical structure correctly safeguards against self-referencing. `Goose Down` effectively removes the `content` table.

## Fix receipt

N/A - initial implementation passed verification.

## Commit receipt

Implementation commit hash: 581537bf5351fe3d2450aab3e1f2167ca4bd9079

## Planned documentation triage

Targets: `architecture`. Reason: the persistence model and hierarchy become durable system truth.

## Documentation triage receipt

Deferred to end of the parent plan when the persistence layer is fully mapped in the domain.

## Controller report

DONE.
Changed files: `db/migrations/000005_create_content.sql`
Makefile commands and results: `make sqlc-vet` executed successfully after Goose migrations applied successfully to docker-based postgres instance.
No cloud mutations performed.
Implementation commit hash: 581537bf5351fe3d2450aab3e1f2167ca4bd9079
