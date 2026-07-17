# Kilo `/plan-task` Command Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an opt-in global `/plan-task` command that produces granular, durable task records and hands them to Kilo's normal coordinator when the user says `go`.

**Architecture:** The command will live in the global Kilo command directory and use one reusable Markdown task-record template. It will plan only; the existing execution command and global rules will recognize its task records, select the next ready task, determine safe parallel subagent dispatch, then enforce review, fixer, and documentation closure.

**Tech Stack:** Kilo Markdown commands, Kilo Markdown agent definitions, global `kilo.jsonc` command discovery.

## Global Constraints

- Do not replace Kilo's default coordinator or default agent routing.
- Keep ADK Go routing out of `/plan-task` and its template.
- `/plan-task` must produce independently auditable, granular tasks rather than umbrella tickets.
- The coordinator, not the user, determines whether implementation lanes can run in parallel.
- Executors cannot close or review tasks; reviewers are read-only; fixers address one named finding.
- Task records, not agent chat, are the durable execution authority.
- No cloud, credential, deployment, database, IAM, secret, or other external mutation is authorized.

---

### Task 1: Add the global granular task-record template

**Files:**
- Create: `/Users/brian/.config/kilo/templates/task-record.md`

**Interfaces:**
- Consumes: an objective and repository-specific authority gathered by `/plan-task`.
- Produces: a task Markdown document with fields required by `/plan-task` and `/execute-task`.

- [ ] **Step 1: Create the template directory and task-record template**

Create `/Users/brian/.config/kilo/templates/task-record.md` with these required sections, in this order:

```markdown
# Task <ID>: <short objective>

Status: PLANNED

## Objective

## Dependencies

## Owned scope

## Non-goals

## Acceptance criteria

## Verification

## Dispatch decision

## Implementation receipt

## Review receipt

## Fix receipt

## Documentation triage

## Controller report
```

Include prompts under each heading that require concrete paths, commands,
results, and rationale; do not include generic `TODO` placeholders.

- [ ] **Step 2: Verify the template is complete and has no placeholder markers**

Run:

```bash
rg -n 'TODO|TBD|\[.*\]' /Users/brian/.config/kilo/templates/task-record.md
```

Expected: no output.

- [ ] **Step 3: Record task evidence**

Record the exact template path and the successful placeholder scan in the
implementation task's controller report.

### Task 2: Add `/plan-task` as an opt-in global planning command

**Files:**
- Create: `/Users/brian/.config/kilo/commands/plan-task.md`
- Reads: `/Users/brian/.config/kilo/templates/task-record.md`

**Interfaces:**
- Consumes: `/plan-task <objective>` and the global task-record template.
- Produces: an ordered set of repository-local `docs/tasks/<ID>-<slug>.md` task records, followed by a request for the user to say `go`.

- [ ] **Step 1: Write command metadata and planning boundary**

Create the command with this frontmatter and opening contract:

```markdown
---
description: Create granular durable tasks for an objective; plan only until the user says go
agent: planner
---

# Plan Task

Do not edit implementation files, dispatch executors, or close tasks. Read
project authority, active tasks, repository status, and relevant source before
planning.
```

- [ ] **Step 2: Add exact granular-decomposition rules**

Require the planner to create one task per independently auditable change lane.
Each task must copy the template and fill objective, dependencies, owned scope,
non-goals, acceptance criteria, verification, review scope, and documentation
triage. Require ordered task IDs and a dependency condition. Specify that the
planner must split migrations, persistence, business behavior, API contracts,
tests, and operational docs when a reviewer could accept one while rejecting
another.

- [ ] **Step 3: Add the `go` handoff contract**

Require the command to print the ordered ready-task list, identify the first
unblocked task, and end with exactly this handoff:

```markdown
Planning is complete. Reply `go` to execute Task <ID> through Kilo's normal coordinator.
```

- [ ] **Step 4: Verify the command only plans**

Run:

```bash
rg -n 'dispatch executors|Do not edit implementation files|Reply `go`|docs/tasks' /Users/brian/.config/kilo/commands/plan-task.md
```

Expected: one or more matches for every required planning boundary and handoff.

### Task 3: Teach the coordinator and execution command how to consume planned tasks

**Files:**
- Modify: `/Users/brian/.config/kilo/AGENTS.md`
- Modify: `/Users/brian/.config/kilo/commands/execute-task.md`

**Interfaces:**
- Consumes: `PLANNED` or ready task records created by `/plan-task`.
- Produces: a coordinator-controlled sequence of implementation, review, optional fix, re-review, and closure.

- [ ] **Step 1: Add opt-in workflow rules to global instructions**

Append a `/plan-task` section to `AGENTS.md` that states: this command is
opt-in and does not replace defaults; after the user says `go`, the coordinator
selects the first ready task; only the coordinator updates status and receipts.

- [ ] **Step 2: Add coordinator parallel-dispatch rules**

Require the coordinator to inspect task ownership, dependency inputs, and
verification commands. It may dispatch subagents in parallel only when files
do not overlap, lanes have no producer-consumer dependency, and no shared
contract is unresolved. Otherwise it sequences them. It records the decision
and short rationale in the parent task, checks changed-file overlap on return,
and requests integrated independent review.

- [ ] **Step 3: Update `/execute-task` task-adoption rule**

Replace its “create one with objective” fallback for `/plan-task` work with:

```markdown
When a ready task exists under `docs/tasks/`, adopt the first ordered task whose
dependencies are satisfied. Preserve its owned scope, verification, review,
and documentation-triage contract. Create a new task only when no matching
planned task exists.
```

Keep the existing lane selection, handoff receipts, independent reviewer, and
`DONE`/`BLOCKED`/`NEEDS_CONTEXT` constraints.

- [ ] **Step 4: Verify roles and closure rules are present**

Run:

```bash
rg -n 'plan-task|first ready task|parallel|overlap|independent review|Only the controller' \
  /Users/brian/.config/kilo/AGENTS.md \
  /Users/brian/.config/kilo/commands/execute-task.md
```

Expected: evidence that the command remains opt-in, the coordinator owns
dispatch decisions, and review/closure responsibilities stay separated.

### Task 4: Extend the global authority audit and validate the installed workflow

**Files:**
- Modify: `/Users/brian/.config/kilo/commands/kilo-config-audit.md`
- Reads: `/Users/brian/.config/kilo/commands/plan-task.md`
- Reads: `/Users/brian/.config/kilo/templates/task-record.md`

**Interfaces:**
- Consumes: installed global Kilo command and template paths.
- Produces: `KILO_AUTHORITY_PASS` only when the new opt-in workflow surfaces exist and remain global.

- [ ] **Step 1: Add the command and template to the required-global-file list**

Insert these required entries into audit step 1:

```markdown
- `~/.config/kilo/commands/plan-task.md`
- `~/.config/kilo/templates/task-record.md`
```

- [ ] **Step 2: Protect the global command name from project-local overrides**

In the local-command audit rule, fail `plan-task.md` alongside `execute-task.md`
and `kilo-config-audit.md` when any is found under an active workspace's
`.kilo/commands/` directory.

- [ ] **Step 3: Run the static authority validation**

From any workspace, check that all required global files exist and are
non-empty, then ensure there are no local `.kilo/commands/plan-task.md`
overrides. Report `KILO_AUTHORITY_PASS` only when both checks succeed.

- [ ] **Step 4: Perform documentation triage and record closure**

Record that this changes Kilo operating procedure, so the durable global
`AGENTS.md` and command documentation are updated. Record the static audit
result and confirm no external mutation occurred.

## Plan self-review

- Spec coverage: Task 1 supplies durable records; Task 2 plans granular work
  and waits for `go`; Task 3 coordinates sequential or collision-free parallel
  implementation with review/fix closure; Task 4 protects and validates the
  global surface.
- Placeholder scan: completed; the plan contains no `TODO` or `TBD` markers.
- Interface consistency: every planned-task consumer uses the same
  `docs/tasks/` location and the same template sections.
